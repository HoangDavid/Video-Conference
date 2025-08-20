package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	sfu "vidcall/api/proto"
	"vidcall/internal/sfu/domain"
	"vidcall/internal/sfu/service/rtc"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/metadata"
)

type PeerObj struct {
	*domain.PeerObj
}

func NewPeer(ctx context.Context, stream sfu.SFU_SignalServer, poolSize int, log *slog.Logger) (domain.Peer, error) {
	log = log.With("layer", "service")

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

	// peer metadata
	md, _ := metadata.FromIncomingContext(ctx)

	get_md := func(v []string) string {
		if len(v) > 0 {
			return v[0]
		} else {
			return ""
		}
	}

	var r sfu.RoleType
	mdRole := get_md(md.Get("role"))
	if mdRole == "host" {
		r = sfu.RoleType_ROLE_HOST
	} else if mdRole == "guest" {
		r = sfu.RoleType_ROLE_GUEST
	} else if mdRole == "bot" {
		r = sfu.RoleType_ROLE_BOT
	}

	peermd := &domain.PeerMD{
		Name:   get_md(md.Get("name")),
		PeerID: get_md(md.Get("peer-id")),
		RoomID: get_md(md.Get("room-id")),
		Role:   r,
	}

	// wire call backs
	pub.WireCallBacks(peermd.PeerID)
	sub.WireCallBacks()

	pCtx, pCancel := context.WithCancel(ctx)

	return &PeerObj{
		PeerObj: &domain.PeerObj{
			Metadata:   peermd,
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

func (p *PeerObj) GetMetaData() *domain.PeerMD {
	return p.Metadata
}

func (p *PeerObj) Pub() domain.Publisher {
	return p.Publisher
}

func (p *PeerObj) Sub() domain.Subscriber {
	return p.Subscriber
}

func (p *PeerObj) Connect() error {
	g, _ := errgroup.WithContext(p.Ctx)

	// start uplink/downlink peer connection
	g.Go(func() error { return p.Publisher.Connect() })
	g.Go(func() error { return p.Subscriber.Connect() })

	// start send loop and on event loop
	g.Go(func() error { return p.sendCycle() })
	g.Go(func() error { return p.eventCycle() })

	// main loop
	g.Go(func() error {
		for {
			select {
			case <-p.Ctx.Done():
				return nil
			default:
				msg, err := p.Stream.Recv()

				if err != nil {
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
					err := p.handleActions(pl)
					if err != nil {
						return err
					}
				}
			}

		}
	})

	err := g.Wait()
	if err != nil {
		return err
	}

	return nil
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

	p.Log.Info("Peer disconnected!")
	return nil
}

func (p *PeerObj) EnqueueEvent(event *sfu.PeerSignal_Event) {
	select {
	case p.EventQ <- event:
	default:
	}
}

func (p *PeerObj) EnqueueSend(msg *sfu.PeerSignal) {
	select {
	case p.SendQ <- msg:
	default:
	}
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
				errMsg := fmt.Sprintf("unable to send signal: %v", err)
				p.Log.Error(errMsg)
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
