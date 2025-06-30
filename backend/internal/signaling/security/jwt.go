package security

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	RoomID string
	Role   string
	jwt.RegisteredClaims
}

type Issuer struct {
	secret []byte
	ttl    time.Duration
}

func NewIssuer(secret string) *Issuer {
	//  TODO: sync with meeting duration somehow???
	exp := 2 * time.Hour
	return &Issuer{secret: []byte(secret), ttl: exp}
}

func (i *Issuer) Issue(roomID string, peerID string, role string) (string, error) {
	now := time.Now()
	c := Claims{
		RoomID: roomID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   peerID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(i.ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)

	return token.SignedString(i.secret)
}

func (i *Issuer) Parse(raw string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		raw,
		&Claims{},
		func(_ *jwt.Token) (any, error) { return i.secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)

	if err != nil || !token.Valid {
		return nil, err
	}

	return token.Claims.(*Claims), nil
}
