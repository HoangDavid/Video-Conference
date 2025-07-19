package rtc

import (
	"log/slog"
	"time"
	sfu "vidcall/api/proto"
	"vidcall/internal/sfu/domain"

	"github.com/pion/webrtc/v3"
)

type Subscriber struct {
	*domain.Subscriber
	conn      *conn
	log       *slog.Logger
	direction webrtc.RTPTransceiverInit
}

func NewSubscriber(sendQ chan *sfu.PeerSignal, stuns []string, log *slog.Logger, poolSize int, debounceInterval time.Duration) (*Subscriber, error) {
	c, err := newPeerConnection(sendQ, stuns, log, debounceInterval)

	if err != nil {
		return nil, err
	}

	//  Add tracks to send
	direction := webrtc.RTPTransceiverInit{Direction: webrtc.RTPTransceiverDirectionSendonly}

	var idleAudios []*webrtc.RTPTransceiver
	var idleVideos []*webrtc.RTPTransceiver
	for i := 0; i < poolSize; i++ {

		audioT, err := c.pc.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio, direction)
		if err != nil {
			log.Error("unable to create audio track")
		}

		videoT, err := c.pc.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo, direction)
		if err != nil {
			log.Error("unable to create video track")
		}

		idleAudios = append(idleAudios, audioT)
		idleVideos = append(idleVideos, videoT)

	}

	// add dubbing audio track
	dub, err := c.pc.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio, direction)
	if err != nil {
		log.Error("unable to create dub audiotrack")
	}

	ordered := false
	maxRetransmits := uint16(0)

	//  add subtitle track
	sub, err := c.pc.CreateDataChannel(
		"subtitles",
		&webrtc.DataChannelInit{
			Ordered:        &ordered,
			MaxRetransmits: &maxRetransmits,
		},
	)

	if err != nil {
		log.Error("unable to create subtitle track")
	}

	return &Subscriber{
		Subscriber: &domain.Subscriber{
			PC:           c.pc,
			IdleAudios:   idleAudios,
			IdleVideos:   idleVideos,
			Dub:          dub,
			Sub:          sub,
			ActiveAudios: make(map[string]*webrtc.RTPTransceiver),
			ActiveVideos: make(map[string]*webrtc.RTPTransceiver),
		},
		conn:      c,
		log:       log,
		direction: direction,
	}, nil
}

func (s *Subscriber) SendOffer() {
	s.conn.sendOffer()
}

func (s *Subscriber) HandleRemoteIceCandidate(ice *sfu.PeerSignal_Ice) error {
	if err := s.conn.handleRemoteIceCandidate(ice); err != nil {
		return err
	}

	return nil
}

func (s *Subscriber) HandleAnswer(sdp *sfu.PeerSignal_Sdp) error {
	if err := s.conn.handleAnswer(sdp); err != nil {
		return err
	}

	return nil
}

func (s *Subscriber) SubscribeRoom(ownerID string, room *domain.Room) error {
	room.Mu.Lock()
	defer room.Mu.Unlock()

	for id, peer := range room.Peers {
		if ownerID == id {
			continue
		}

		if err := s.subscribe(peer); err != nil {
			return err
		}
	}

	return nil
}

func (s *Subscriber) SubscribePeer(peer *domain.Peer) error {
	if err := s.subscribe(peer); err != nil {
		return err
	}

	return nil
}

func (s *Subscriber) UnsubscribeRoom(ownerID string, room *domain.Room) error {
	room.Mu.Lock()
	defer room.Mu.Unlock()

	for id, peer := range room.Peers {
		if ownerID == id {
			continue
		}

		if err := s.unsubscribe(peer.ID); err != nil {
			return err
		}
	}

	return nil
}

func (s *Subscriber) UnsubscribePeer(peerID string) error {
	if err := s.unsubscribe(peerID); err != nil {
		return err
	}

	return nil
}

// subscribe to remote peer track
func (s *Subscriber) subscribe(peer *domain.Peer) error {

	var audioT *webrtc.RTPTransceiver
	var videoT *webrtc.RTPTransceiver
	var err error = nil

	s.Mu.Lock()
	audioT, s.IdleAudios = s.pop(s.IdleAudios)
	videoT, s.IdleVideos = s.pop(s.IdleVideos)
	s.Mu.Unlock()

	// Subcribe to audio track

	if audioT == nil {
		audioT, err = s.PC.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio, s.direction)

		if err != nil {
			s.log.Error("unable to add audio track")
			return err
		}
	}

	if videoT == nil {
		videoT, err = s.PC.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo, s.direction)

		if err != nil {
			s.log.Error("unable to add video track")
			return err
		}
	}

	// attach track dynamically
	go s.attachTrack(peer.ID, peer.Publisher.LocalAudio, audioT)
	go s.attachTrack(peer.ID, peer.Publisher.LocalVideo, videoT)

	return nil

}

// Unsubcribe to remote peers track
func (s *Subscriber) unsubscribe(peerID string) error {
	var audioT *webrtc.RTPTransceiver
	var videoT *webrtc.RTPTransceiver

	s.Mu.Lock()
	audioT = s.ActiveAudios[peerID]
	videoT = s.ActiveVideos[peerID]
	s.Mu.Unlock()

	if audioT != nil {
		if err := audioT.Sender().ReplaceTrack(nil); err != nil {
			s.log.Error("unable to unsubscribe to audio track")
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
func (s *Subscriber) pop(tl []*webrtc.RTPTransceiver) (*webrtc.RTPTransceiver, []*webrtc.RTPTransceiver) {
	if len(tl) == 0 {
		return nil, tl
	}

	return tl[0], tl[1:]
}

// attach track when pc connection is stable
func (s *Subscriber) attachTrack(peerID string, local *webrtc.TrackLocalStaticRTP, tx *webrtc.RTPTransceiver) error {

	if local == nil || tx == nil {
		return nil
	}

	for {
		if s.PC.SignalingState() == webrtc.SignalingStateStable {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if err := tx.Sender().ReplaceTrack(local); err != nil {
		s.log.Error("unable to attach track")
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

func (s *Subscriber) detachTrack(peerID string, tx *webrtc.RTPTransceiver) error {
	if tx == nil {
		return nil
	}

	for {
		if s.PC.SignalingState() == webrtc.SignalingStateStable {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if err := tx.Sender().ReplaceTrack(nil); err != nil {
		s.log.Error("unable to detach track")
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

func (s *Subscriber) WireCallBacks() {
	s.PC.OnICECandidate(s.conn.handleLocalIceCandidate)
	s.PC.OnTrack(nil)
	s.PC.OnNegotiationNeeded(s.conn.handleNegotiationNeeded)
}
