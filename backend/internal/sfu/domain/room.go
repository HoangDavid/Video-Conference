package domain

import (
	"sync"
)

type Room struct {
	Mu    sync.RWMutex
	ID    string
	Peers map[string]*Peer
	Live  bool
}
