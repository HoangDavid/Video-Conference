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
		IDOrder:    []string{},
		IDToTracks: make(map[string]*webrtc.TrackLocalStaticRTP),

		Slots:       make(map[int]*domain.Slot),
		OwnerToSlot: make(map[string]int),
		SlotToOwner: make(map[int]string),
	}

	direction := webrtc.RTPTransceiverInit{Direction: webrtc.RTPTransceiverDirectionSendonly}
	// Preallocate video transceivers
	for i := 0; i < poolSize; i++ {
		tx, err := conn.GetPC().AddTransceiverFromKind(webrtc.RTPCodecTypeVideo, direction)

		if err != nil {
			log.Error("unable to add new transceiver")
			return nil, err
		}

		new_slot := &domain.Slot{
			Tx: tx,
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

		err := s.SubscribeVideo(peer)
		if err != nil {
			s.Log.Error("unable to subscribe room")
			return err
		}
	}

	return nil
}

// subscribe to remote peer video track
func (s *SubConn) SubscribeVideo(peer domain.Peer) error {

	s.Mu.Lock()
	defer s.Mu.Unlock()

	v := s.Videos
	peerID := peer.GetMetaData().PeerID

	if _, ok := v.OwnerToSlot[peerID]; ok {
		return nil
	}

	local, err := webrtc.NewTrackLocalStaticRTP(
		peer.Pub().GetLocalAV().Video.Codec().RTPCodecCapability,
		"loop"+peerID,
		"pion",
	)

	if err != nil {
		s.Log.Error("unable to create local track")
		return err
	}

	v.IDOrder = append(v.IDOrder, peerID)
	v.IDToTracks[peerID] = local

	if len(v.OwnerToSlot) < len(v.Slots) {
		for i := range len(v.Slots) {

			// Found a slot
			if _, ok := v.SlotToOwner[i]; !ok {
				v.SlotToOwner[i] = peerID
				v.OwnerToSlot[peerID] = i

				slot := v.Slots[i]
				slot.Tx.Sender().ReplaceTrack(local)

				pumpCtx, pumpCancel := context.WithCancel(s.Ctx)
				slot.PumpCtx = pumpCtx
				slot.PumpCancel = pumpCancel
				go peer.Pub().PumpVideo(pumpCtx, local, slot.Tx)
				break

			}
		}
	}

	return nil
}

// Unsubcribe to remote peers track
func (s *SubConn) UnsubscribeVideo(peer domain.Peer) error {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	v := s.Videos
	peerID := peer.GetMetaData().PeerID

	slotID, ok := v.OwnerToSlot[peerID]
	if ok {
		err := v.Slots[slotID].Tx.Sender().ReplaceTrack(nil)
		if err != nil {
			s.Log.Error("unable to detach track")
			return err
		}

		delete(v.OwnerToSlot, peerID)
		delete(v.SlotToOwner, slotID)
		v.Slots[slotID].PumpCtx = nil
		v.Slots[slotID].PumpCancel = nil
	}

	delete(v.IDToTracks, peerID)
	for i, id := range v.IDOrder {
		if id == peerID {
			v.IDOrder = append(v.IDOrder[:i], v.IDOrder[i+1:]...)
			break
		}
	}

	return nil
}

func (s *SubConn) SubcribeAudio(peer domain.Peer) {

}
