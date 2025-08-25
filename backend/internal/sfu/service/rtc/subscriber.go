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

	v := &domain.SubVideo{
		IDOrder:         []string{},
		IDToVideoTracks: make(map[string]*webrtc.TrackLocalStaticRTP),
		IDToAudioTracks: make(map[string]*webrtc.TrackLocalStaticRTP),

		Slots:       make(map[int]*domain.Slot),
		OwnerToSlot: make(map[string]int),
		SlotToOwner: make(map[int]string),
	}

	direction := webrtc.RTPTransceiverInit{Direction: webrtc.RTPTransceiverDirectionSendonly}
	// Preallocate video transceivers
	for i := 0; i < poolSize; i++ {
		vtx, err := conn.GetPC().AddTransceiverFromKind(webrtc.RTPCodecTypeVideo, direction)
		if err != nil {
			log.Error("unable to add new video transceiver")
			return nil, err
		}

		atx, err := conn.GetPC().AddTransceiverFromKind(webrtc.RTPCodecTypeAudio, direction)
		if err != nil {
			log.Error("unable to add new audio transceiver")
			return nil, err
		}

		new_slot := &domain.Slot{
			VideoTx: vtx,
			AudioTx: atx,
		}
		v.Slots[i] = new_slot
	}

	subCtx, subCancel := context.WithCancel(ctx)

	return &SubConn{
		SubConn: &domain.SubConn{
			Log:    log,
			Conn:   conn,
			Ctx:    subCtx,
			Cancel: subCancel,
			Videos: v,

			RecvSdp: make(chan *sfu.PeerSignal_Sdp, 64),
			RecvIce: make(chan *sfu.PeerSignal_Ice, 64),
		},
	}, nil
}

// wire pc call backs
func (s *SubConn) WireCallBacks() {
	pc := s.Conn.GetPC()
	pc.OnICECandidate(func(c *webrtc.ICECandidate) {
		s.Conn.HandleLocalIce(c, sfu.PcType_SUB)
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

// subcriber the whole room beside themselves
func (s *SubConn) SubscribeRoom(subscriberID string, room domain.Room) error {

	for id, peer := range room.ListPeers() {
		if id == subscriberID {
			continue
		}

		err := s.Subscribe(peer)
		if err != nil {
			s.Log.Error("unable to subscribe room")
			return err
		}
	}

	return nil
}

// subscribe to remote peer video track
func (s *SubConn) Subscribe(peer domain.Peer) error {

	s.Mu.Lock()
	defer s.Mu.Unlock()

	v := s.Videos
	peerID := peer.GetMetaData().PeerID

	if _, ok := v.OwnerToSlot[peerID]; ok {
		return nil
	}

	vlocal, err := webrtc.NewTrackLocalStaticRTP(
		peer.Pub().GetLocalAV().Video.Codec().RTPCodecCapability,
		"loop"+peerID,
		"pion",
	)

	if err != nil {
		s.Log.Error("unable to create local track")
		return err
	}

	alocal, err := webrtc.NewTrackLocalStaticRTP(
		peer.Pub().GetLocalAV().Audio.Codec().RTPCodecCapability,
		"loop"+peerID,
		"pion",
	)

	if err != nil {
		s.Log.Error("unable to create local track")
		return err
	}

	v.IDOrder = append(v.IDOrder, peerID)
	v.IDToVideoTracks[peerID] = vlocal
	v.IDToAudioTracks[peerID] = alocal

	if len(v.OwnerToSlot) < len(v.Slots) {
		for i := range len(v.Slots) {

			// Found a slot
			if _, ok := v.SlotToOwner[i]; !ok {
				v.SlotToOwner[i] = peerID
				v.OwnerToSlot[peerID] = i

				slot := v.Slots[i]
				slot.VideoTx.Sender().ReplaceTrack(vlocal)
				slot.AudioTx.Sender().ReplaceTrack(alocal)

				pumpCtx, pumpCancel := context.WithCancel(s.Ctx)
				slot.PumpCtx = pumpCtx
				slot.PumpCancel = pumpCancel
				go peer.Pub().PumpAudio(pumpCtx, alocal)
				go peer.Pub().PumpVideo(pumpCtx, vlocal, slot.VideoTx)
				break

			}
		}
	}

	return nil
}

// Unsubcribe to remote peers track
func (s *SubConn) Unsubscribe(peerID string) error {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	v := s.Videos

	slotID, ok := v.OwnerToSlot[peerID]
	if ok {
		slot := v.Slots[slotID]
		if err := slot.VideoTx.Sender().ReplaceTrack(nil); err != nil {
			s.Log.Error("unable to detach video track")
			return err
		}

		if err := slot.AudioTx.Sender().ReplaceTrack(nil); err != nil {
			s.Log.Error("unable to detach audio track")
			return err
		}

		slot.PumpCancel()

		delete(v.OwnerToSlot, peerID)
		delete(v.SlotToOwner, slotID)
		v.Slots[slotID].PumpCtx = nil
		v.Slots[slotID].PumpCancel = nil
	}

	delete(v.IDToVideoTracks, peerID)
	delete(v.IDToAudioTracks, peerID)

	for i, id := range v.IDOrder {
		if id == peerID {
			v.IDOrder = append(v.IDOrder[:i], v.IDOrder[i+1:]...)
			break
		}
	}

	return nil
}

func (s *SubConn) SwitchNext() error {
	return nil
}

func (s *SubConn) SwitchPrev() error {
	return nil
}
