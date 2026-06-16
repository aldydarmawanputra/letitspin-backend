package router

import (
	"let-it-spin/internal/websocket"

	"github.com/gin-gonic/gin"
)

func SetupWebSocket(r *gin.Engine, wsHub *websocket.Hub) {
	r.GET("/ws", func(c *gin.Context) {
		wsHub.HandleWebSocket(c)
	})
}
