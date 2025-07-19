package peer

import (
	"context"
	"time"

	sfu "vidcall/api/proto"
	"vidcall/internal/sfu/domain"
	"vidcall/internal/sfu/service/hub"
	"vidcall/internal/sfu/service/peer/rtc"
	"vidcall/pkg/logger"
)

type Peer struct {
	*domain.Peer
	publisher *rtc.Publisher
	subcriber *rtc.Subscriber
}

func NewPeer(ctx context.Context, peerID string, stream sfu.SFU_SignalServer, poolSize int) (*Peer, error) {
	log := logger.GetLog(ctx).With("layer", "service")
	stuns := hub.Hub().Stuns

	sendQ := make(chan *sfu.PeerSignal)
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
			ID:         peerID,
			Log:        log,
			Stream:     stream,
			Publisher:  pub.Publisher,
			Subscriber: sub.Subscriber,
			SendQ:      sendQ,
		},
		publisher: pub,
		subcriber: sub,
	}, nil
}

func (p *Peer) Connect() error {

	p.subcriber.SendOffer()

	for {
		msg, err := p.Stream.Recv()
		if err != nil {
			p.Log.Error("unable to recieve msg from stream")
		}

		switch pl := msg.Payload.(type) {
		case *sfu.PeerSignal_Sdp:
			if pl.Sdp.Pc == sfu.PcType_PUB && pl.Sdp.Type == sfu.SdpType_OFFER {
				if err := p.publisher.HandleOffer(pl.Sdp.Sdp); err != nil {
					return err
				}
			}

			if pl.Sdp.Pc == sfu.PcType_SUB && pl.Sdp.Type == sfu.SdpType_ANSWER {
				if err := p.subcriber.HandleAnswer(pl.Sdp.Sdp); err != nil {
					return err
				}
			}

		case *sfu.PeerSignal_Ice:
			if pl.Ice.Pc == sfu.PcType_PUB {
				if err := p.publisher.HandleRemoteIceCandidate(pl); err != nil {
					return err
				}
			}

			if pl.Ice.Pc == sfu.PcType_SUB {
				if err := p.subcriber.HandleRemoteIceCandidate(pl); err != nil {
					return err
				}
			}
		case *sfu.PeerSignal_Action:
			p.handleAction(pl)

		case *sfu.PeerSignal_Event:
			p.handleEvent(pl)

		}
	}
}

func (p *Peer) Disconnect() {

}

func (p *Peer) handleAction(action *sfu.PeerSignal_Action) {
	actionType := action.Action.Type

	switch actionType {
	case sfu.ActionType_JOIN:
	case sfu.ActionType_LEAVE:
	}
}

func (p *Peer) handleEvent(event *sfu.PeerSignal_Event) {

}

func (p *Peer) sendLoop() {
	for {
		msg := <-p.SendQ
		if err := p.Stream.Send(msg); err != nil {
			p.Log.Warn("unable to send payload")
		}
	}

}
