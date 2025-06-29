package signaling

import (
	"fmt"
	"net/http"

	"vidcall/internal/signaling/infra/mongo"
	"vidcall/internal/signaling/infra/redisx"
	"vidcall/internal/signaling/transport"
	"vidcall/pkg/logger"
)

func Execute() {

	mux := http.NewServeMux()

	// TODO: add ENV soon

	// Fire up infra: MongoDB and Redis
	mongo.Init("mongodb://localhost:27017", "Meeting", 10)
	redisx.Init("localhost:6379", "", 0)

	// mux.HandleFunc("/ws", transport.WsHandler)
	mux.HandleFunc("/start_room/{duration}", transport.HandleStartRoom)
	mux.HandleFunc("/rooms/{room_id}/join", transport.HandleJoinRoom)

	server := &http.Server{
		Addr:    ":8080",
		Handler: logger.SlogMiddleware(mux), // Slog handle server logging

		/*
			TODO: add TLS config so to use wss:/ and https:/
		*/
	}

	fmt.Println("Server starting at port :8080")
	server.ListenAndServe()

}
