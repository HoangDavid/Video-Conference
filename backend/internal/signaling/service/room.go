package service

import (
	"context"
	"time"

	"vidcall/internal/signaling/domain"
	"vidcall/internal/signaling/infra/mongo"
	"vidcall/internal/signaling/repo/mongo_repo"
	"vidcall/pkg/logger"
	"vidcall/pkg/utils"
)

type Room struct {
	*domain.Room
}

func NewRoom(ctx context.Context, duration time.Duration) *domain.Room {

	pin := utils.GeneratePin(ctx)
	roomID := utils.GenerateRoomID(ctx)

	room := domain.Room{
		RoomID:   roomID,
		HostID:   "",
		Members:  make(map[string]domain.Member),
		Date:     time.Now().UTC(),
		Duration: duration,
	}

	log := logger.GetLog(ctx).With("layer", "service", "roomID", roomID)

	// Save room data
	hash := utils.PinHash(ctx, pin)
	room.Pin = hash

	db := mongo.DB()
	mongo_repo.CreateRoomDoc(ctx, db, room)
	log.Info("new room created")

	room.Pin = pin

	return &room
}

// func NewMember(name string, role string ) *domain.Member

func JoinRoom(ctx context.Context, name string, room_id string) {

	// TODO: broadcast to notify join action

}

func (r *Room) Leave(member domain.Member) {
	// TODO: broadcast to leave action
	// TODO: if host leave

}
