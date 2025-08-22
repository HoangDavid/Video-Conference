package signaling

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"

	sfu "vidcall/api/proto"
	"vidcall/internal/signaling/infra"
	"vidcall/internal/signaling/security"
	"vidcall/internal/signaling/transport/httpx"
	"vidcall/internal/signaling/transport/wsx"
	"vidcall/pkg/logger"

	_ "github.com/joho/godotenv/autoload"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Execute() {

	issuer := security.NewIssuer(os.Getenv("JWT_SECRET"))

	mux := http.NewServeMux()

	// Fire up infra: MongoDB and Redis
	infra.Init(os.Getenv("MONGODB_URI"), os.Getenv("DB_NAME"), 10)

	// fire a gRPC connection between signaling and sfu
	sfu_host := os.Getenv("SFU_HOST")
	if sfu_host == "" {
		sfu_host = "localhost" + os.Getenv("SFU_PORT")
	}

	sfuConn, err := grpc.Dial(sfu_host, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
		return
	}

	sfuClient := sfu.NewSFUClient(sfuConn)

	// create new room and auth
	mux.HandleFunc("GET /api/rooms/new/{duration}", security.WithIssuer(issuer)(httpx.HandleCreateRoom))
	mux.HandleFunc("POST /api/rooms/{room_id}/auth", security.WithIssuer(issuer)(httpx.HandleAuth))

	// secured endpoints
	mux.HandleFunc("GET /api/me", security.RequireAuth(issuer)(func(w http.ResponseWriter, r *http.Request) {
		httpx.HandleClaims(w, r)
	}))
	mux.HandleFunc("GET /ws", security.RequireAuth(issuer)(func(w http.ResponseWriter, r *http.Request) {
		wsx.HandleWS(w, r, sfuClient)
	}))

	port := os.Getenv("SIGNALING_PORT")
	log.Println("Signaling server starting at port " + port)

	server := &http.Server{
		Addr:    port,
		Handler: logger.SlogMiddleware(mux), // Slog handle server logging
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}

	// Load TLS cert and key
	cert := os.Getenv("TLS_CERT")
	key := os.Getenv("TLS_KEY")

	if cert == "" || key == "" {
		log.Printf("TLS_CERT or TLS_KEY are not set. Serving HTTP...")
		log.Fatal(server.ListenAndServe())
	} else {
		log.Fatal(server.ListenAndServeTLS(cert, key))
	}

}
