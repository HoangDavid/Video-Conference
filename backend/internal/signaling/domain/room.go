package domain

import (
	"time"
)

type Room struct {
	RoomID   string
	HostID   string
	Pin      string
	Date     time.Time
	Duration time.Duration
}
