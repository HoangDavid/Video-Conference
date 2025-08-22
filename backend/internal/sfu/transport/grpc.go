package transport

import (
	"fmt"
	sfu "vidcall/api/proto"
	"vidcall/internal/sfu/service"
	"vidcall/pkg/logger"
)

type Server struct {
	sfu.UnimplementedSFUServer
}

func (s *Server) Signal(stream sfu.SFU_SignalServer) error {

	ctx := stream.Context()

	// Temoporary: max 4 people in a meeting, for demo
	log := logger.GetLog(ctx)
	newPeer, err := service.NewPeer(ctx, stream, 1, log)
	if err != nil {
		return nil
	}
	log.Info("new peer created")

	// auto cut peer connections by manual/error
	defer newPeer.Disconnect()

	if err := newPeer.Connect(); err != nil {
		errMsg := fmt.Sprint("Peer unable to connect: %w", err)
		log.Error(errMsg)
	}

	log.Info("peer disconnected")

	return nil

}
