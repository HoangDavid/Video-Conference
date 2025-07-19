package transport

import (
	sfu "vidcall/api/proto"
)

type Server struct {
	sfu.UnimplementedSFUServer
}

func NewServer() {
	//  TODO: add TURN server
	//  TODO: take all of this into a yaml file ?

}

func (s *Server) Signal(stream sfu.SFU_SignalServer) error {

	return nil

}
