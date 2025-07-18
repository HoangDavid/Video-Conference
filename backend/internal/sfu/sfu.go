package sfu

import (
	"log"
	"net"
	"os"
	sfu "vidcall/api/proto"
	"vidcall/internal/sfu/service/hub"
	"vidcall/internal/sfu/transport"

	_ "github.com/joho/godotenv/autoload"
	"google.golang.org/grpc"
)

func Execute() {
	// TODO: add TLS for security + certs

	// initialize a hub
	hub.Init()

	port := os.Getenv("SFU_PORT")
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed tp listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	sfu.RegisterSFUServer(grpcServer, &transport.Server{})

	log.Println("SFU server starting at port " + port)
	if err = grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to server: %v", err)
	}

}
