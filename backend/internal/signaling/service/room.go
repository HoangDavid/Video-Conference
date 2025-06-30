package service

import (
	"context"
	"time"

	"vidcall/internal/signaling/domain"
	"vidcall/internal/signaling/infra/mongo"
	"vidcall/internal/signaling/repo/mongorepo"
	"vidcall/internal/signaling/security"
	"vidcall/pkg/logger"
	"vidcall/pkg/utils"
)

type Room struct {
	*domain.Room
}

func NewRoom(ctx context.Context, duration time.Duration) *domain.Room {

	pin := security.GeneratePin(ctx)
	roomID := utils.GenerateRoomID(ctx)
	host_token := utils.GenerateHostToken(ctx)

	room := domain.Room{
		RoomID:   roomID,
		HostID:   host_token,
		Members:  make(map[string]domain.Member),
		Date:     time.Now().UTC(),
		Duration: duration,
	}

	log := logger.GetLog(ctx).With("layer", "service", "roomID", roomID)

	// Save room data
	hash := security.PinHash(ctx, pin)
	room.Pin = hash

	db := mongo.DB()
	mongorepo.CreateRoomDoc(ctx, db, room)
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
