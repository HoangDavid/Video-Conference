package wsx

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	sfu "vidcall/api/proto"
	"vidcall/internal/signaling/security"
	"vidcall/pkg/logger"

	"github.com/gorilla/websocket"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/metadata"
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
	Type string `json:"type"`
}

type event struct {
	Name   string `json:"name"`
	PeerID string `json:"peerID"`
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
	claims := security.ClaimsFrom(ctx)

	md := metadata.Pairs(
		"name", claims.Name,
		"peer-id", claims.PeerID,
		"room-id", claims.RoomID,
		"role", claims.Role,
	)
	ctxMD := metadata.NewOutgoingContext(ctx, md)

	log := logger.GetLog(ctx).With("layer", "transport")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("unable to upgrade to websocket")
		return
	}

	// Checking for join/start meeting before conecting to SFU
	intent, first, err := handleFirstMsg(conn, log)
	if err != nil || intent != IntentJoin {
		return
	}

	stream, err := sfuCLient.Signal(ctxMD)
	if err != nil {
		log.Error("unable to create stream to SFU")
		return
	}

	defer stream.CloseSend()

	// Send the first signal
	if err := stream.Send(first); err != nil {
		log.Error("unable to send first signal")
		return
	} else {
		log.Info("send first msg to sfu")
	}

	g, _ := errgroup.WithContext(ctx)

	g.Go(func() error { return onListenClient(conn, stream, log) })
	g.Go(func() error { return onListenSFU(conn, stream, log) })

	if err := g.Wait(); err != nil {
		log.Error(fmt.Sprintf("websocket disconnected with error: %v", err))
	}

	stream.CloseSend()

	// on websocket disconnect
	CloseOne(conn, websocket.CloseNormalClosure, "")
}

func onListenClient(conn *websocket.Conn, stream sfu.SFU_SignalClient, log *slog.Logger) error {

	log = log.With("from", "client")

	var msg signal
	for {
		err := conn.ReadJSON(&msg)
		if err != nil {
			return err
		}

		switch msg.Type {
		case "sdp":
			sdp, err := handleClientSDP(msg.Payload, log)

			if err != nil {
				return err
			}

			if err := stream.Send(sdp); err != nil {
				log.Error("unable to send sdp to sfu")
				return err
			}

			log.Info("sent sdp to sfu")

		case "ice":
			ice, err := handleClientIce(msg.Payload, log)

			if err != nil {
				return err
			}

			if err := stream.Send(ice); err != nil {
				log.Error("unable to send ice to sfu")
				return err
			}

			log.Info("sent ice to sfu")

		case "action":

			// TODO: add permisions
			action, err := handleClientAction(msg.Payload, log)

			if err != nil {
				return err
			}

			if err := stream.Send(action); err != nil {
				log.Error("unable to send action to sfu")
				return err
			}

			log.Info("sent action to sfu")
		}

	}
}

func onListenSFU(conn *websocket.Conn, stream sfu.SFU_SignalClient, log *slog.Logger) error {

	log = log.With("from", "SFU")

	for {
		msg, err := stream.Recv()
		if err != nil {
			return err
		}

		switch pl := msg.Payload.(type) {
		case *sfu.PeerSignal_Sdp:
			sdp, err := handleSfuSDP(pl, log)
			if err != nil {
				return err
			}

			if err := conn.WriteJSON(sdp); err != nil {
				log.Error("unable to send sdp to client")
				return err
			}

			log.Info("sent sdp to client")

		case *sfu.PeerSignal_Ice:
			ice, err := handleSfuIce(pl, log)
			if err != nil {
				return err
			}

			if err := conn.WriteJSON(ice); err != nil {
				log.Error("unable to send ice to client")
				return err
			}

			log.Info("send ice to client")

		case *sfu.PeerSignal_Event:
			event, err := handleSfuEvent(pl, log)
			if err != nil {
				return err
			}

			if err := conn.WriteJSON(event); err != nil {
				log.Error("unablet to send event to client")
				return err
			}

			log.Info("send event to client")
		}
	}
}
