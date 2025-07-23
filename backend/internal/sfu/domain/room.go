package domain

import (
	"context"
	"sync"
	sfu "vidcall/api/proto"
)

type Room interface {
	AddPeer(peerID string, peer Peer)
	RemovePeer(peerID string) Peer
	GetPeer(peerID string) Peer
	BroadCast(peerID string, event *sfu.PeerSignal_Event)
	ListPeers() map[string]Peer
	Close()
}

type RoomObj struct {
	Mu       sync.RWMutex
	ID       string
	Peers    map[string]Peer
	Detector Detector
	Ctx      context.Context
	Cancel   context.CancelFunc
}
