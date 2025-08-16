package transport

import (
	"fmt"
	sfu "vidcall/api/proto"
	"vidcall/internal/sfu/service"
)

type Server struct {
	sfu.UnimplementedSFUServer
}

func (s *Server) Signal(stream sfu.SFU_SignalServer) error {

	ctx := stream.Context()

	fmt.Println("hi")
	// Temoporary: max 4 people in a meeting, for demo
	newPeer, err := service.NewPeer(ctx, stream, 3)
	if err != nil {
		return nil
	}

	// auto cut peer connections by manual/error
	defer newPeer.Disconnect()

	newPeer.Connect()

	return nil

}
