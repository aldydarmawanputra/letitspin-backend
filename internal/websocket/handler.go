package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	ID     string
	Conn   *websocket.Conn
	UserID string
	Send   chan []byte
}

type Hub struct {
	Clients    map[string]*Client
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
	Mu         sync.RWMutex
}

type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type GameUpdate struct {
	SessionID string      `json:"session_id"`
	GameCode  string      `json:"game_code"`
	Action    string      `json:"action"`
	Data      interface{} `json:"data"`
}

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[string]*Client),
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Mu.Lock()
			h.Clients[client.ID] = client
			h.Mu.Unlock()
			log.Printf("Client registered: %s (User: %s)", client.ID, client.UserID)

		case client := <-h.Unregister:
			h.Mu.Lock()
			if _, ok := h.Clients[client.ID]; ok {
				delete(h.Clients, client.ID)
				close(client.Send)
			}
			h.Mu.Unlock()
			log.Printf("Client unregistered: %s", client.ID)

		case message := <-h.Broadcast:
			h.Mu.RLock()
			for _, client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.Clients, client.ID)
				}
			}
			h.Mu.RUnlock()
		}
	}
}

func (h *Hub) SendToUser(userID string, message []byte) {
	h.Mu.RLock()
	defer h.Mu.RUnlock()

	for _, client := range h.Clients {
		if client.UserID == userID {
			select {
			case client.Send <- message:
			default:
				close(client.Send)
				delete(h.Clients, client.ID)
			}
		}
	}
}

func (h *Hub) SendToSession(sessionID string, message []byte) {
	h.Broadcast <- message
}

func (h *Hub) HandleWebSocket(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	client := &Client{
		ID:     uuid.New().String(),
		Conn:   conn,
		UserID: userID.(string),
		Send:   make(chan []byte, 256),
	}

	h.Register <- client

	go h.writePump(client)
	go h.readPump(client)
}

func (h *Hub) writePump(client *Client) {
	defer func() {
		client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := client.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		}
	}
}

func (h *Hub) readPump(client *Client) {
	defer func() {
		h.Unregister <- client
		client.Conn.Close()
	}()

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Failed to parse message: %v", err)
			continue
		}

		h.handleMessage(client, msg)
	}
}

func (h *Hub) handleMessage(client *Client, msg Message) {
	switch msg.Type {
	case "ping":
		response := map[string]string{"type": "pong", "message": "pong"}
		data, _ := json.Marshal(response)
		client.Send <- data

	case "subscribe_game":
		var req struct {
			SessionID string `json:"session_id"`
			GameCode  string `json:"game_code"`
		}
		if err := json.Unmarshal(msg.Payload, &req); err != nil {
			return
		}

		response := map[string]interface{}{
			"type":    "subscribed",
			"message": "Subscribed to game updates",
			"data": map[string]string{
				"session_id": req.SessionID,
				"game_code":  req.GameCode,
			},
		}
		data, _ := json.Marshal(response)
		client.Send <- data

	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}
}
