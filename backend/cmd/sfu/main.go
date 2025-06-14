package main

import (
	"log"
	"time"
	"vidcall/pkg/signaling"
)

func main() {
	client, err := signaling.NewClient("localhost:8080")

	if err != nil {
		log.Fatalf("failed to connect to Signaling API: %v", err)
	}

	defer client.Close()

	client.SendEvery([]byte("hi from sfu"), 2*time.Second)

	client.Listen(func(msg []byte) {
		log.Printf("got reply: %s", msg)
	})
}
