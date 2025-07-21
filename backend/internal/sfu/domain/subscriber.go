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
	Subscribe(peer Peer) error
	Unsubscribe(peerID string) error
	SubscribeRoom(ownerID string, room Room) error
	UnsubscribeRoom(ownerID string, room Room) error
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
	IdleAudios []*webrtc.RTPTransceiver
	IdleVideos []*webrtc.RTPTransceiver
	Dub        *webrtc.RTPTransceiver
	Sub        *webrtc.DataChannel

	ActiveVideos map[string]*webrtc.RTPTransceiver
	ActiveAudios map[string]*webrtc.RTPTransceiver

	RecvSdp chan *sfu.PeerSignal_Sdp
	RecvIce chan *sfu.PeerSignal_Ice
}
