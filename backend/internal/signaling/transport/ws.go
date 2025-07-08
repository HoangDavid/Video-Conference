package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	sfu "vidcall/api/proto"
	"vidcall/internal/signaling/service"
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

	// TODO: add channel to catch errors
	go onSendSFU(ctx, conn, stream)
	go onListenSFU(ctx, conn, stream)

	// TODO: something here to keep live

}

type Signal struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

func onSendSFU(ctx context.Context, conn *websocket.Conn, stream sfu.SFU_SignalClient) {
	for {
		var msg Signal
		// TODO: Error handle here
		err := conn.ReadJSON(&msg)
		if err != nil {
			fmt.Println(err) // TODO: change this to logging
			return
		}

		// TODO: error handle
		switch msg.Type {
		case "offer":
			var offer struct {
				SDP string `json:"offer"`
			}

			_ = json.Unmarshal(msg.Payload, &offer)
			_ = stream.Send(&sfu.PeerRequest{
				Payload: &sfu.PeerRequest_Offer{
					Offer: &sfu.SDP{
						Type: sfu.SdpType_OFFER,
						Sdp:  offer.SDP,
					},
				},
			})

		case "ice":
			var ice struct {
				Candidate string `json:"candidate"`
			}

			_ = json.Unmarshal(msg.Payload, &ice)
			_ = stream.Send(&sfu.PeerRequest{
				Payload: &sfu.PeerRequest_Ice{
					Ice: &sfu.IceCandidate{
						Candidate: ice.Candidate,
					},
				},
			})

		case "start_room":
			service.StartRoom(ctx)
		case "join":
			service.JoinRoom(ctx)
		case "leave":
			service.LeaveRoom(ctx)
		case "mute_video":
		case "mute_audio":
		default:
		}
	}

}

func onListenSFU(context context.Context, conn *websocket.Conn, stream sfu.SFU_SignalClient) {

	// TODO: error handling
	for {
		resp, _ := stream.Recv()

		switch p := resp.Payload.(type) {
		case *sfu.PeerResponse_Answer:

			type answer struct {
				SDP string
			}

			_ = conn.WriteJSON(Signal{
				Type: "answer",
				SDP:  p.Answer.Sdp,
			})
		case *sfu.PeerResponse_Ice:
			_ = conn.WriteJSON(Signal{
				Type: "ice",
				ICE:  p.Ice.Candidate,
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
