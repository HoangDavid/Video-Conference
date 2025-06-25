package service

import (
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

type Peer struct {
	sigAddr string
	config  webrtc.Configuration
	SDP     string
}

func NewPeer(sigAddr string, config webrtc.Configuration) *Peer {

	peerConn, _ := webrtc.NewPeerConnection(config)

	// TODO: Add media/audio track here

	offer, _ := peerConn.CreateOffer(nil)

	peerConn.SetLocalDescription(offer)

	ice_can := webrtc.GatheringCompletePromise(peerConn)

	// Wait to gather ICE Candidates
	<-ice_can

	dsp := peerConn.LocalDescription()

	return &Peer{
		sigAddr: sigAddr,
		config:  config,
		SDP:     dsp.SDP,
	}
}

func (p *Peer) Connect() {
	ws, _, err := websocket.DefaultDialer.Dial(p.sigAddr, nil)

	if err != nil {
		fmt.Print("Error: %$w", err)
		return
	}

	defer ws.Close()

	// Send SDP to signaling
	ws.WriteMessage(websocket.TextMessage, []byte(p.SDP))

}
