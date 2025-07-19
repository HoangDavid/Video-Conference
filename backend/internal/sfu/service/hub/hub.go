package hub

import (
	"sync"
	"vidcall/internal/sfu/domain"
)

var (
	once sync.Once
	hub  *domain.Hub
)

func Init() {
	once.Do(func() {
		hub = &domain.Hub{
			Stuns: []string{"stun:stun.l.google.com:19302"},
			Rooms: make(map[string]*domain.Room),
		}
	})
}

func Hub() *domain.Hub {
	return hub
}

func GetRoom(roomID string) *domain.Room {
	hub.Mu.RLock()
	room := hub.Rooms[roomID]
	hub.Mu.RUnlock()

	return room
}
