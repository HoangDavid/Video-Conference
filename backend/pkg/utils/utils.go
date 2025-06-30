package utils

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"vidcall/pkg/logger"

	gonanoid "github.com/matoous/go-nanoid"
)

func GenerateRoomID(ctx context.Context) string {
	log := logger.GetLog(ctx).With("layer", "utils")

	id, err := gonanoid.Generate(
		"0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz",
		16,
	)

	if err != nil {
		log.Error("unable to generate roomid")
		// TODO: context cancel
	}

	return id
}

func GenerateMemeberID(ctx context.Context) string {
	log := logger.GetLog(ctx).With("layer", "utils")

	id, err := gonanoid.Generate(
		"0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz",
		12,
	)

	if err != nil {
		log.Error("unable to generate memberid")
		// TODO: context cancel
	}

	return id
}

func GenerateHostToken(ctx context.Context) string {
	log := logger.GetLog(ctx).With("layer", "utils")

	var b [32]byte
	_, err := rand.Read(b[:])
	if err != nil {
		log.Error("unable to generate host token")
	}

	return hex.EncodeToString(b[:])
}
