package transport

import (
	"net/http"
	"time"

	"vidcall/internal/signaling/service"
	"vidcall/pkg/logger"
	"vidcall/pkg/utils"
)

// /start_room/{duration}
func HandleCreateRoom(w http.ResponseWriter, r *http.Request) {

	type Resp struct {
		RoomID    string `json:"roomID"`
		Pin       string `json:"pin"`
		HostToken string `json:"hostToken"`
	}

	ctx := r.Context()
	log := logger.GetLog(ctx).With("layer", "transport")

	duration, err := time.ParseDuration(r.PathValue("duration"))
	if err != nil {
		log.Warn("Unable to parse meeting duration")
	}

	room, host_token := service.NewRoom(ctx, duration)

	utils.Respond(w, http.StatusCreated,
		Resp{
			RoomID:    room.RoomID,
			Pin:       room.Pin,
			HostToken: host_token,
		})
}

// /rooms/{room_id}/auth
func HandleAuth(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Pin string `json:"pin"`
	}

	ctx := r.Context()
	log := logger.GetLog(ctx).With("layer", "transport")

	roomID := r.PathValue("room_id")
	err := utils.Decode(r, &req)
	if err != nil {
		log.Error("unable to decode request payload")
	}

	service.Auth(ctx, roomID, req.Pin)

}

// /rooms/{room_id}/start
func HandleStartRoom(w http.ResponseWriter, r *http.Request) {

}

// /rooms/{room_id}/join
func HandleJoinRoom(w http.ResponseWriter, r *http.Request) {

}

func HandleLobby(w http.ResponseWriter, r *http.Request) {

}
