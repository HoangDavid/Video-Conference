package wsx

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	sfu "vidcall/api/proto"
	"vidcall/pkg/logger"

	"github.com/gorilla/websocket"
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

	stream, _ := sfuCLient.Signal(ctx)

	go onListenClient(log, conn, stream)

	// go onListenSFU(ctx, conn, stream)

	// on websocket disconnect
	stream.CloseSend()
	CloseOne(conn, websocket.CloseNormalClosure, "")

}

func onListenClient(log *slog.Logger, conn *websocket.Conn, stream sfu.SFU_SignalClient) error {
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
				log.Error("unable to send sdp to SFU")
				return err
			}

		case "ice":
			ice, err := handleClientIce(msg.Payload)

			if err != nil {
				return err
			}

			if err := stream.Send(ice); err != nil {
				log.Error("unable to send ice to SFU")
				return err
			}

		case "action":
			action, err := handleClientAction(msg.Payload)

			if err != nil {
				return err
			}

			if err := stream.Send(action); err != nil {
				log.Error("unable to send action to SFU")
				return err
			}
		}

	}
}

func onListenSFU(log *slog.Logger, conn *websocket.Conn, stream sfu.SFU_SignalClient) error {

}
