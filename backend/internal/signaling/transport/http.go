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

	type resp struct {
		RoomID string `json:"roomID"`
		Pin    string `json:"pin"`
		Token  string `json:"token"`
	}

	ctx := r.Context()
	log := logger.GetLog(ctx).With("layer", "transport")

	duration, err := time.ParseDuration(r.PathValue("duration"))
	if err != nil {
		log.Warn("Unable to parse meeting duration")
		utils.Error(w, http.StatusBadRequest, "invalid payload format")
		return
	}

	room, host_token, err := service.NewRoom(ctx, duration)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "internal error")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    host_token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	utils.Respond(w, http.StatusCreated,
		&resp{
			RoomID: room.RoomID,
			Pin:    room.Pin,
			Token:  host_token,
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
		utils.Error(w, http.StatusBadRequest, "invalid payload format")
	}

	token, err := service.Auth(ctx, roomID, req.Pin)
	switch err {
	case nil:

		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    token,
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
		})

		utils.Respond(w, http.StatusOK, map[string]string{"token": token})
	case service.ErrBadPin:
		utils.Error(w, http.StatusUnauthorized, "unathorized")
	case service.ErrNotFound:
		utils.Error(w, http.StatusNotFound, "room not found")
	default:
		utils.Error(w, http.StatusInternalServerError, "internal error")
	}

}

// /rooms/{room_id}/start
func HandleStartRoom(w http.ResponseWriter, r *http.Request) {

}

// /rooms/{room_id}/join
func HandleJoinRoom(w http.ResponseWriter, r *http.Request) {

}

func HandleLobby(w http.ResponseWriter, r *http.Request) {

}
