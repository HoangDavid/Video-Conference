package domain

import (
	"context"
	"log/slog"
	sfu "vidcall/api/proto"

	"github.com/pion/webrtc/v3"
)

type Publisher interface {
	WireCallBacks(peerID string)
	PumpAudio(ctx context.Context, local *webrtc.TrackLocalStaticRTP)
	PumpVideo(ctx context.Context, local *webrtc.TrackLocalStaticRTP, tx *webrtc.RTPTransceiver)
	Connect() error
	Disconnect() error
	GetLocalAV() *PubAV
	EnqueueSdp(sdp *sfu.PeerSignal_Sdp)
	EnqueueIce(ice *sfu.PeerSignal_Ice)
}

type PubConn struct {
	Conn   Connection
	Ctx    context.Context
	Cancel context.CancelFunc
	Log    *slog.Logger

	AV      *PubAV
	RecvSdp chan *sfu.PeerSignal_Sdp
	RecvIce chan *sfu.PeerSignal_Ice
}

type PubAV struct {
	Video *webrtc.TrackRemote
	Audio *webrtc.TrackRemote
}
