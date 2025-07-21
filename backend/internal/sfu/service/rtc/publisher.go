package rtc

import (
	"context"
	"log/slog"
	"time"
	sfu "vidcall/api/proto"
	"vidcall/internal/sfu/domain"

	"github.com/pion/webrtc/v3"
)

type PubConn struct {
	*domain.PubConn
}

// Create conncection for client to push media
func NewPublisher(ctx context.Context, sendQ chan *sfu.PeerSignal, log *slog.Logger, debounceInterval time.Duration) (domain.Publisher, error) {

	conn, err := NewPConn(sendQ, log, debounceInterval)
	if err != nil {
		return nil, err
	}

	pubCtx, pubCancel := context.WithCancel(ctx)

	return &PubConn{
		PubConn: &domain.PubConn{
			Conn:    conn,
			Ctx:     pubCtx,
			Cancel:  pubCancel,
			RecvSdp: make(chan *sfu.PeerSignal_Sdp),
			RecvIce: make(chan *sfu.PeerSignal_Ice),
		},
	}, nil

}

// set up pc callbacks
func (p *PubConn) WireCallBacks() {
	pc := p.Conn.GetPC()
	pc.OnICECandidate(p.Conn.HandleLocalIce)
	pc.OnNegotiationNeeded(nil)
	pc.OnTrack(p.handleOnTrack)
}

func (p *PubConn) Connect() error {

	for {
		select {
		case <-p.Ctx.Done():
			return nil

		case sdp, ok := <-p.RecvSdp:
			if !ok {
				return nil
			}

			if sdp.Sdp.Type == sfu.SdpType_OFFER {
				if err := p.Conn.HandleOffer(sdp); err != nil {
					return err
				}

			}

		case ice, ok := <-p.RecvIce:
			if !ok {
				return nil
			}

			if err := p.Conn.HandleRemoteIce(ice); err != nil {
				return err
			}
		}
	}

}

func (p *PubConn) Disconnect() error {
	//  Cancel all groutines
	if err := p.Conn.Close(); err != nil {
		return err
	}

	p.Cancel()

	close(p.RecvSdp)
	close(p.RecvIce)

	p.Log.Info("Publisher pc disconnected")

	return nil
}

func (p *PubConn) GetLocalAudio() *webrtc.TrackLocalStaticRTP {
	return p.LocalAudio
}

func (p *PubConn) GetLocalVideo() *webrtc.TrackLocalStaticRTP {
	return p.LocalVideo
}

func (p *PubConn) EnqueueSdp(sdp *sfu.PeerSignal_Sdp) {
	p.RecvSdp <- sdp
}

func (p *PubConn) EnqueueIce(ice *sfu.PeerSignal_Ice) {
	p.RecvIce <- ice
}

// set up on track
func (p *PubConn) handleOnTrack(remote *webrtc.TrackRemote, recv *webrtc.RTPReceiver) {
	local, err := webrtc.NewTrackLocalStaticRTP(
		remote.Codec().RTPCodecCapability,
		remote.ID(),
		remote.StreamID(),
	)

	if err != nil {
		p.Log.Error("unable to create new local strack")
		p.Cancel()
		return
	}

	switch remote.Kind() {
	case webrtc.RTPCodecTypeAudio:
		p.LocalAudio = local
	case webrtc.RTPCodecTypeVideo:
		p.LocalVideo = local
	}

	// Pump RTP packets from remote tracks
	go func() {
		for {
			select {
			case <-p.Ctx.Done():
				return
			default:

			}

			pkt, _, err := remote.ReadRTP()
			if err != nil {
				return
			}

			if err := local.WriteRTP(pkt); err != nil {
				p.Log.Warn("unable to send rtp packet to tracks")
			}
		}
	}()
}
