package room

import (
	"context"
	sfu "vidcall/api/proto"
	"vidcall/internal/sfu/domain"
	"vidcall/internal/sfu/service/hub"
)

type RoomObj struct {
	*domain.RoomObj
}

func NewRoom(roomID string) domain.Room {

	rCtx, rCancel := context.WithCancel(context.Background())

	room := &RoomObj{
		RoomObj: &domain.RoomObj{
			ID:       roomID,
			Live:     false,
			Peers:    make(map[string]domain.Peer),
			Ctx:      rCtx,
			Cancel:   rCancel,
			JoinChan: make(chan domain.Peer, 64),
		},
	}

	// Add new room to hub
	hub.Hub().AddRoom(roomID, room)

	return room

}

func (r *RoomObj) Close() {
	r.Cancel()
	close(r.JoinChan)
	hub.Hub().RemoveRoom(r.ID)
}

func (r *RoomObj) MakeLive() {
	r.Mu.Lock()
	defer r.Mu.Unlock()
	r.Live = true
}

func (r *RoomObj) IsLive() bool {
	r.Mu.RLock()
	defer r.Mu.RUnlock()
	return r.Live
}

func (r *RoomObj) AddPeer(peerID string, peer domain.Peer) {
	r.Mu.Lock()
	defer r.Mu.Unlock()
	r.Peers[peerID] = peer

	// trigger new peer join to subcriber audio
	r.JoinChan <- peer
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
	r.Mu.RLock()
	defer r.Mu.RUnlock()

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

		select {
		case <-r.Ctx.Done():
			return
		default:
			peer.EnqueueEvent(event)
		}
	}
}

func (r *RoomObj) ListPeers() map[string]domain.Peer {
	r.Mu.RLock()
	defer r.Mu.RUnlock()

	peers := r.Peers

	return peers
}
