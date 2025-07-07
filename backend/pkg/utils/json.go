package utils

import (
	"encoding/json"
	"net/http"
)

func Decode(r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(nil, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(dst)
	if err != nil {
		return err
	}

	return nil
}

func Respond(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")

	if body == nil {
		w.WriteHeader(status)
		return
	}

	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func Error(w http.ResponseWriter, status int, msg string) {
	Respond(w, status, map[string]string{"error": msg})
}

func Cookie(w http.ResponseWriter, token string, path string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    token,
		Path:     path,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}
