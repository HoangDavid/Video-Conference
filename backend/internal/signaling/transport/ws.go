package transport

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(*http.Request) bool {
		return true
	},
}

// Handling SDP, ICE exchanges
func WsHandler(w http.ResponseWriter, r *http.Request) {
	ws, _ := upgrader.Upgrade(w, r, nil)

	// TODO: error handler here
	defer ws.Close()

	for {
		_, data, _ := ws.ReadMessage()

		// TODO: error handler here

		fmt.Println("Recieved: %w", data)
	}

}
