package domain

import "sync"

type Hub struct {
	Mu    sync.RWMutex
	Stuns []string
	Turn  string // TODO: add this later
	Rooms map[string]*Room
}
