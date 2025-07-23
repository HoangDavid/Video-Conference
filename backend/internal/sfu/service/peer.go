package service

import (
	"context"
	"time"

	sfu "vidcall/api/proto"
	"vidcall/internal/sfu/domain"
	"vidcall/internal/sfu/service/rtc"
	"vidcall/pkg/logger"
)

type PeerObj struct {
	*domain.PeerObj
}

func NewPeer(ctx context.Context, stream sfu.SFU_SignalServer, poolSize int) (domain.Peer, error) {
	log := logger.GetLog(ctx).With("layer", "service")

	// Create channel to send msg and events
	sendQ := make(chan *sfu.PeerSignal, 64)
	eventQ := make(chan *sfu.PeerSignal_Event, 64)

	duration := time.Duration(50 * time.Millisecond)

	pub, err := rtc.NewPublisher(ctx, sendQ, log, duration)
	if err != nil {
		return nil, err
	}

	sub, err := rtc.NewSubscriber(ctx, sendQ, log, poolSize, duration)
	if err != nil {
		return nil, err
	}

	// wire call backs
	pub.WireCallBacks()
	sub.WireCallBacks()

	pCtx, pCancel := context.WithCancel(ctx)

	return &PeerObj{
		PeerObj: &domain.PeerObj{
			Log:        log,
			Ctx:        pCtx,
			Cancel:     pCancel,
			Stream:     stream,
			Publisher:  pub,
			Subscriber: sub,
			SendQ:      sendQ,
			EventQ:     eventQ,
		},
	}, nil
}

func (p *PeerObj) GetID() string {
	return p.ID
}

func (p *PeerObj) Pub() domain.Publisher {
	return p.Publisher
}

func (p *PeerObj) Sub() domain.Subscriber {
	return p.Subscriber
}

func (p *PeerObj) Connect() error {
	// start uplink/downlink peer connection
	go p.Pub().Connect()
	go p.Sub().Connect()

	// start send loop and on event loop
	go p.sendCycle()
	go p.eventCycle()

	for {
		select {
		case <-p.Ctx.Done():
			return nil
		default:
			msg, err := p.Stream.Recv()
			if err != nil {
				p.Log.Error("peer unable to recieve msg")
				return err
			}

			switch pl := msg.Payload.(type) {
			case *sfu.PeerSignal_Sdp:
				pc := pl.Sdp.Pc

				if pc == sfu.PcType_PUB {
					p.Publisher.EnqueueSdp(pl)
				}

				if pc == sfu.PcType_SUB {
					p.Subscriber.EnqueueSdp(pl)
				}

			case *sfu.PeerSignal_Ice:
				pc := pl.Ice.Pc

				if pc == sfu.PcType_PUB {
					p.Publisher.EnqueueIce(pl)
				}

				if pc == sfu.PcType_SUB {
					p.Subscriber.EnqueueIce(pl)
				}

			case *sfu.PeerSignal_Action:
				p.handleActions(pl)
			}
		}

	}

}

func (p *PeerObj) Disconnect() error {
	if err := p.Publisher.Disconnect(); err != nil {
		return err
	}

	if err := p.Subscriber.Disconnect(); err != nil {
		return err
	}

	p.Cancel()

	close(p.SendQ)
	close(p.EventQ)

	return nil
}

func (p *PeerObj) EnqueueEvent(event *sfu.PeerSignal_Event) {
	p.EventQ <- event
}

func (p *PeerObj) sendCycle() error {
	for {
		select {
		case <-p.Ctx.Done():
			return nil
		case msg, ok := <-p.SendQ:
			if !ok {
				return nil
			}

			if err := p.Stream.Send(msg); err != nil {
				return err
			}
		}
	}
}

func (p *PeerObj) eventCycle() error {
	for {
		select {
		case <-p.Ctx.Done():
			return nil
		case msg, ok := <-p.EventQ:
			if !ok {
				return nil
			}

			p.handleEvents(msg)
		}
	}
}

func (p *PeerObj) handleActions(act *sfu.PeerSignal_Action) error {

	actType := act.Action.Type
	roomID := act.Action.Roomid

	if p.ID == "" {
		p.ID = act.Action.Peerid
	}

	switch actType {
	case sfu.ActionType_START_ROOM:
		p.handleStartRoom(roomID)

	case sfu.ActionType_JOIN:
		p.handleJoinRoom(roomID)

	case sfu.ActionType_LEAVE:
		p.handleLeaveRoom(roomID)

	case sfu.ActionType_END_ROOM:
		p.handleEndRoom(roomID)
	}

	return nil
}

func (p *PeerObj) handleEvents(evt *sfu.PeerSignal_Event) error {
	evtType := evt.Event.Type

	switch evtType {
	case sfu.EventType_ROOM_ACTIVE:
		p.handleRoomActiveEvent(evt)

	case sfu.EventType_ROOM_INACTIVE:
		p.handleRoomInactiveEvent(evt)

	case sfu.EventType_JOIN_EVENT:
		if err := p.handleJoinEvent(evt); err != nil {
			return err
		}

	case sfu.EventType_LEAVE_EVENT:
		if err := p.handleLeaveEvent(evt); err != nil {
			return err
		}

	case sfu.EventType_ROOM_ENEDED:
		p.handleRoomEndedEvent()
	}
	return nil
}
