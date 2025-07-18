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
	stuns := hub.GetHub().Stuns

	duration := time.Duration(50 * time.Millisecond)
	pub, err := rtc.NewPublisher(stream, stuns, log, duration)
	if err != nil {
		return nil, err
	}

	sub, err := rtc.NewSubscriber(stream, stuns, log, poolSize, duration)
	if err != nil {
		return nil, err
	}

	// wire call backs
	pub.WireCallBacks()
	sub.WireCallBacks()

	return &Peer{
		Peer: &domain.Peer{
			ID:     peerID,
			Log:    log,
			Stream: stream,
		},
		publisher: pub,
		subcriber: sub,
	}, nil
}
