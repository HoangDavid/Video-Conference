package service

import (
	"context"
	"time"

	"vidcall/internal/signaling/domain"
	"vidcall/internal/signaling/infra/mongox"
	"vidcall/internal/signaling/repo/mongorepo"
	"vidcall/internal/signaling/security"
	"vidcall/pkg/logger"
	"vidcall/pkg/utils"
)

type Room struct {
	*domain.Room
}

func NewRoom(ctx context.Context, duration time.Duration) (*domain.Room, string, error) {

	log := logger.GetLog(ctx).With("layer", "service")

	pin := security.GeneratePin(ctx)
	roomID := utils.GenerateRoomID()
	hostID := utils.GenerateHostID()

	room := domain.Room{
		RoomID:   roomID,
		HostID:   hostID,
		Pin:      security.PinHash(ctx, pin),
		Date:     time.Now().UTC(),
		Duration: duration,
	}

	// Save room data
	db := mongox.DB()
	if err := mongorepo.CreateRoomDoc(ctx, db, room); err != nil {
		return nil, "", err
	}

	// Tokenize
	issuer := security.IssuerFrom(ctx)
	host_token, err := issuer.Issue(roomID, hostID, "host")
	if err != nil {
		log.Error("unable to tokenize")
		return nil, "", err
	}

	log = log.With("roomID", roomID)
	log.Info("created new room")
	room.Pin = pin

	return &room, host_token, nil
}

func StartRoom(ctx context.Context, roomID string) {

}

func JoinRoom(ctx context.Context, name string, room_id string) {

	// TODO: broadcast to notify join action

}
