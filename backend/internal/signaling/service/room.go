package service

import (
	"context"
	"time"

	"vidcall/internal/signaling/domain"
	"vidcall/pkg/logger"
	"vidcall/pkg/utils"
)

type Room struct {
	*domain.Room
}

func NewRoom(ctx context.Context, duration time.Duration) *domain.Room {

	pin := utils.GeneratePin()
	roomID := utils.GenerateRoomID()

	room := domain.Room{
		RoomID:   roomID,
		HostID:   "",
		Members:  make(map[string]domain.Member),
		Pin:      pin,
		Date:     time.Now().UTC(),
		Duration: duration,
	}

	log := logger.GetLog(ctx).With("layer", "service", "roomID", roomID)

	// TODO: pinHashed :=
	log.Info("new room created")

	return &room
}

// func NewMember(name string, role string, conn *websocket.Conn) *domain.Member

func (r *Room) Join(member domain.Member) {

	// TODO: broadcast to notify join action

}

func (r *Room) Leave(member domain.Member) {
	// TODO: broadcast to leave action
	// TODO: if host leave

}
