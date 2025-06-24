package sfu

import (
	"fmt"

	"github.com/pion/webrtc/v3"
)

func CreateOffer() {
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun1.l.google.com:5349"},
			},
		},
	}

	peerConn, _ := webrtc.NewPeerConnection(config)

	// TODO: add a data track here

	offer, _ := peerConn.CreateOffer(nil)

	peerConn.SetLocalDescription(offer)

	// Gather ICE candidates, waiting line
	ice_can := webrtc.GatheringCompletePromise(peerConn)
	<-ice_can

	local_desp := peerConn.LocalDescription()

	fmt.Println(local_desp.SDP)

}
