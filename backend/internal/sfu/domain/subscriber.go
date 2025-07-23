package domain

import (
	"context"
	"log/slog"
	"sync"
	sfu "vidcall/api/proto"

	"github.com/pion/webrtc/v3"
)

type Subscriber interface {
	WireCallBacks()
	Connect() error
	Disconnect() error
	SubscribeVideo(peer Peer) error
	UnsubscribeVideo(peerID string) error
	SubcribeAudio(peer Peer)
	EnqueueSdp(sdp *sfu.PeerSignal_Sdp)
	EnqueueIce(sdp *sfu.PeerSignal_Ice)
}

type SubConn struct {
	Conn   Connection
	Mu     sync.RWMutex
	Ctx    context.Context
	Cancel context.CancelFunc
	Log    *slog.Logger

	Direction  webrtc.RTPTransceiverInit
	AudioOut   *webrtc.RTPTransceiver
	IdleVideos []*webrtc.RTPTransceiver
	Sub        *webrtc.DataChannel

	ActiveVideos map[string]*webrtc.RTPTransceiver

	RecvSdp chan *sfu.PeerSignal_Sdp
	RecvIce chan *sfu.PeerSignal_Ice
}
