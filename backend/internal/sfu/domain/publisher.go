package domain

import (
	"context"
	"log/slog"
	sfu "vidcall/api/proto"

	"github.com/pion/webrtc/v3"
)

type Publisher interface {
	WireCallBacks(peerID string)
	PumpVideo(ctx context.Context, local *webrtc.TrackLocalStaticRTP, tx *webrtc.RTPTransceiver)
	Connect() error
	Disconnect() error
	AttachDetector(id string, d Detector)
	GetLocalAV() *PubAV
	EnqueueSdp(sdp *sfu.PeerSignal_Sdp)
	EnqueueIce(ice *sfu.PeerSignal_Ice)
}

type PubConn struct {
	Conn     Connection
	Ctx      context.Context
	Cancel   context.CancelFunc
	Log      *slog.Logger
	Detector Detector

	AV      *PubAV
	RecvSdp chan *sfu.PeerSignal_Sdp
	RecvIce chan *sfu.PeerSignal_Ice
}

type PubAV struct {
	Video *webrtc.TrackRemote
	Audio *webrtc.TrackRemote
}
