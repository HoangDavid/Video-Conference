package service

import (
	"time"

	"vidcall/internal/signaling/domain"
	"vidcall/pkg/utils"

	"github.com/gorilla/websocket"
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

func NewRoom(duration time.Duration) *domain.Room {
	return &domain.Room{
		RoomID:   utils.GenerateRoomID(),
		HostID:   "",
		Members:  make(map[string]domain.Member),
		Pin:      utils.GeneratePin(),
		Date:     time.Now().UTC(),
		Duration: duration,
	}
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
