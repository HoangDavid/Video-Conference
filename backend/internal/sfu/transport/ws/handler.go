package transport

import (
	"vidcall/internal/sfu/service"

	"github.com/pion/webrtc/v3"
)

func Init() {
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun1.l.google.com:5349"},
			},
		},
	}
	peer := service.NewPeer("ws://localhost:8080/ws", config)

	peer.Connect()

}
