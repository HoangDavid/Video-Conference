package hub

import (
	sfu "vidcall/api/proto"
	"vidcall/internal/sfu/domain"
)

func NewRoom(roomID string) {
	room := domain.Room{
		ID:    roomID,
		Peers: make(map[string]*domain.Peer),
	}

	h := Hub()

	h.Mu.Lock()
	h.Rooms[roomID] = &room
	h.Mu.Unlock()
}

func AddPeer(roomID string, peer *domain.Peer) {
	room := GetRoom(roomID)

	room.Mu.Lock()
	room.Peers[peer.ID] = peer
	room.Mu.Unlock()
}

func RemovePeer(roomID string, peer *domain.Peer) {
	room := GetRoom(roomID)

	room.Mu.Lock()
	delete(room.Peers, peer.ID)
	room.Mu.Unlock()
}

func BroadCast(ownerID string, roomID string, event *sfu.PeerSignal) {
	room := GetRoom(roomID)

	for id, peer := range room.Peers {
		if id == ownerID {
			continue
		}

		peer.EventQ <- event
	}
}
