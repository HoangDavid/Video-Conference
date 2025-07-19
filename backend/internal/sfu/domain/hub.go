package domain

import "sync"

type Hub struct {
	Mu    sync.RWMutex
	Turn  string
	Stuns []string
	Rooms map[string]*Room
}
