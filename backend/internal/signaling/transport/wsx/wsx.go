package wsx

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	sfu "vidcall/api/proto"
	"vidcall/pkg/logger"

	"github.com/gorilla/websocket"
	"golang.org/x/sync/errgroup"
)

type signal struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type sdp struct {
	Pc   string `json:"pc"`
	Type string `json:"type"`
	SDP  string `json:"sdp"`
}

type ice struct {
	Pc                string `json:"pc"`
	Candidate         string `json:"candidate"`
	SdpMid            string `json:"sdpMid"`
	SdpMLineIndex     uint32 `json:"sdpMLineIndex"`
	UsernameFragmment string `json:"usernameFragment,omitempty"`
}

type action struct {
	PeerID string `json:"peerID"`
	RoomID string `json:"roomID"`
	Type   string `json:"type"`
}

type event struct {
	PeerID string `json:"peerID"`
	RoomID string `json:"roomID"`
	Type   string `json:"type"`
}

func CloseOne(c *websocket.Conn, code int, reason string) {
	// TODO: error handling
	msg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, reason)
	_ = c.WriteControl(websocket.CloseMessage, msg, time.Now().Add(time.Second))
	defer c.Close()
	fmt.Println("Close connection")

}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func HandleWS(w http.ResponseWriter, r *http.Request, sfuCLient sfu.SFUClient) {
	ctx := r.Context()
	log := logger.GetLog(ctx).With("layer", "transport")
	conn, _ := upgrader.Upgrade(w, r, nil)

	stream, err := sfuCLient.Signal(ctx)
	if err != nil {
		log.Error("unable to create stream to SFU")
		return
	}

	g, _ := errgroup.WithContext(ctx)

	g.Go(func() error { return onListenClient(conn, stream) })
	g.Go(func() error { return onListenSFU(conn, stream) })

	if err := g.Wait(); err != nil {
		log.Error(fmt.Sprintf("websocket disconnected with error: %v", err))
	}
	// on websocket disconnect
	stream.CloseSend()
	CloseOne(conn, websocket.CloseNormalClosure, "")

}

func onListenClient(conn *websocket.Conn, stream sfu.SFU_SignalClient) error {
	var msg signal
	for {
		err := conn.ReadJSON(&msg)
		if err != nil {
			return err
		}

		switch msg.Type {
		case "sdp":
			sdp, err := handleClientSDP(msg.Payload)

			if err != nil {
				return err
			}

			if err := stream.Send(sdp); err != nil {
				return err
			}

		case "ice":
			ice, err := handleClientIce(msg.Payload)

			if err != nil {
				return err
			}

			if err := stream.Send(ice); err != nil {
				return err
			}

		case "action":
			action, err := handleClientAction(msg.Payload)

			if err != nil {
				return err
			}

			if err := stream.Send(action); err != nil {
				return err
			}
		}

	}
}

func onListenSFU(conn *websocket.Conn, stream sfu.SFU_SignalClient) error {
	for {
		msg, err := stream.Recv()
		if err != nil {
			return err
		}

		switch pl := msg.Payload.(type) {
		case *sfu.PeerSignal_Sdp:
			sdp, err := handleSfuSDP(pl)
			if err != nil {
				return err
			}

			if err := conn.WriteJSON(sdp); err != nil {
				return err
			}

		case *sfu.PeerSignal_Ice:
			ice, err := handleSfuIce(pl)
			if err != nil {
				return err
			}

			if err := conn.WriteJSON(ice); err != nil {
				return err
			}
		case *sfu.PeerSignal_Event:
			event, err := handleSfuEvent(pl)
			if err != nil {
				return err
			}

			if err := conn.WriteJSON(event); err != nil {
				return err
			}
		}
	}

}
