package room

import (
	sfu "vidcall/api/proto"
	"vidcall/internal/sfu/domain"
	"vidcall/internal/sfu/service/hub"
)

type RoomObj struct {
	*domain.RoomObj
}

func NewRoom(roomID string) domain.Room {
	r := &domain.RoomObj{
		ID:    roomID,
		Peers: make(map[string]domain.Peer),
	}

	room := &RoomObj{
		RoomObj: r,
	}

	hub.Hub().AddRoom(roomID, room)
	return room

}

func (r *RoomObj) AddPeer(peerID string, peer domain.Peer) {
	r.Mu.Lock()
	defer r.Mu.Unlock()

	r.Peers[peerID] = peer
}

func (r *RoomObj) RemovePeer(peerID string) domain.Peer {
	r.Mu.Lock()
	defer r.Mu.Unlock()

	v, ok := r.Peers[peerID]

	if !ok {
		return nil
	}

	delete(r.Peers, peerID)
	return v
}

func (r *RoomObj) GetPeer(peerID string) domain.Peer {
	r.Mu.Lock()
	defer r.Mu.Unlock()

	v, ok := r.Peers[peerID]

	if !ok {
		return nil
	}

	return v
}

func (r *RoomObj) BroadCast(peerID string, event *sfu.PeerSignal_Event) {
	r.Mu.Lock()
	defer r.Mu.Unlock()

	for id, peer := range r.Peers {
		if peerID == id {
			continue
		}

		peer.EnqueueEvent(event)
	}
}

func (r *RoomObj) ListPeers() map[string]domain.Peer {
	r.Mu.Lock()
	defer r.Mu.Unlock()

	peers := r.Peers

	return peers
}
