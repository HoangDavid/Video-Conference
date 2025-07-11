package domain

import (
	"context"
	sfu "vidcall/api/proto"

	"github.com/pion/webrtc/v3"
)

type Peer struct {
	Ctx        context.Context
	PC         *webrtc.PeerConnection
	IceCanDone chan struct{}
	Stream     sfu.SFU_SignalServer
}
