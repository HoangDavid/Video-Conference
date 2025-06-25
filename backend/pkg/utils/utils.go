package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"

	gonanoid "github.com/matoous/go-nanoid"
	"golang.org/x/crypto/bcrypt"
)

// TODO: add this to config
const (
	pinDigits  = 6
	bcryptCost = bcrypt.DefaultCost
	saltBytes  = 16
)

// Generate Room Pin
func GeneratePin() string {
	max := big.NewInt(1)
	max.Exp(big.NewInt(10), big.NewInt(pinDigits), nil)

	// TODO: Error check this later
	n, _ := rand.Int(rand.Reader, max)

	return fmt.Sprintf("%0*d", pinDigits, n.Int64())
}

func GenerateRoomID() string {
	// TODO: Error check here
	id, _ := gonanoid.Generate(
		"0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz",
		16,
	)

	return id
}

func GenerateMemeberID() string {
	// TODO: error check here
	id, _ := gonanoid.Generate(
		"0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz",
		12,
	)

	return id
}
