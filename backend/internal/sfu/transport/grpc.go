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
	peer := service.NewPeer(ctx, stream, nil)

	peer.Negotiate()

	return nil

}
