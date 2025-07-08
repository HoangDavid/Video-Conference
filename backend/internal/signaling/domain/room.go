package domain

import (
	"errors"
	"time"
)

type Room struct {
	RoomID   string
	HostID   string
	Pin      string
	Date     time.Time
	Duration time.Duration
}

var (
	ErrNotFound  = errors.New("room not found")
	ErrBadPin    = errors.New("invalid pin")
	ErrForbidden = errors.New("not permitted")
)
