package signaling

import (
	"fmt"
	"net/http"
	"os"

	"vidcall/internal/signaling/infra/mongo"
	"vidcall/internal/signaling/infra/redisx"
	"vidcall/internal/signaling/security"
	"vidcall/internal/signaling/transport"
	"vidcall/pkg/logger"

	_ "github.com/joho/godotenv/autoload"
)

func Execute() {

	issuer := security.NewIssuer(os.Getenv("JWT_SECRET"))

	mux := http.NewServeMux()
	// TODO: add ENV soon

	// Fire up infra: MongoDB and Redis
	mongo.Init(os.Getenv("MONGODB_URI"), os.Getenv("DB_NAME"), 10)
	redisx.Init(os.Getenv("REDIS_URI"), os.Getenv("REDIS_PASSWORD"), 0)

	// mux.HandleFunc("/ws", transport.WsHandler)
	mux.HandleFunc("GET /start_room/{duration}", transport.HandleStartRoom)

	// secured endpoints
	mux.HandleFunc("POST /rooms/{room_id}/join", security.MiddleWare(issuer)(transport.HandleJoinRoom))

	port := os.Getenv("SIGNALING_PORT")
	server := &http.Server{
		Addr:    port,
		Handler: logger.SlogMiddleware(mux), // Slog handle server logging

		/*
			TODO: add TLS config so to use wss:/ and https:/
		*/
	}

	fmt.Println("Server starting at port " + port)
	server.ListenAndServe()

}
