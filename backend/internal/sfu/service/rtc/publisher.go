package rtc

import (
	"context"
	"log/slog"
	"sync"
	"time"
	sfu "vidcall/api/proto"
	"vidcall/internal/sfu/domain"

	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v3"
)

type PubConn struct {
	*domain.PubConn
	wg sync.WaitGroup
}

// Create conncection for client to push media
func NewPublisher(ctx context.Context, sendQ chan *sfu.PeerSignal, log *slog.Logger, debounceInterval time.Duration) (domain.Publisher, error) {

	conn, err := NewPConn(sendQ, log, debounceInterval, true)
	if err != nil {
		return nil, err
	}

	pubCtx, pubCancel := context.WithCancel(ctx)

	p := &PubConn{
		PubConn: &domain.PubConn{
			Log:     log,
			Conn:    conn,
			AV:      &domain.PubAV{},
			Ctx:     pubCtx,
			Cancel:  pubCancel,
			RecvSdp: make(chan *sfu.PeerSignal_Sdp),
			RecvIce: make(chan *sfu.PeerSignal_Ice),
		},
	}

	// wait group for publisher audio and video tracks attachment
	p.wg.Add(2)

	return p, nil

}

// set up pc callbacks
func (p *PubConn) WireCallBacks(peerID string) {
	pc := p.Conn.GetPC()
	pc.OnICECandidate(func(c *webrtc.ICECandidate) {
		p.Conn.HandleLocalIce(c, sfu.PcType_PUB)

	})
	pc.OnTrack(func(remote *webrtc.TrackRemote, recv *webrtc.RTPReceiver) {
		p.handleOnTrack(remote, recv)
	})

}

// start ice/sdp exchange for pc
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

// tear down goroutines and pc
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

func (p *PubConn) GetLocalAV() *domain.PubAV {
	p.wg.Wait()
	return p.AV
}

func (p *PubConn) EnqueueSdp(sdp *sfu.PeerSignal_Sdp) {
	select {
	case p.RecvSdp <- sdp:
	default:
	}
}

func (p *PubConn) EnqueueIce(ice *sfu.PeerSignal_Ice) {
	select {
	case p.RecvIce <- ice:
	default:
	}
}

// set up on track
func (p *PubConn) handleOnTrack(remote *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {

	switch remote.Kind() {
	case webrtc.RTPCodecTypeVideo:
		p.AV.Video = remote
		p.wg.Done()

	case webrtc.RTPCodecTypeAudio:
		p.AV.Audio = remote
		p.wg.Done()
	}
}

func (p *PubConn) PumpAudio(ctx context.Context, local *webrtc.TrackLocalStaticRTP) {
	remote := p.GetLocalAV().Audio
	for {
		select {
		case <-ctx.Done():
			p.Log.Info("stop pumping audio")
			return
		default:
		}
		pkt, _, err := remote.ReadRTP()

		if err != nil {
			p.Log.Error("unable to read audio RTP packet")
			return
		}

		if err := local.WriteRTP(pkt); err != nil {
			p.Log.Error("unable to send audio RTP packet")
			return
		}
	}
}

// pump video to subcribers
func (p *PubConn) PumpVideo(ctx context.Context, local *webrtc.TrackLocalStaticRTP, tx *webrtc.RTPTransceiver) {
	remote := p.GetLocalAV().Video

	go p.checkRTCP(ctx, tx, remote.SSRC())

	for {
		select {
		case <-ctx.Done():
			p.Log.Info("stop pumping video")
			return
		default:
		}
		pkt, _, err := remote.ReadRTP()

		if err != nil {
			p.Log.Error("unable to read video RTP packet")
			return
		}

		if err := local.WriteRTP(pkt); err != nil {
			p.Log.Error("unable to send video RTP packet")
			return
		}
	}
}

func (p *PubConn) checkRTCP(ctx context.Context, tx *webrtc.RTPTransceiver, upSRRC webrtc.SSRC) {

	var lastPLI time.Time
	const minInt = 1000 * time.Millisecond

	for {
		select {
		case <-ctx.Done():
			p.Log.Info("stop RTCP checking")
			return
		default:
		}

		pkts, _, err := tx.Sender().ReadRTCP()

		if err != nil {
			p.Log.Error("unable to read read RTCP")
			return
		}

		for _, pkt := range pkts {
			if pli, ok := pkt.(*rtcp.PictureLossIndication); ok {
				if time.Since(lastPLI) >= minInt {
					lastPLI = time.Now()
					pli.MediaSSRC = uint32(upSRRC)
					if err := p.Conn.GetPC().WriteRTCP([]rtcp.Packet{pli}); err != nil {
						p.Log.Error("Failed to write pli RTCP")
					}

					p.Log.Info("sent key frame")
				}
			}
		}

	}
}
