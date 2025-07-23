package rtc

import (
	"context"
	"log/slog"
	"sync"
	"time"
	sfu "vidcall/api/proto"
	"vidcall/internal/sfu/domain"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
)

type PubConn struct {
	*domain.PubConn
	wg sync.WaitGroup
	ID string
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
			Conn:    conn,
			Ctx:     pubCtx,
			Cancel:  pubCancel,
			RecvSdp: make(chan *sfu.PeerSignal_Sdp),
			RecvIce: make(chan *sfu.PeerSignal_Ice),
		},
	}

	// wait group for publisher audio and video tracks
	p.wg.Add(2)

	return p, nil

}

// set up pc callbacks
func (p *PubConn) WireCallBacks() {
	pc := p.Conn.GetPC()
	pc.OnICECandidate(p.Conn.HandleLocalIce)
	pc.OnNegotiationNeeded(nil)
	pc.OnTrack(p.handleOnTrack)
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

func (p *PubConn) GetLocalAudio() *webrtc.TrackLocalStaticRTP {
	p.wg.Wait()
	return p.LocalAudio
}

func (p *PubConn) GetLocalVideo() *webrtc.TrackLocalStaticRTP {
	p.wg.Wait()
	return p.LocalVideo
}

func (p *PubConn) EnqueueSdp(sdp *sfu.PeerSignal_Sdp) {
	p.RecvSdp <- sdp
}

func (p *PubConn) EnqueueIce(ice *sfu.PeerSignal_Ice) {
	p.RecvIce <- ice
}

func (p *PubConn) AttachDetector(id string, d domain.Detector) {
	p.ID = id
	p.Detector = d
}

// set up on track
func (p *PubConn) handleOnTrack(remote *webrtc.TrackRemote, recv *webrtc.RTPReceiver) {

	local, err := webrtc.NewTrackLocalStaticRTP(
		remote.Codec().RTPCodecCapability,
		remote.ID(),
		remote.StreamID(),
	)

	if err != nil {
		p.Log.Error("unable to create new video local track")
		p.Cancel()
		return
	}

	switch remote.Kind() {

	case webrtc.RTPCodecTypeVideo:
		p.LocalVideo = local
		p.wg.Done()

		go p.pumpVideo(remote)

	case webrtc.RTPCodecTypeAudio:
		p.LocalAudio = local
		p.wg.Done()

		var audioLevelExtID uint8
		for _, ext := range recv.GetParameters().HeaderExtensions {
			if ext.URI == p.Conn.GetAudioURI() {
				audioLevelExtID = uint8(ext.ID)
				break
			}
		}

		go p.pumpAudio(remote, audioLevelExtID)
	}
}

// pump video to subcribers
func (p *PubConn) pumpVideo(remote *webrtc.TrackRemote) {

	for {
		pkt, _, err := remote.ReadRTP()
		if err != nil {
			p.Log.Error("unable to read video RTP packet")
			return
		}

		if err := p.LocalVideo.WriteRTP(pkt); err != nil {
			p.Log.Error("unable to send video RTP packet")
			return
		}

	}
}

// pump audio to subcribers
func (p *PubConn) pumpAudio(remote *webrtc.TrackRemote, extID uint8) {

	for {
		pkt, _, err := remote.ReadRTP()
		if err != nil {
			p.Log.Error("unable to read audio RTP packet")
			return
		}

		// Sample audio to find active speaker
		// TODO: need a better place to put this
		if p.Detector != nil {
			lvl := p.audioLevel(pkt, extID)
			p.Detector.Sample(p.ID, lvl)
		}

		if err := p.LocalAudio.WriteRTP(pkt); err != nil {
			p.Log.Error("unable to send audio RTP packet")
			return
		}

	}
}

// get audio level from packet
func (p *PubConn) audioLevel(pkt *rtp.Packet, extID uint8) int {
	b := pkt.GetExtension(extID)
	if len(b) == 0 {
		return 127
	}
	return int(b[0] & 0x7F)
}
