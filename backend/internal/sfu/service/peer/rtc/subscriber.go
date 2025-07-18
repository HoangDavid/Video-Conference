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

func NewSubscriber(stream sfu.SFU_SignalServer, stuns []string, log *slog.Logger, poolSize int, debounceInterval time.Duration) (*Subscriber, error) {
	c, err := newPeerConnection(stream, stuns, log, debounceInterval)

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
			ActiveTracks: make(map[*webrtc.TrackLocalStaticRTP]*webrtc.RTPTransceiver),
		},
		conn:      c,
		log:       log,
		direction: direction,
	}, nil

}

func (s *Subscriber) HandleOffer(sdp string) error {
	if err := s.conn.handleOffer(sdp); err != nil {
		return err
	}

	return nil
}

func (s *Subscriber) HandleAnswer(sdp string) error {
	if err := s.conn.handleAnswer(sdp); err != nil {
		return err
	}

	return nil
}

// subscribe to remote peer track
func (s *Subscriber) Subcribe(pub domain.Publisher) error {

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
	go s.attachTrack(pub.LocalAudio, audioT)
	go s.attachTrack(pub.LocalVideo, videoT)

	return nil

}

// Unsubcribe to remote peers track
func (s *Subscriber) Unsubscribe(pub domain.Publisher) error {
	var audioT *webrtc.RTPTransceiver
	var videoT *webrtc.RTPTransceiver

	s.Mu.Lock()
	audioT = s.ActiveTracks[pub.LocalAudio]
	videoT = s.ActiveTracks[pub.LocalVideo]
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
	s.Mu.Lock()
	if audioT != nil {
		s.IdleAudios = append(s.IdleAudios, audioT)
		delete(s.ActiveTracks, pub.LocalAudio)
	}
	if videoT != nil {
		s.IdleVideos = append(s.IdleVideos, videoT)
		delete(s.ActiveTracks, pub.LocalVideo)
	}
	s.Mu.Unlock()

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
func (s *Subscriber) attachTrack(local *webrtc.TrackLocalStaticRTP, tx *webrtc.RTPTransceiver) error {

	if local == nil || tx == nil {
		return nil
	}

	for {
		if s.PC.SignalingState() == webrtc.SignalingStateStable && tx.Mid() != "" {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if err := tx.Sender().ReplaceTrack(local); err != nil {
		s.log.Error("unable to attach track")
		return err

	}

	s.Mu.Lock()
	s.ActiveTracks[local] = tx
	s.Mu.Unlock()

	return nil

}

func (s *Subscriber) WireCallBacks() {
	s.PC.OnICECandidate(s.conn.handleLocalIceCandidate)
	s.PC.OnTrack(nil)
	s.PC.OnNegotiationNeeded(s.conn.handleNegotiationNeeded)
}
