package transport

import "vidcall/internal/signaling/domain"

type RoomDTO struct {
	RoomID string
	HostID string
	Pin    string
}

func RoomCreatedResponse(r *domain.Room) RoomDTO {
	return RoomDTO{
		RoomID: r.RoomID,
		HostID: r.HostID,
		Pin:    r.Pin,
	}
}
