package main

import (
	"log"
	"vidcall/pkg/signaling"
)

func main() {
	client, err := signaling.NewClient("ws://<host>:8080/ws")

	if err != nil {
		log.Fatalf("failed to connect to Signaling API: %v", err)
	}

	defer client.Close()

}
