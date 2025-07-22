package transport

import (
	sfu "vidcall/api/proto"
	"vidcall/internal/sfu/service"
)

type Server struct {
	sfu.UnimplementedSFUServer
}

func (s *Server) Signal(stream sfu.SFU_SignalServer) error {

	ctx := stream.Context()

	// Temoporary: max 5 people in a meeting
	newPeer, err := service.NewPeer(ctx, stream, 5)
	if err != nil {
		return nil
	}

	defer newPeer.Disconnect()

	newPeer.Connect()

	return nil

}
