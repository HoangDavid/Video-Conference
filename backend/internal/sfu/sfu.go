package sfu

import (
	"context"
	"log"
	"net"
	pb "vidcall/api/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

type server struct {
	pb.UnimplementedSFUServer
}

func (s *server) Ping(ctx context.Context, in *pb.PingReq) (*pb.Pong, error) {
	log.Println("SFU got: ", in.Msg)

	return &pb.Pong{Msg: "pong from SFU"}, nil
}

func Execute() {
	// TODO: add TLS for security + certs
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	grpcServer := grpc.NewServer(grpc.Creds(insecure.NewCredentials()))
	pb.RegisterSFUServer(grpcServer, &server{})

	reflection.Register(grpcServer)

	log.Println("SFU is listening on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
