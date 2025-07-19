package rtc

import (
	"log/slog"
	"time"
	sfu "vidcall/api/proto"
	"vidcall/internal/sfu/domain"

	"github.com/pion/webrtc/v3"
)

type Publisher struct {
	*domain.Publisher
	conn *conn
	log  *slog.Logger
}

// Create conncection for client to push media
func NewPublisher(sendQ chan *sfu.PeerSignal, stuns []string, log *slog.Logger, debounceInterval time.Duration) (*Publisher, error) {

	c, err := newPeerConnection(sendQ, stuns, log, debounceInterval)
	if err != nil {
		return nil, err
	}

	return &Publisher{
		Publisher: &domain.Publisher{
			PC: c.pc,
		},
		conn: c,
		log:  log,
	}, nil

}

func (p *Publisher) HandleOffer(sdp *sfu.PeerSignal_Sdp) error {
	if err := p.conn.handleOffer(sdp); err != nil {
		return err
	}

	return nil
}

func (p *Publisher) HandleRemoteIceCandidate(ice *sfu.PeerSignal_Ice) error {
	if err := p.conn.handleRemoteIceCandidate(ice); err != nil {
		return err
	}

	return nil
}

// set up pc callbacks
func (p *Publisher) WireCallBacks() {
	p.PC.OnICECandidate(p.conn.handleLocalIceCandidate)
	p.PC.OnNegotiationNeeded(nil)
	p.PC.OnTrack(p.handleOnTrack)
}

// set up on track
func (p *Publisher) handleOnTrack(remote *webrtc.TrackRemote, recv *webrtc.RTPReceiver) {
	local, err := webrtc.NewTrackLocalStaticRTP(
		remote.Codec().RTPCodecCapability,
		remote.ID(),
		remote.StreamID(),
	)

	if err != nil {
		p.log.Error("unable to create new local strack")
		return
	}

	switch remote.Kind() {
	case webrtc.RTPCodecTypeAudio:
		p.LocalAudio = local
	case webrtc.RTPCodecTypeVideo:
		p.LocalVideo = local
	}

	go func() {
		for {
			pkt, _, err := remote.ReadRTP()
			if err != nil {
				return
			}

			if err := local.WriteRTP(pkt); err != nil {
				p.log.Warn("unable to send rtp packet to tracks")
			}
		}
	}()
}
