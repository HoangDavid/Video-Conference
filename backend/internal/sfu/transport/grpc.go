package transport

import (
	sfu "vidcall/api/proto"
)

type Server struct {
	sfu.UnimplementedSFUServer
}

func NewServer() {
	//  TODO: add TURN server
	//  TODO: take all of this into a yaml file ??
	stuns := []string{
		"stun:stun.l.google.com:19302",
		"stun:stun1.l.google.com:19302",
		"stun:stun2.l.google.com:19302",
	}

}

func (s *Server) Signal(stream sfu.SFU_SignalServer) error {

	return nil

}
