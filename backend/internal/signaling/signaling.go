package signaling

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"

	"vidcall/internal/signaling/infra/mongox"
	"vidcall/internal/signaling/infra/redisx"
	"vidcall/internal/signaling/security"
	"vidcall/internal/signaling/transport"
	"vidcall/pkg/logger"

	_ "github.com/joho/godotenv/autoload"
)

func Execute() {

	issuer := security.NewIssuer(os.Getenv("JWT_SECRET"))

	mux := http.NewServeMux()

	// Load TLS cert and key
	cert := os.Getenv("TLS_CERT")
	key := os.Getenv("TLS_KEY")
	if cert == "" || key == "" {
		log.Fatalf("TLS_CERT or TLS_KEY are not set")
	}

	// Fire up infra: MongoDB and Redis
	mongox.Init(os.Getenv("MONGODB_URI"), os.Getenv("DB_NAME"), 10)
	redisx.Init(os.Getenv("REDIS_URI"), os.Getenv("REDIS_PASSWORD"), 0)

	// mux.HandleFunc("/ws", transport.WsHandler)
	mux.HandleFunc("GET /rooms/new/{duration}", security.WithIssuer(issuer)(transport.HandleCreateRoom))
	mux.HandleFunc("POST /rooms/{room_id}/auth", security.WithIssuer(issuer)(transport.HandleAuth))

	// secured endpoints
	mux.HandleFunc("POST /rooms/{room_id}/start", security.RequireAuth(issuer)(transport.HandleStartRoom))
	mux.HandleFunc("POST /rooms/{room_id}/join", security.RequireAuth(issuer)(transport.HandleJoinRoom))
	mux.HandleFunc("POST /rooms/{room_id}/lobby", security.RequireAuth(issuer)(transport.HandleLobby))

	port := os.Getenv("SIGNALING_PORT")
	server := &http.Server{
		Addr:    port,
		Handler: logger.SlogMiddleware(mux), // Slog handle server logging
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}

	log.Println("Server starting at port " + port)
	server.ListenAndServeTLS(cert, key)

}
