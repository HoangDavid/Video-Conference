package domain

import "sync"

type Hub interface {
	GetStuns() []string
	GetTurn() string
	AddRoom(roomID string, room Room)
	RemoveRoom(roomID string) Room
	GetRoom(roomID string) Room
}

type HubObj struct {
	Mu    sync.RWMutex
	Turn  string
	Stuns []string
	Rooms map[string]Room
}
