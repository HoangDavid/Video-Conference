package service

import (
	"context"
	"errors"
	"vidcall/internal/signaling/infra/mongox"
	"vidcall/internal/signaling/repo/mongorepo"
	"vidcall/internal/signaling/security"
	"vidcall/pkg/logger"
	"vidcall/pkg/utils"

	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrNotFound = errors.New("room not found")
	ErrBadPin   = errors.New("invalid pin")
)

func Auth(ctx context.Context, roomID string, pin string) (string, error) {
	// TODO: cases handling
	log := logger.GetLog(ctx).With("layer", "service")

	memberID := utils.GenerateMemeberID()

	// Find valid room doc
	db := mongox.DB()
	room, err := mongorepo.GetRoomDoc(ctx, db, roomID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", ErrNotFound
		}
		return "", err
	}

	// Verify Pin
	if ok := security.VerifyPin(pin, room.Pin); !ok {
		return "", ErrBadPin
	}

	// JWT Token
	issuer := security.IssuerFrom(ctx)
	member_token, err := issuer.Issue(roomID, memberID, "guest")
	if err != nil {
		log.Error("unable to tokenize")
		return "", err
	}

	return member_token, nil

}
