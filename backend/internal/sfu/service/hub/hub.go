package hub

import (
	"sync"
	"vidcall/internal/sfu/domain"
)

type HubObj struct {
	*domain.HubObj
}

var (
	once sync.Once
	hub  *domain.HubObj
)

func Init() {
	once.Do(func() {
		hub = &domain.HubObj{
			Stuns: []string{"stun:stun.l.google.com:19302"},
			Rooms: make(map[string]domain.Room),
		}
	})
}

func Hub() domain.Hub {
	return &HubObj{
		HubObj: hub,
	}
}

func (h *HubObj) GetStuns() []string {
	return h.Stuns
}

func (h *HubObj) GetTurn() string {
	return h.Turn
}

func (h *HubObj) AddRoom(roomID string, room domain.Room) {
	h.Mu.Lock()
	defer h.Mu.Unlock()
	h.Rooms[roomID] = room
}

func (h *HubObj) RemoveRoom(roomID string) domain.Room {
	h.Mu.Lock()
	defer h.Mu.Unlock()

	v, ok := h.Rooms[roomID]
	if !ok {
		return nil
	}

	delete(h.Rooms, roomID)
	return v
}

func (h *HubObj) GetRoom(roomID string) domain.Room {
	h.Mu.Lock()
	defer h.Mu.Unlock()

	v, ok := h.Rooms[roomID]
	if !ok {
		return nil
	}

	return v
}
