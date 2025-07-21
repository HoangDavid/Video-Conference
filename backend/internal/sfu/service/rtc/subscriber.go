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

	conn, err := NewPConn(sendQ, log, debounceInterval)

	if err != nil {
		return nil, err
	}

	//  Add tracks to send
	direction := webrtc.RTPTransceiverInit{Direction: webrtc.RTPTransceiverDirectionSendonly}

	var idleAudios []*webrtc.RTPTransceiver
	var idleVideos []*webrtc.RTPTransceiver
	for i := 0; i < poolSize; i++ {

		audioT, err := conn.GetPC().AddTransceiverFromKind(webrtc.RTPCodecTypeAudio, direction)
		if err != nil {
			log.Error("unable to create audio track")
		}

		videoT, err := conn.GetPC().AddTransceiverFromKind(webrtc.RTPCodecTypeVideo, direction)
		if err != nil {
			log.Error("unable to create video track")
		}

		idleAudios = append(idleAudios, audioT)
		idleVideos = append(idleVideos, videoT)

	}

	// add dubbing audio track
	dub, err := conn.GetPC().AddTransceiverFromKind(webrtc.RTPCodecTypeAudio, direction)
	if err != nil {
		log.Error("unable to create dub audiotrack")
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
			IdleAudios: idleAudios,
			IdleVideos: idleVideos,
			Dub:        dub,
			Sub:        sub,

			ActiveAudios: make(map[string]*webrtc.RTPTransceiver),
			ActiveVideos: make(map[string]*webrtc.RTPTransceiver),

			RecvSdp: make(chan *sfu.PeerSignal_Sdp),
			RecvIce: make(chan *sfu.PeerSignal_Ice),
		},
	}, nil
}

func (s *SubConn) WireCallBacks() {
	pc := s.Conn.GetPC()
	pc.OnICECandidate(s.Conn.HandleLocalIce)
	pc.OnTrack(nil)
	pc.OnNegotiationNeeded(s.Conn.HandleNegotiationNeeded)
}

func (s *SubConn) Connect() error {
	// Send an offer to client
	if err := s.Conn.SendOffer(); err != nil {
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
	s.RecvSdp <- sdp
}

func (s *SubConn) EnqueueIce(ice *sfu.PeerSignal_Ice) {
	s.RecvIce <- ice
}

func (s *SubConn) SubscribeRoom(ownerID string, room domain.Room) error {
	peers := room.ListPeers()

	for id, peer := range peers {
		if ownerID == id {
			continue
		}

		if err := s.Subscribe(peer); err != nil {
			return err
		}
	}

	return nil
}

func (s *SubConn) UnsubscribeRoom(ownerID string, room domain.Room) error {
	peers := room.ListPeers()

	for id, peer := range peers {
		if ownerID == id {
			continue
		}

		if err := s.Unsubscribe(peer.GetID()); err != nil {
			return err
		}
	}

	return nil
}

// subscribe to remote peer track
func (s *SubConn) Subscribe(peer domain.Peer) error {

	var audioT *webrtc.RTPTransceiver
	var videoT *webrtc.RTPTransceiver
	var err error = nil

	s.Mu.Lock()
	audioT, s.IdleAudios = s.pop(s.IdleAudios)
	videoT, s.IdleVideos = s.pop(s.IdleVideos)
	s.Mu.Unlock()

	// Subcribe to audio track

	if audioT == nil {
		audioT, err = s.Conn.GetPC().AddTransceiverFromKind(webrtc.RTPCodecTypeAudio, s.Direction)

		if err != nil {
			s.Log.Error("unable to add audio track")
			return err
		}
	}

	if videoT == nil {
		videoT, err = s.Conn.GetPC().AddTransceiverFromKind(webrtc.RTPCodecTypeVideo, s.Direction)

		if err != nil {
			s.Log.Error("unable to add video track")
			return err
		}
	}

	// attach track dynamically
	go s.attachTrack(peer.GetID(), peer.Pub().GetLocalAudio(), audioT)
	go s.attachTrack(peer.GetID(), peer.Pub().GetLocalVideo(), videoT)

	return nil

}

// Unsubcribe to remote peers track
func (s *SubConn) Unsubscribe(peerID string) error {
	var audioT *webrtc.RTPTransceiver
	var videoT *webrtc.RTPTransceiver

	s.Mu.Lock()
	audioT = s.ActiveAudios[peerID]
	videoT = s.ActiveVideos[peerID]
	s.Mu.Unlock()

	if audioT != nil {
		if err := audioT.Sender().ReplaceTrack(nil); err != nil {
			s.Log.Error("unable to unsubscribe to audio track")
			return err
		}

	}

	if videoT != nil {
		if err := videoT.Sender().ReplaceTrack(nil); err != nil {
			return err
		}
	}

	// update active and idle map
	go s.detachTrack(peerID, audioT)
	go s.detachTrack(peerID, videoT)

	return nil
}

// list pop helper function
func (s *SubConn) pop(tl []*webrtc.RTPTransceiver) (*webrtc.RTPTransceiver, []*webrtc.RTPTransceiver) {
	if len(tl) == 0 {
		return nil, tl
	}

	return tl[0], tl[1:]
}

// attach track when pc connection is stable
func (s *SubConn) attachTrack(peerID string, local *webrtc.TrackLocalStaticRTP, tx *webrtc.RTPTransceiver) error {

	if local == nil || tx == nil {
		return nil
	}

	for {
		select {
		case <-s.Ctx.Done():
			return nil
		default:
		}

		if s.Conn.GetPC().SignalingState() == webrtc.SignalingStateStable {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if err := tx.Sender().ReplaceTrack(local); err != nil {
		s.Log.Error("unable to attach track")
		return err

	}

	s.Mu.Lock()
	switch tx.Kind() {
	case webrtc.RTPCodecTypeAudio:
		s.ActiveAudios[peerID] = tx
	case webrtc.RTPCodecTypeVideo:
		s.ActiveVideos[peerID] = tx
	}
	s.Mu.Unlock()

	return nil
}

func (s *SubConn) detachTrack(peerID string, tx *webrtc.RTPTransceiver) error {
	if tx == nil {
		return nil
	}

	for {
		select {
		case <-s.Ctx.Done():
			return nil
		default:
		}

		if s.Conn.GetPC().SignalingState() == webrtc.SignalingStateStable {
			break
		}

		time.Sleep(10 * time.Millisecond)
	}

	if err := tx.Sender().ReplaceTrack(nil); err != nil {
		s.Log.Error("unable to detach track")
		return err

	}

	s.Mu.Lock()
	switch tx.Kind() {
	case webrtc.RTPCodecTypeAudio:
		s.IdleAudios = append(s.IdleAudios, tx)
		delete(s.ActiveAudios, peerID)
	case webrtc.RTPCodecTypeVideo:
		s.IdleVideos = append(s.IdleVideos, tx)
		delete(s.ActiveVideos, peerID)
	}
	s.Mu.Unlock()

	return nil
}
