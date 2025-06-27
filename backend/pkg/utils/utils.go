package utils

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"vidcall/pkg/logger"

	gonanoid "github.com/matoous/go-nanoid"
	"golang.org/x/crypto/bcrypt"
)

const (
	pinDigits  = 6
	bcryptCost = bcrypt.DefaultCost
)

// Generate Room Pin
func GeneratePin(ctx context.Context) string {
	log := logger.GetLog(ctx).With("layer", "utils")

	max := big.NewInt(1)
	max.Exp(big.NewInt(10), big.NewInt(pinDigits), nil)

	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		log.Error("Unable to generate pin")
		// TODO: context cancel
	}

	return fmt.Sprintf("%0*d", pinDigits, n.Int64())
}

func PinHash(ctx context.Context, pin string) string {
	log := logger.GetLog(ctx).With("layer", "utils")

	hash, err := bcrypt.GenerateFromPassword([]byte(pin), bcryptCost)
	if err != nil {
		log.Error("Unable to hash pin")
		// TODO: context cancel
	}

	return string(hash)
}

func VerifyPin(pin string, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pin)) == nil
}

func GenerateRoomID(ctx context.Context) string {
	log := logger.GetLog(ctx).With("layer", "utils")

	id, err := gonanoid.Generate(
		"0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz",
		16,
	)

	if err != nil {
		log.Error("Unable to Generate RoomID")
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
		log.Error("Unable to Generate MemberID")
		// TODO: context cancel
	}

	return id
}
