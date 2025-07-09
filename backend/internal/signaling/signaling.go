package signaling

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"

	sfu "vidcall/api/proto"
	"vidcall/internal/signaling/infra/mongox"
	"vidcall/internal/signaling/infra/redisx"
	"vidcall/internal/signaling/security"
	"vidcall/internal/signaling/transport"
	"vidcall/pkg/logger"

	_ "github.com/joho/godotenv/autoload"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Execute() {

	issuer := security.NewIssuer(os.Getenv("JWT_SECRET"))

	mux := http.NewServeMux()

	// Load TLS cert and key
	cert := os.Getenv("TLS_CERT")
	key := os.Getenv("TLS_KEY")
	if cert == "" || key == "" {
		log.Fatalf("TLS_CERT or TLS_KEY are not set")
		return
	}

	// Fire up infra: MongoDB and Redis
	mongox.Init(os.Getenv("MONGODB_URI"), os.Getenv("DB_NAME"), 10)
	redisx.Init(os.Getenv("REDIS_URI"), os.Getenv("REDIS_PASSWORD"), 0)

	// Fire a gRPC connection between signaling and sfu
	sfuConn, err := grpc.Dial("localhost"+os.Getenv("SFU_PORT"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
		return
	}
	sfuClient := sfu.NewSFUClient(sfuConn)

	// create new room and auth
	mux.HandleFunc("GET /rooms/new/{duration}", security.WithIssuer(issuer)(transport.HandleCreateRoom))
	mux.HandleFunc("POST /rooms/{room_id}/auth", security.WithIssuer(issuer)(transport.HandleAuth))

	// secured endpoints
	// mux.HandleFunc("GET /ws", security.RequireAuth(issuer)(func(w http.ResponseWriter, r *http.Request) {
	// 	transport.HandleWS(w, r, sfuClient) }))
	mux.HandleFunc("GET /ws", func(w http.ResponseWriter, r *http.Request) {
		transport.HandleWS(w, r, sfuClient)
	})

	port := os.Getenv("SIGNALING_PORT")
	server := &http.Server{
		Addr:    port,
		Handler: logger.SlogMiddleware(mux), // Slog handle server logging
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}

	log.Println("Signaling server starting at port " + port)
	log.Fatal(server.ListenAndServeTLS(cert, key))

}
