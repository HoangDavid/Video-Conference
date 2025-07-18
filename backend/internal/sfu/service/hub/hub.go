package hub

import (
	"sync"
	"vidcall/internal/sfu/domain"
)

var (
	once sync.Once
	hub  domain.Hub
)

type Hub struct {
	*domain.Hub
}

func Init() {
	once.Do(func() {

		//  TODO: load turn/turn configs from yaml files

		hub = domain.Hub{
			Stuns: []string{"stun:stun.l.google.com:19302"},
			Rooms: make(map[string]*domain.Room),
		}
	})
}

func GetHub() *Hub {
	return &Hub{
		Hub: &hub,
	}
}

func (h *Hub) AddRoom(room *domain.Room) {
	hub.Mu.Lock()
	defer hub.Mu.Unlock()

	h.Rooms[room.ID] = room
}

func (h *Hub) RemoveRoom(RoomID string) {
	hub.Mu.Lock()
	defer hub.Mu.Unlock()

	delete(h.Rooms, RoomID)

}

func (h *Hub) GetRoom(roomID string) *domain.Room {
	hub.Mu.RLock()
	defer hub.Mu.RUnlock()

	room, ok := h.Rooms[roomID]

	if !ok {
		return nil
	}

	return room
}
