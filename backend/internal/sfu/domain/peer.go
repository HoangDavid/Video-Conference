package domain

import (
	sfu "vidcall/api/proto"

	"github.com/pion/webrtc/v3"
)

type Peer struct {
	PC         *webrtc.PeerConnection
	IceCanDone chan struct{}
	Stream     sfu.SFU_SignalServer
}
