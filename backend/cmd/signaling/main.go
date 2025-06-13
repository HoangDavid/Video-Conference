package main

import (
	"fmt"
	"net/http"
	"vidcall/internal/signaling"
)

func main() {
	http.HandleFunc("/ws", signaling.WsHandler)
	fmt.Println("Websocket server started on :8080")
	err := http.ListenAndServe(":8080", nil)

	if err != nil {
		fmt.Println("Error starting server")
	}
}
