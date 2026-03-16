package websocket

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/mendmzury/food-delivery/pkg/config"
	"github.com/mendmzury/food-delivery/pkg/utils"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type NotificationMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func HandleWebSocket(hub *Hub, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			token := c.Query("token")
			if token == "" {
				authHeader := c.GetHeader("Authorization")
				if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
					token = strings.TrimPrefix(authHeader, "Bearer ")
				}
			}
			if token == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
				return
			}
			jwtManager := utils.NewJWTManager(cfg.JWTSecret, cfg.JWTExpiry)
			claims, err := jwtManager.ValidateToken(token)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
				return
			}
			userID = claims.UserID
		}
		userIDStr, ok := userID.(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user ID"})
			return
		}
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to upgrade connection"})
			return
		}
		client := NewClient(hub, conn, userIDStr)
		hub.register <- client
		go client.WritePump()
		go client.ReadPump()
	}
}

func BroadcastNotification(hub *Hub, userID string, notification interface{}) bool {
	message := NotificationMessage{
		Type: "notification",
		Data: notification,
	}
	messageBytes, err := json.Marshal(message)
	if err != nil {
		return false
	}
	return hub.SendToUser(userID, messageBytes)
}
