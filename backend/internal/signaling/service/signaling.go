package service

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

type Signaling struct {
	Addr string
}

func NewSignaling(signalingURL string) *Signaling {
	return &Signaling{
		Addr: signalingURL,
	}
}

// Websocket upgarder
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (s *Signaling) WsHandler(w http.ResponseWriter, r *http.Request) {

	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		fmt.Println("Websocket upgrade failed: %w", err)
		return
	}

	// revoke when client connects
	fmt.Println("Connected successfully!")

	defer conn.Close()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}

		fmt.Println("Sent SDP to client")

		fmt.Print(msg)

	}
}
