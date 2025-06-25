package transport

import (
	"fmt"
	"net/http"
	"vidcall/internal/signaling/service"
)

func Init() {
	sig := service.NewSignaling("/ws")
	http.HandleFunc(sig.Addr, sig.WsHandler)
	fmt.Println("Websocket server started on :8080")
	err := http.ListenAndServe(":8080", nil)

	if err != nil {
		fmt.Println("Error starting server")
	}

}
