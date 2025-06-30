package service

import (
	"context"
	"vidcall/internal/signaling/infra/mongo"
	"vidcall/internal/signaling/repo/mongorepo"
	"vidcall/internal/signaling/security"
	"vidcall/pkg/logger"
	"vidcall/pkg/utils"
)

func Auth(ctx context.Context, roomID string, pin string) string {
	// TODO: cases handling
	log := logger.GetLog(ctx).With("layer", "service")

	memberID := utils.GenerateMemeberID()

	db := mongo.DB()
	room := mongorepo.GetRoomDoc(ctx, db, roomID)
	if room == nil {
		return ""
	}

	if !security.VerifyPin(pin, room.Pin) {
		//  TODO: Invalid Pin
		return ""
	}

	issuer := security.IssuerFrom(ctx)
	member_token, err := issuer.Issue(roomID, memberID, "guest")
	if err != nil {
		log.Error("unable to tokenize")
		return ""
	}

	return member_token

}
