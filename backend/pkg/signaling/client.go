package signaling

import (
	"fmt"
	"net/url"

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
