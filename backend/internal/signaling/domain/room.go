package domain

import (
	"time"

	"github.com/gorilla/websocket"
)

type Room struct {
	RoomID   string
	HostID   string
	Members  map[string]Member
	Pin      string
	Date     time.Time
	Duration time.Duration
}

type Member struct {
	Name     string
	PeerID   string
	Conn     *websocket.Conn
	Role     string // Host / Guest
	JoinedAt time.Time
}
