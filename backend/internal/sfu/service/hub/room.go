package hub

import (
	"vidcall/internal/sfu/domain"
)

type Room struct {
	*domain.Room
}

func NewRoom(RoomID string) *Room {
	room := domain.Room{
		ID:    RoomID,
		Peers: make(map[string]*domain.Peer),
	}

	hub := GetHub()
	hub.AddRoom(&room)

	return &Room{
		Room: &room,
	}
}

func (r *Room) AddPeer(peerID string, peer *domain.Peer) {
	r.Mu.Lock()
	defer r.Mu.Unlock()
	r.Peers[peerID] = peer
}

func (r *Room) RemovePeer(peerID string) {
	r.Mu.Lock()
	defer r.Mu.Lock()

	delete(r.Peers, peerID)
}

func (r *Room) GetPeer(peerID string) *domain.Peer {
	r.Mu.RLock()
	defer r.Mu.RUnlock()

	peer, ok := r.Peers[peerID]

	if !ok {
		return nil
	}

	return peer
}
