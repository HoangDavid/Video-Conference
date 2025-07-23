package domain

import (
	"context"
	"sync"
)

type Detector interface {
	Sample(id string, lvl int)
	Remove(id string)
	ActiveSpeaker() <-chan string
}

type DetectorObj struct {
	Mu      sync.Mutex
	Sum     map[string]int
	Count   map[string]int
	Current string
	Winner  chan string
	Margin  int
	Cancel  context.CancelFunc
}
