package transport

import "vidcall/internal/signaling/domain"

type RoomDTO struct {
	RoomID string
	Pin    string
}

func RoomCreatedResponse(r *domain.Room) RoomDTO {
	return RoomDTO{
		RoomID: r.RoomID,
		Pin:    r.Pin,
	}
}
