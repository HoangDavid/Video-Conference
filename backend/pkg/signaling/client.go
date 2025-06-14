package signaling

import (
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

// pkg for clients listening in for signaling api

// Temporary
type Client struct {
	Conn *websocket.Conn
}

func NewClient(wsURL string) (*Client, error) {
	u := url.URL{Scheme: "ws", Host: wsURL, Path: "/ws"}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)

	if err != nil {
		return nil, fmt.Errorf("websocket error: %w", err)
	}

	return &Client{Conn: conn}, nil
}

// close ws connection with signaling
func (c *Client) Close() error {
	return c.Conn.Close()
}

// print out oncoming msg from signaling
func (c *Client) Listen(onMessage func(msg []byte)) {
	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			return
		}

		onMessage(msg)
	}
}

func (c *Client) SendEvery(payload []byte, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for range ticker.C {
			if err := c.Conn.WriteMessage(websocket.TextMessage, payload); err != nil {
				log.Println("periodic send error:", err)
				return
			}
		}
	}()
}
