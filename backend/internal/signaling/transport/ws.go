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
	"golang.org/x/sync/errgroup"
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

	stream, err := sfuCLient.Signal(ctx)
	if err != nil {
		log.Error(fmt.Sprintf("websocket closed with error: %v", err))
	}

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return onSendSFU(ctx, conn, stream)
	})

	g.Go(func() error {
		return onListenSFU(ctx, conn, stream)
	})

	if err := g.Wait(); err != nil {
		log.Error(fmt.Sprintf("websocket closed with error: %v", err))
	}

	hub.CloseOne(conn, websocket.CloseNormalClosure, "")

}

type signal struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type sdp struct {
	SDP string `json:"sdp"`
}

type ice struct {
	Candidate         string `json:"candidate"`
	SpdMid            string `json:"sdpMid"`
	SpdMLineIndex     uint32 `json:"sdpMLineIndex"`
	UsernameFragmment string `json:"usernameFragment,omitempty"`
}

func onSendSFU(ctx context.Context, conn *websocket.Conn, stream sfu.SFU_SignalClient) error {
	for {
		var msg signal
		err := conn.ReadJSON(&msg)
		if err != nil {
			return err
		}

		switch msg.Type {
		case "offer":
			var offer sdp

			if err := json.Unmarshal(msg.Payload, &offer); err != nil {
				return err
			}
			if err := stream.Send(&sfu.PeerSignal{
				Payload: &sfu.PeerSignal_Sdp{
					Sdp: &sfu.SDP{
						Type: sfu.SdpType_OFFER,
						Sdp:  offer.SDP,
					},
				},
			}); err != nil {
				return err
			}

		case "answer":
			var answer sdp
			if err := json.Unmarshal(msg.Payload, &answer); err != nil {
				return err
			}
			if err := stream.Send(&sfu.PeerSignal{
				Payload: &sfu.PeerSignal_Sdp{
					Sdp: &sfu.SDP{
						Type: sfu.SdpType_OFFER,
						Sdp:  answer.SDP,
					},
				},
			}); err != nil {
				return err
			}

		case "ice":
			var candidate ice

			if err := json.Unmarshal(msg.Payload, &candidate); err != nil {
				return err
			}

			if err := stream.Send(&sfu.PeerSignal{
				Payload: &sfu.PeerSignal_Ice{
					Ice: &sfu.IceCandidate{
						Candidate:        candidate.Candidate,
						SdpMid:           candidate.SpdMid,
						SdpMlineIndex:    candidate.SpdMLineIndex,
						UsernameFragment: candidate.UsernameFragmment,
					},
				},
			}); err != nil {
				return err
			}

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

func onListenSFU(context context.Context, conn *websocket.Conn, stream sfu.SFU_SignalClient) error {

	for {
		resp, err := stream.Recv()

		if err != nil {
			return err
		}

		switch p := resp.Payload.(type) {
		case *sfu.PeerSignal_Sdp:

			raw, err := json.Marshal(sdp{SDP: p.Sdp.Sdp})
			if err != nil {
				return err
			}

			switch p.Sdp.Type {
			case sfu.SdpType_OFFER:
				if err := conn.WriteJSON(signal{
					Type:    "offer",
					Payload: raw,
				}); err != nil {
					return err
				}
			case sfu.SdpType_ANSWER:
				if err := conn.WriteJSON(signal{
					Type:    "answer",
					Payload: raw,
				}); err != nil {
					return err
				}
			}

		case *sfu.PeerSignal_Ice:
			raw, err := json.Marshal(ice{
				Candidate:     p.Ice.Candidate,
				SpdMid:        p.Ice.SdpMid,
				SpdMLineIndex: p.Ice.SdpMlineIndex,
			})

			if err != nil {
				return err
			}

			if err := conn.WriteJSON(signal{
				Type:    "ice",
				Payload: raw,
			}); err != nil {
				return err
			}
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
