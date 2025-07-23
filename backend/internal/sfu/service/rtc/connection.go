package rtc

import (
	"log/slog"
	"time"
	sfu "vidcall/api/proto"
	"vidcall/internal/sfu/domain"
	"vidcall/internal/sfu/service/hub"

	"github.com/pion/webrtc/v3"
)

type PConn struct {
	*domain.PConn
}

// create new peer connection
func NewPConn(sendQ chan *sfu.PeerSignal, log *slog.Logger, debounceInterval time.Duration, withAudioLevel bool) (domain.Connection, error) {

	m := &webrtc.MediaEngine{}
	m.RegisterDefaultCodecs()
	var audioLevelURI string
	if withAudioLevel {
		audioLevelURI = "urn:ietf:params:rtp-hdrext:ssrc-audio-level"
		m.RegisterHeaderExtension(
			webrtc.RTPHeaderExtensionCapability{URI: audioLevelURI},
			webrtc.RTPCodecTypeAudio,
		)
	}

	pc, err := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{URLs: hub.Hub().GetStuns()},
		},
	})

	if err != nil {
		log.Error("unable to create new peer connection")
		return nil, err
	}

	return &PConn{
		PConn: &domain.PConn{
			PC:            pc,
			AudioLevelURI: audioLevelURI,
			Log:           log,
			IceBuffers:    make(chan webrtc.ICECandidateInit),
			SendQ:         sendQ,
			DebounceTimer: &domain.DebounceTimer{
				Interval: debounceInterval,
			},
		},
	}, nil
}

func (c *PConn) GetPC() *webrtc.PeerConnection {
	return c.PC
}

func (c *PConn) GetAudioURI() string {
	return c.AudioLevelURI
}

// add ice from client
func (c *PConn) HandleRemoteIce(candidate *sfu.PeerSignal_Ice) error {
	mline := uint16(candidate.Ice.SdpMlineIndex)
	ice := webrtc.ICECandidateInit{
		Candidate:        candidate.Ice.Candidate,
		SDPMid:           &candidate.Ice.SdpMid,
		SDPMLineIndex:    &mline,
		UsernameFragment: &candidate.Ice.UsernameFragment,
	}

	if c.PC.RemoteDescription() == nil {
		c.IceBuffers <- ice
	} else {
		if err := c.PC.AddICECandidate(ice); err != nil {
			return err
		}
	}

	return nil
}

// Send offer to client
func (c *PConn) SendOffer() error {
	offer, err := c.PC.CreateOffer(nil)
	if err != nil {
		c.Log.Error("unable to create offer")
		return err
	}
	if err = c.PC.SetLocalDescription(offer); err != nil {
		c.Log.Error("unable to set local description")
		return err
	}

	r := &sfu.PeerSignal{
		Payload: &sfu.PeerSignal_Sdp{
			Sdp: &sfu.Sdp{
				Type: sfu.SdpType_OFFER,
				Sdp:  c.PC.LocalDescription().SDP,
			},
		},
	}

	c.SendQ <- r

	return nil
}

// handle offer from client
func (c *PConn) HandleOffer(sdp *sfu.PeerSignal_Sdp) error {
	offer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  sdp.Sdp.Sdp,
	}

	// Set remote description
	if err := c.PC.SetRemoteDescription(offer); err != nil {
		c.Log.Info("unable to set remote description")
		return err
	}

	// Flush already received ice candidates
	go c.flushIce()

	// Create answer and set local description
	answer, err := c.PC.CreateAnswer(nil)
	if err != nil {
		c.Log.Error("unable to create answer")
		return err
	}

	if err := c.PC.SetLocalDescription(answer); err != nil {
		c.Log.Error("unable to set local description")
		return err
	}

	// send answer to client
	res := &sfu.PeerSignal{
		Payload: &sfu.PeerSignal_Sdp{
			Sdp: &sfu.Sdp{
				Type: sfu.SdpType_ANSWER,
				Sdp:  c.PC.LocalDescription().SDP,
			},
		},
	}

	c.SendQ <- res

	return nil
}

// handle answer from client
func (c *PConn) HandleAnswer(sdp *sfu.PeerSignal_Sdp) error {
	answer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeAnswer,
		SDP:  sdp.Sdp.Sdp,
	}

	// Set Remote Description
	if err := c.PC.SetRemoteDescription(answer); err != nil {
		c.Log.Error("unable to set remote description")
		return err
	}

	// Flush already received ice candidates
	go c.flushIce()

	return nil

}

// send ice to client
func (c *PConn) HandleLocalIce(candidate *webrtc.ICECandidate) {
	if c == nil {
		return
	}

	can := candidate.ToJSON()
	ice := &sfu.IceCandidate{
		Candidate:     can.Candidate,
		SdpMid:        *can.SDPMid,
		SdpMlineIndex: uint32(*can.SDPMLineIndex),
	}

	req := &sfu.PeerSignal{
		Payload: &sfu.PeerSignal_Ice{
			Ice: ice,
		},
	}

	c.SendQ <- req
}

func (c *PConn) flushIce() {
	for {
		select {
		case ice, ok := <-c.IceBuffers:

			if !ok {
				return
			}

			if err := c.PC.AddICECandidate(ice); err != nil {
				c.Log.Error("unable to add ice candidate")
			}
		default:
			return
		}
	}
}

// Rengegotiation with debounce (Future: if add tracks/ remove tracks)
func (c *PConn) HandleNegotiationNeeded() {
	dt := c.DebounceTimer

	dt.Mu.Lock()
	defer dt.Mu.Unlock()
	if dt.Timer != nil {
		dt.Timer.Stop()
	}

	dt.Timer = time.AfterFunc(dt.Interval, c.renegotiate)
}

func (c *PConn) renegotiate() {
	dt := c.DebounceTimer
	dt.Mu.Lock()
	dt.Timer = nil
	dt.Mu.Unlock()

	if c.PC.SignalingState() != webrtc.SignalingStateStable {
		// TODO: retry:
		return
	}

	c.SendOffer()
}

func (c *PConn) Close() error {
	c.DebounceTimer.Mu.Lock()
	if c.DebounceTimer.Timer != nil {
		c.DebounceTimer.Timer.Stop()
		c.DebounceTimer.Timer = nil
	}

	c.DebounceTimer.Mu.Unlock()

	close(c.IceBuffers)

	if err := c.PC.Close(); err != nil {
		c.Log.Error("Can't close peer connection")
		return err
	}

	return nil
}
