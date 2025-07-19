package service

import (
	"context"
	"time"

	sfu "vidcall/api/proto"
	"vidcall/internal/sfu/domain"
	"vidcall/internal/sfu/service/hub"
	"vidcall/internal/sfu/service/rtc"
	"vidcall/pkg/logger"
)

type Peer struct {
	*domain.Peer
	publisher  *rtc.Publisher
	subscriber *rtc.Subscriber
}

func NewPeer(ctx context.Context, stream sfu.SFU_SignalServer, poolSize int) (*Peer, error) {
	log := logger.GetLog(ctx).With("layer", "service")
	stuns := hub.Hub().Stuns

	// Create channel to send msg and events
	sendQ := make(chan *sfu.PeerSignal)
	eventQ := make(chan *sfu.PeerSignal)

	duration := time.Duration(50 * time.Millisecond)
	pub, err := rtc.NewPublisher(sendQ, stuns, log, duration)
	if err != nil {
		return nil, err
	}

	sub, err := rtc.NewSubscriber(sendQ, stuns, log, poolSize, duration)
	if err != nil {
		return nil, err
	}

	// wire call backs
	pub.WireCallBacks()
	sub.WireCallBacks()

	return &Peer{
		Peer: &domain.Peer{
			Log:        log,
			Stream:     stream,
			Publisher:  pub.Publisher,
			Subscriber: sub.Subscriber,
			SendQ:      sendQ,
			EventQ:     eventQ,
		},
		publisher:  pub,
		subscriber: sub,
	}, nil
}

func (p *Peer) Connect() error {

	// send pc offer for subscriber
	p.subscriber.SendOffer()

	// start peer stream send loop
	go p.sendLoop()

	for {
		msg, err := p.Stream.Recv()
		if err != nil {
			p.Log.Error("unable to recieve msg from stream")
		}

		switch pl := msg.Payload.(type) {
		case *sfu.PeerSignal_Sdp:

			// exchange sdp offer/answer
			if pl.Sdp.Pc == sfu.PcType_PUB && pl.Sdp.Type == sfu.SdpType_OFFER {

				if err := p.publisher.HandleOffer(pl); err != nil {
					return err
				}
			}

			if pl.Sdp.Pc == sfu.PcType_SUB && pl.Sdp.Type == sfu.SdpType_ANSWER {

				if err := p.subscriber.HandleAnswer(pl); err != nil {
					return err
				}
			}

		case *sfu.PeerSignal_Ice:

			// add Ice to peer connections
			if pl.Ice.Pc == sfu.PcType_PUB {

				if err := p.publisher.HandleRemoteIceCandidate(pl); err != nil {
					return err
				}
			}

			if pl.Ice.Pc == sfu.PcType_SUB {

				if err := p.subscriber.HandleRemoteIceCandidate(pl); err != nil {
					return err
				}
			}
		case *sfu.PeerSignal_Action:
			p.RoomID = msg.RoomId
			p.ID = msg.PeerId
			p.handleAction(pl)

		case *sfu.PeerSignal_Event:
			//TODO: consumer from redis with a second SFU node

		}
	}
}

func (p *Peer) handleAction(action *sfu.PeerSignal_Action) {
	actionType := action.Action.Type

	switch actionType {
	case sfu.ActionType_START_ROOM:
		p.startRoom()
	case sfu.ActionType_END_ROOM:
		p.endRoom()
	case sfu.ActionType_JOIN:
		p.joinRoom()
	case sfu.ActionType_LEAVE:
		p.leaveRoom()

	}
}

// TODO: create redis instance for other SFU instance
func (p *Peer) startRoom() {
	hub.NewRoom(p.ID)
	hub.AddPeer(p.RoomID, p.Peer)

	// room live
	activeRoomE := p.createEventSignal(sfu.EventType_ROOM_INACTIVE)
	hub.BroadCast(p.ID, p.RoomID, activeRoomE)
}

func (p *Peer) joinRoom() {
	hub.AddPeer(p.RoomID, p.Peer)

	sub := p.subscriber
	room := hub.GetRoom(p.RoomID)

	// room is not live
	if room == nil {
		inactiveRoomE := p.createEventSignal(sfu.EventType_ROOM_INACTIVE)
		hub.BroadCast(p.ID, p.RoomID, inactiveRoomE)
		return
	}

	if err := sub.SubscribeRoom(p.ID, room); err != nil {
		return
	}

	// join event
	joinRoomE := p.createEventSignal(sfu.EventType_JOIN_EVENT)
	hub.BroadCast(p.ID, p.RoomID, joinRoomE)

}

func (p *Peer) endRoom() {
	// TODO: need the room inst
}

func (p *Peer) leaveRoom() {
	hub.RemovePeer(p.RoomID, p.Peer)
	sub := p.subscriber
	room := hub.GetRoom(p.RoomID)
	if err := sub.UnsubscribeRoom(p.ID, room); err != nil {
		return
	}

	// TODO: peer connection tear down

	// leave event
	leaveRoomE := p.createEventSignal(sfu.EventType_LEAVE_EVENT)
	hub.BroadCast(p.ID, p.RoomID, leaveRoomE)
}

func (p *Peer) sendLoop() {
	for {
		msg := <-p.SendQ
		if err := p.Stream.Send(msg); err != nil {
			p.Log.Warn("unable to send payload")
		}
	}

}

func (p *Peer) createEventSignal(e sfu.EventType) *sfu.PeerSignal {
	return &sfu.PeerSignal{
		RoomId: p.RoomID,
		PeerId: p.ID,
		Payload: &sfu.PeerSignal_Event{
			Event: e,
		},
	}
}
