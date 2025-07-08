package service

import (
	"context"
	"time"

	"vidcall/internal/signaling/domain"
	"vidcall/internal/signaling/infra/mongox"
	"vidcall/internal/signaling/infra/redisx"
	"vidcall/internal/signaling/repo/mongorepo"
	"vidcall/internal/signaling/repo/redisrepo"
	"vidcall/internal/signaling/security"
	"vidcall/pkg/logger"
	"vidcall/pkg/utils"
)

func NewRoom(ctx context.Context, duration time.Duration, name string) (*domain.Room, string, error) {

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
	host_token, err := issuer.Issue(roomID, hostID, name, "host")
	if err != nil {
		log.Error("unable to tokenize")
		return nil, "", err
	}

	log = log.With("roomID", roomID)
	log.Info("created new room")
	room.Pin = pin

	return &room, host_token, nil
}

func StartRoom(ctx context.Context) error {

	log := logger.GetLog(ctx).With("layer", "service")

	// Check for member role
	claims := security.ClaimsFrom(ctx)
	role := claims.Role
	if role != "host" {
		return domain.ErrForbidden
	}

	// Create room and add member
	c := redisx.C()
	roomID := claims.RoomID
	memeberID := claims.Subject

	if err := redisrepo.CreateRoom(ctx, c, roomID); err != nil {
		return err
	}

	if err := redisrepo.AddMember(ctx, c, roomID, memeberID); err != nil {
		return err
	}

	log = log.With("roomID", roomID)
	log.Info("room is live")

	return nil
}

func JoinRoom(ctx context.Context) error {
	// TODO: broadcast to notify join action
	log := logger.GetLog(ctx).With("layer", "service")

	claims := security.ClaimsFrom(ctx)
	roomID := claims.RoomID
	memberID := claims.Subject
	c := redisx.C()

	if err := redisrepo.AddMember(ctx, c, roomID, memberID); err != nil {
		return err
	}

	log = log.With("roomID", roomID)
	log.Info("a member joined")

	return nil
}

func LeaveRoom(ctx context.Context) error {
	return nil
}

func Lobby(ctx context.Context, roomID string) {

}
