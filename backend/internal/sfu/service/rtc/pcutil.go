package rtc

import (
	"log/slog"
	"sync"
	"time"
	sfu "vidcall/api/proto"

	"github.com/pion/webrtc/v3"
)

type conn struct {
	pc            *webrtc.PeerConnection
	sendQ         chan *sfu.PeerSignal
	iceBuffers    chan webrtc.ICECandidateInit
	log           *slog.Logger
	debounceTimer *debounceTimer
}

type debounceTimer struct {
	mu       sync.Mutex
	timer    *time.Timer
	interval time.Duration
}

// create new peer connection
func newPeerConnection(sendQ chan *sfu.PeerSignal, stuns []string, log *slog.Logger, debounceInterval time.Duration) (*conn, error) {
	pc, err := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{URLs: stuns},
		},
	})

	if err != nil {
		log.Error("unable to create new peer connection")
		return nil, err
	}

	return &conn{
		pc:         pc,
		sendQ:      sendQ,
		iceBuffers: make(chan webrtc.ICECandidateInit),
		log:        log,
		debounceTimer: &debounceTimer{
			interval: debounceInterval,
		},
	}, nil
}

// add ice from client
func (c *conn) handleRemoteIceCandidate(candidate *sfu.PeerSignal_Ice) error {
	mline := uint16(candidate.Ice.SdpMlineIndex)
	ice := webrtc.ICECandidateInit{
		Candidate:        candidate.Ice.Candidate,
		SDPMid:           &candidate.Ice.SdpMid,
		SDPMLineIndex:    &mline,
		UsernameFragment: &candidate.Ice.UsernameFragment,
	}

	if c.pc.RemoteDescription() == nil {
		c.iceBuffers <- ice
	} else {
		if err := c.pc.AddICECandidate(ice); err != nil {
			return err
		}
	}

	return nil
}

// send ice to client
func (c *conn) handleLocalIceCandidate(candidate *webrtc.ICECandidate) {
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

	c.sendQ <- req
}

// handle offer from client
func (c *conn) handleOffer(sdp *sfu.PeerSignal_Sdp) error {
	offer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  sdp.Sdp.Sdp,
	}

	// Set remote description
	if err := c.pc.SetRemoteDescription(offer); err != nil {
		c.log.Info("unable to set remote description")
		return err
	}

	// Flush already received ice candidates
	go c.flushIce()

	// Create answer and set local description
	answer, err := c.pc.CreateAnswer(nil)
	if err != nil {
		c.log.Error("unable to create answer")
		return err
	}

	if err := c.pc.SetLocalDescription(answer); err != nil {
		c.log.Error("unable to set local description")
		return err
	}

	// send answer to client
	res := &sfu.PeerSignal{
		Payload: &sfu.PeerSignal_Sdp{
			Sdp: &sfu.Sdp{
				Type: sfu.SdpType_ANSWER,
				Sdp:  c.pc.LocalDescription().SDP,
			},
		},
	}

	c.sendQ <- res

	return nil
}

// handle answer from client
func (c *conn) handleAnswer(sdp *sfu.PeerSignal_Sdp) error {
	answer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeAnswer,
		SDP:  sdp.Sdp.Sdp,
	}

	// Set Remote Description
	if err := c.pc.SetRemoteDescription(answer); err != nil {
		c.log.Info("unable to set remote description")
		return err
	}

	// Flush already received ice candidates
	go c.flushIce()

	return nil

}

func (c *conn) flushIce() {
	for {
		select {
		case ice := <-c.iceBuffers:
			if err := c.pc.AddICECandidate(ice); err != nil {
				c.log.Error("unable to add ice candidate")
			}
		default:
			return
		}
	}
}

// Rengegotiation with debounce
func (c *conn) handleNegotiationNeeded() {
	dt := c.debounceTimer

	dt.mu.Lock()
	defer dt.mu.Unlock()
	if dt.timer != nil {
		dt.timer.Stop()
	}

	dt.timer = time.AfterFunc(dt.interval, c.renegotiate)
}

func (c *conn) renegotiate() {
	dt := c.debounceTimer
	dt.mu.Lock()
	dt.timer = nil
	dt.mu.Unlock()

	if c.pc.SignalingState() != webrtc.SignalingStateStable {
		// TODO: retry:
		return
	}

	c.sendOffer()
}

func (c *conn) sendOffer() {
	offer, err := c.pc.CreateOffer(nil)
	if err != nil {
		c.log.Error("unable to create offer")
		return
	}
	if err = c.pc.SetLocalDescription(offer); err != nil {
		c.log.Error("unable to set local description")
		return
	}

	r := &sfu.PeerSignal{
		Payload: &sfu.PeerSignal_Sdp{
			Sdp: &sfu.Sdp{
				Type: sfu.SdpType_OFFER,
				Sdp:  c.pc.LocalDescription().SDP,
			},
		},
	}

	c.sendQ <- r
}
