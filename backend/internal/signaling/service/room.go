package service

import (
	"context"
	"time"

	"vidcall/internal/signaling/domain"
	"vidcall/internal/signaling/repo"
	"vidcall/pkg/utils"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/mongo"
)

type Room struct {
	*domain.Room
}

func NewMember(name string, role string, conn *websocket.Conn) *domain.Member {
	return &domain.Member{
		Name:     name,
		PeerID:   utils.GenerateMemeberID(),
		Conn:     conn,
		Role:     role,
		JoinedAt: time.Now().UTC(),
	}
}

func NewRoom(ctx context.Context, db *mongo.Database, duration time.Duration) *domain.Room {

	pin := utils.GeneratePin()
	roomID := utils.GenerateRoomID()

	room := domain.Room{
		RoomID:   roomID,
		HostID:   "",
		Members:  make(map[string]domain.Member),
		Pin:      pin,
		Date:     time.Now().UTC(),
		Duration: duration,
	}

	// TODO: pinHashed :=

	repo.CreateRoomDoc(ctx, db, room)

	return &room
}

func (r *Room) Join(member domain.Member) {
	if member.Role == "host" {
		r.HostID = member.PeerID
		r.Members[member.PeerID] = member

	} else {
		r.Members[member.PeerID] = member
	}

	// TODO: broadcast to notify join action

}

func (r *Room) Leave(member domain.Member) {
	delete(r.Members, member.PeerID)

	// TODO: broadcast to leave action
	// TODO: if host leave

}
