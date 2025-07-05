package transport

import (
	"fmt"
	"net/http"
	"sync"
	"time"
	"vidcall/pkg/logger"
	"vidcall/pkg/utils"

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

func HandleWS(w http.ResponseWriter, r *http.Request) {

	type Signal struct {
		Type string `json:"type"`
		SDP  string `json:"sdp"`
	}

	log := logger.GetLog(r.Context()).With("layer", "transport")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("unable to upgrade to websocket")
		utils.Error(w, http.StatusInternalServerError, "internal error")
		return
	}

	defer hub.CloseOne(conn, websocket.CloseNormalClosure, "")

	var msg Signal

	for {
		// Error handle here

		err := conn.ReadJSON(&msg)

		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(msg.SDP, msg.Type)

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
