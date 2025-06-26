package transport

import (
	"encoding/json"
	"net/http"
	"time"

	"vidcall/internal/signaling/service"
)

// /start_room/{duration}
func HandleStartRoom(w http.ResponseWriter, r *http.Request) {

	duration, _ := time.ParseDuration(r.PathValue("duration") + "m")

	room := service.NewRoom(duration)
	// TODO: database store room meta data

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)

	payload := RoomCreatedResponse(room)

	// TODO: add error log
	_ = json.NewEncoder(w).Encode(payload)

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
