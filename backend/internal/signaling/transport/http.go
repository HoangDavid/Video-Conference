package transport

import (
	"encoding/json"
	"net/http"
	"time"

	"vidcall/internal/signaling/service"
	"vidcall/pkg/logger"
)

// /start_room/{duration}
func HandleStartRoom(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	log := logger.GetLog(ctx).With("layer", "transport")
	duration, err := time.ParseDuration(r.PathValue("duration"))

	if err != nil {
		log.Warn("Unable to parse meeting duration")
	}

	room := service.NewRoom(ctx, duration)
	// TODO: database store room meta data

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)

	payload := RoomCreatedResponse(room)
	err = json.NewEncoder(w).Encode(payload)

	if err != nil {
		log.Error("Unable to send payload")
		return
	}
}

// /rooms/{room_id}/join?pin=######
func HandleJoinRoom(w http.ResponseWriter, r *http.Request) {
	r.PathValue("room_id")
	r.URL.Query().Get("pin")

	// TODO: search for roomID in DB and check pin

}

// room_id/leave_room
func HandleLeaveRoom(w http.ResponseWriter, r *http.Request) {

}
