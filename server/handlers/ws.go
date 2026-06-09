package handlers

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Morolis/cb/server/ws"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		allowedOrigin := os.Getenv("CB_CORS_ORIGIN")
		if allowedOrigin == "" || allowedOrigin == "*" {
			return true
		}
		origin := r.Header.Get("Origin")
		for _, o := range strings.Split(allowedOrigin, ",") {
			if strings.TrimSpace(o) == origin {
				return true
			}
		}
		return false
	},
}

type WSHandler struct {
	hub *ws.Hub
}

func NewWSHandler(hub *ws.Hub) *WSHandler {
	return &WSHandler{hub: hub}
}

func (h *WSHandler) HandleWS(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("websocket upgrade error: %v", err)
		return
	}

	client := &ws.Client{
		Hub:    h.hub,
		UserID: userID,
		Send:   make(chan []byte, 256),
	}

	h.hub.Register(client)

	// Write pump
	go func() {
		defer conn.Close()
		for {
			msg, ok := <-client.Send
			if !ok {
				conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		}
	}()

	// Read pump (keep connection alive, handle close)
	go func() {
		defer func() {
			h.hub.Unregister(client)
			conn.Close()
		}()
		conn.SetReadLimit(512)
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		conn.SetPongHandler(func(string) error {
			conn.SetReadDeadline(time.Now().Add(60 * time.Second))
			return nil
		})
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}()

	// Send initial ping
	client.Send <- []byte(`{"type":"connected","payload":{}}`)
}
