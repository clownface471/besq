package handler

import (
	"log"
	"net/http"
	"pt-besq-core/internal/websocket"

	"github.com/gin-gonic/gin"
	gorilla "github.com/gorilla/websocket"
)

type WSHandler struct {
	Hub *websocket.Hub
}

func NewWSHandler(hub *websocket.Hub) *WSHandler {
	return &WSHandler{Hub: hub}
}

var upgrader = gorilla.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (h *WSHandler) HandleConnections(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Gagal upgrade WS:", err)
		return
	}

	// 1. Buat Client Wrapper
	client := &websocket.Client{
		Hub:  h.Hub,
		Conn: conn,
		Send: make(chan websocket.Message, 256),
	}

	// 2. Daftarkan ke Hub
	h.Hub.Register <- client

	// 3. Jalankan Goroutine untuk MENGIRIM pesan (Write Pump)
	go func() {
		defer func() {
			h.Hub.Unregister <- client
			conn.Close()
		}()
		for {
			msg, ok := <-client.Send
			if !ok {
				// Channel ditutup Hub
				conn.WriteMessage(gorilla.CloseMessage, []byte{})
				return
			}
			conn.WriteJSON(msg)
		}
	}()

	// 4. Jalankan Loop Utama untuk MEMBACA pesan (Read Pump / Keep Alive)
	// Walaupun kita tidak pakai pesan dari client, loop ini wajib ada biar koneksi gak putus
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			h.Hub.Unregister <- client
			break
		}
	}
}