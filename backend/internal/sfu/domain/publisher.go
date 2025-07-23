package domain

import (
	"context"
	"log/slog"
	sfu "vidcall/api/proto"

	"github.com/pion/webrtc/v3"
)

type Publisher interface {
	WireCallBacks()
	Connect() error
	Disconnect() error
	AttachDetector(id string, d Detector)
	GetLocalAudio() *webrtc.TrackLocalStaticRTP
	GetLocalVideo() *webrtc.TrackLocalStaticRTP
	EnqueueSdp(sdp *sfu.PeerSignal_Sdp)
	EnqueueIce(ice *sfu.PeerSignal_Ice)
}

type PubConn struct {
	Conn       Connection
	Ctx        context.Context
	Cancel     context.CancelFunc
	Log        *slog.Logger
	Detector   Detector
	LocalAudio *webrtc.TrackLocalStaticRTP
	LocalVideo *webrtc.TrackLocalStaticRTP
	RecvSdp    chan *sfu.PeerSignal_Sdp
	RecvIce    chan *sfu.PeerSignal_Ice
}
