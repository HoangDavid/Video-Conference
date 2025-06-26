package signaling

import (
	"fmt"
	"net/http"

	"vidcall/internal/signaling/transport"
)

func Execute() {
	mux := http.NewServeMux()

	mux.HandleFunc("/ws", transport.WsHandler)
	mux.HandleFunc("/start_room/{duration}", transport.HandleStartRoom)
	mux.HandleFunc("/rooms/{room_id}/join", transport.HandleJoinRoom)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,

		/*
			TODO: add TLS config so to use wss:/ and https:/
		*/
	}

	fmt.Println("Server starting at port :8080")
	server.ListenAndServe()

}
