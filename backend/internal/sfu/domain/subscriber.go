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
	SubscribeRoom(subcriberID string, room Room) error
	Subscribe(peer Peer) error
	Unsubscribe(peer string) error
	EnqueueSdp(sdp *sfu.PeerSignal_Sdp)
	EnqueueIce(sdp *sfu.PeerSignal_Ice)
}

type SubConn struct {
	Conn   Connection
	Mu     sync.RWMutex
	Ctx    context.Context
	Cancel context.CancelFunc
	Log    *slog.Logger

	Videos  *SubVideo
	RecvSdp chan *sfu.PeerSignal_Sdp
	RecvIce chan *sfu.PeerSignal_Ice
}

type SubVideo struct {
	IDOrder         []string
	IDToVideoTracks map[string]*webrtc.TrackLocalStaticRTP
	IDToAudioTracks map[string]*webrtc.TrackLocalStaticRTP

	Slots       map[int]*Slot
	OwnerToSlot map[string]int
	SlotToOwner map[int]string
}

type Slot struct {
	VideoTx    *webrtc.RTPTransceiver
	AudioTx    *webrtc.RTPTransceiver
	PumpCtx    context.Context
	PumpCancel context.CancelFunc
}
