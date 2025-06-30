package utils

import (
	"crypto/rand"
	"encoding/hex"

	gonanoid "github.com/matoous/go-nanoid"
)

func GenerateRoomID() string {
	id, err := gonanoid.Generate(
		"0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz",
		16,
	)

	if err != nil {
		return ""
	}

	return id
}

func GenerateMemeberID() string {

	id, err := gonanoid.Generate(
		"0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz",
		12,
	)

	if err != nil {
		return ""
	}

	return id
}

func GenerateHostID() string {

	var b [32]byte
	_, err := rand.Read(b[:])

	if err != nil {
		return ""
	}

	return hex.EncodeToString(b[:])
}
