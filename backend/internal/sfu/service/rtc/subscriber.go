package rtc

import (
	"context"
	"log/slog"
	"time"
	sfu "vidcall/api/proto"
	"vidcall/internal/sfu/domain"

	"github.com/pion/webrtc/v3"
)

type SubConn struct {
	*domain.SubConn
}

func NewSubscriber(ctx context.Context, sendQ chan *sfu.PeerSignal, log *slog.Logger, poolSize int, debounceInterval time.Duration) (domain.Subscriber, error) {

	conn, err := NewPConn(sendQ, log, debounceInterval, false)

	if err != nil {
		return nil, err
	}

	direction := webrtc.RTPTransceiverInit{Direction: webrtc.RTPTransceiverDirectionSendonly}

	// Add audio tracks
	audioT, err := conn.GetPC().AddTransceiverFromKind(webrtc.RTPCodecTypeAudio, direction)
	if err != nil {
		log.Error("unable to create audio track")
		return nil, err
	}

	//  Add video tracks
	var idleVideos []*webrtc.RTPTransceiver
	for i := 0; i < poolSize; i++ {
		videoT, err := conn.GetPC().AddTransceiverFromKind(webrtc.RTPCodecTypeVideo, direction)
		if err != nil {
			log.Error("unable to create video track")
		}

		idleVideos = append(idleVideos, videoT)

	}

	ordered := false
	maxRetransmits := uint16(0)

	//  add subtitle track
	sub, err := conn.GetPC().CreateDataChannel(
		"subtitles",
		&webrtc.DataChannelInit{
			Ordered:        &ordered,
			MaxRetransmits: &maxRetransmits,
		},
	)

	if err != nil {
		log.Error("unable to create subtitle track")
	}

	subCtx, subCancel := context.WithCancel(ctx)

	return &SubConn{
		SubConn: &domain.SubConn{
			Conn:   conn,
			Ctx:    subCtx,
			Cancel: subCancel,

			Direction:  direction,
			AudioOut:   audioT,
			IdleVideos: idleVideos,
			Sub:        sub,

			ActiveVideos: make(map[string]*webrtc.RTPTransceiver),

			RecvSdp: make(chan *sfu.PeerSignal_Sdp, 64),
			RecvIce: make(chan *sfu.PeerSignal_Ice, 64),
		},
	}, nil
}

// wire pc call backs
func (s *SubConn) WireCallBacks() {
	pc := s.Conn.GetPC()
	pc.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			return
		}
		s.Conn.HandleLocalIce(c)

	})
	pc.OnTrack(nil)
	pc.OnNegotiationNeeded(nil)
}

// start ice/sdp exchange for pc
func (s *SubConn) Connect() error {
	// Send an offer to client
	if err := s.Conn.SendOffer(sfu.PcType_SUB); err != nil {
		return err
	}

	for {
		select {
		case <-s.Ctx.Done():
			return nil

		case sdp, ok := <-s.RecvSdp:
			if !ok {
				return nil
			}

			if sdp.Sdp.Type == sfu.SdpType_ANSWER {
				if err := s.Conn.HandleAnswer(sdp); err != nil {
					return err
				}
			}

		case ice, ok := <-s.RecvIce:
			if !ok {
				return nil
			}

			if err := s.Conn.HandleRemoteIce(ice); err != nil {
				return err
			}
		}
	}
}

// tear down goroutines and pc
func (s *SubConn) Disconnect() error {
	if err := s.Conn.Close(); err != nil {
		return err
	}

	s.Cancel()

	close(s.RecvSdp)
	close(s.RecvIce)

	s.Log.Info("Subcriber pc disconnected")

	return nil
}

func (s *SubConn) EnqueueSdp(sdp *sfu.PeerSignal_Sdp) {
	select {
	case s.RecvSdp <- sdp:
	default:
	}
}

func (s *SubConn) EnqueueIce(ice *sfu.PeerSignal_Ice) {
	select {
	case s.RecvIce <- ice:
	default:
	}
}

// subscribe to remote peer video track
func (s *SubConn) SubscribeVideo(peer domain.Peer) error {

	var videoT *webrtc.RTPTransceiver

	s.Mu.Lock()
	videoT, s.IdleVideos = s.pop(s.IdleVideos)
	s.Mu.Unlock()

	if videoT == nil {
		// TODO: add paging for subcriber
	}

	// attach video track dynamically
	go s.attachTrack(peer.GetID(), peer.Pub().GetLocalVideo(), videoT)

	return nil

}

// Unsubcribe to remote peers track
func (s *SubConn) UnsubscribeVideo(peerID string) error {
	var videoT *webrtc.RTPTransceiver

	s.Mu.Lock()
	videoT = s.ActiveVideos[peerID]
	delete(s.ActiveVideos, peerID)
	s.Mu.Unlock()

	if videoT != nil {
		if err := videoT.Sender().ReplaceTrack(nil); err != nil {
			return err
		}
	}

	// update active and idle video track map
	go s.detachTrack(videoT)

	return nil
}

func (s *SubConn) SubcribeAudio(peer domain.Peer) {
	if peer == nil {
		return
	}

	go func() {
		if err := s.AudioOut.Sender().ReplaceTrack(peer.Pub().GetLocalAudio()); err != nil {
			s.Log.Error("unable to replace audio track")
			return
		}
	}()
}

// list pop helper function
func (s *SubConn) pop(tl []*webrtc.RTPTransceiver) (*webrtc.RTPTransceiver, []*webrtc.RTPTransceiver) {
	if len(tl) == 0 {
		return nil, tl
	}

	return tl[0], tl[1:]
}

// attach track
func (s *SubConn) attachTrack(peerID string, local *webrtc.TrackLocalStaticRTP, tx *webrtc.RTPTransceiver) error {

	if tx == nil {
		return nil
	}

	if err := tx.Sender().ReplaceTrack(local); err != nil {
		s.Log.Error("unable to attach track")
		return err

	}

	// Update active video track
	s.Mu.Lock()
	s.ActiveVideos[peerID] = tx
	s.Mu.Unlock()

	return nil
}

// dettach track
func (s *SubConn) detachTrack(tx *webrtc.RTPTransceiver) error {

	if tx == nil {
		return nil
	}

	if err := tx.Sender().ReplaceTrack(nil); err != nil {
		s.Log.Error("unable to detach track")
		return err

	}

	s.Mu.Lock()
	s.IdleVideos = append(s.IdleVideos, tx)
	s.Mu.Unlock()

	return nil
}
