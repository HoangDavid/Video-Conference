package transport

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	sfu "vidcall/api/proto"
	"vidcall/pkg/logger"

	"github.com/gorilla/websocket"
)

type Hub struct {
	conns map[*websocket.Conn]struct{}
	mu    sync.RWMutex
}

func (h *Hub) Add(c *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.conns[c] = struct{}{}
}

func (h *Hub) Remove(c *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.conns, c)
}

func (h *Hub) CloseOne(c *websocket.Conn, code int, reason string) {
	// TODO: error handling
	msg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, reason)
	_ = c.WriteControl(websocket.CloseMessage, msg, time.Now().Add(time.Second))
	defer c.Close()
	h.Remove(c)
	fmt.Println("Close connection")

}

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	hub = &Hub{
		conns: make(map[*websocket.Conn]struct{}),
	}
)

func HandleWS(w http.ResponseWriter, r *http.Request, sfuCLient sfu.SFUClient) {
	ctx := r.Context()
	log := logger.GetLog(ctx).With("layer", "transport")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("unable to upgrade to websocket")
		return
	}

	// TODO: error handle
	stream, _ := sfuCLient.Signal(ctx)

	defer hub.CloseOne(conn, websocket.CloseNormalClosure, "")

	// TODO: ad channel to catch errors
	go onSendSFU(conn, stream)
	go onListenSFU(conn, stream)

}

type Signal struct {
	Type string `json:"type"`
	SDP  string `json:"sdp,omitempty"`
	ICE  string `json:"ice,omitempty"`
}

func onSendSFU(conn *websocket.Conn, stream sfu.SFU_SignalClient) {
	for {
		var msg Signal
		// TODO: Error handle here
		err := conn.ReadJSON(&msg)

		if err != nil {
			fmt.Println(err)
			return
		}

		// TODO: error handle
		switch msg.Type {
		case "offer":
			_ = stream.Send(&sfu.PeerRequest{
				Payload: &sfu.PeerRequest_Offer{
					Offer: &sfu.SDP{
						Type: sfu.SdpType_OFFER,
						Sdp:  msg.SDP,
					},
				},
			})

		case "ice":
			_ = stream.Send(&sfu.PeerRequest{
				Payload: &sfu.PeerRequest_Ice{
					Ice: &sfu.IceCandidate{
						Candidate: msg.ICE,
					},
				},
			})

		default:
		}
	}

}

func onListenSFU(conn *websocket.Conn, stream sfu.SFU_SignalClient) {

	// TODO: error handling
	for {
		resp, _ := stream.Recv()

		switch p := resp.Payload.(type) {
		case *sfu.PeerResponse_Answer:
			_ = conn.WriteJSON(Signal{
				Type: "answer",
				SDP:  p.Answer.Sdp,
			})
		case *sfu.PeerResponse_Ice:
			_ = conn.WriteJSON(Signal{
				Type: "ice",
				SDP:  p.Ice.Candidate,
			})
		default:
		}
	}
}

// TODO:
/*
- Opeinng handshake
- Data transfer
- Closing handshake: send and recieve close control frame
- TCP connection terminated
- ping/pong
*/
