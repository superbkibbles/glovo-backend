package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mendmzury/food-delivery/pkg/config"
	"github.com/mendmzury/food-delivery/services/gateway/internal/grpc"
	"github.com/mendmzury/food-delivery/services/gateway/internal/websocket"
	userpb "github.com/mendmzury/food-delivery/proto/user"
)

func BroadcastNotification(cfg *config.Config, wsHub *websocket.Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			UserID       string      `json:"user_id" binding:"required"`
			Notification interface{} `json:"notification" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if !wsHub.IsUserConnected(req.UserID) {
			c.JSON(http.StatusOK, gin.H{"sent": false, "reason": "user_not_connected", "user_id": req.UserID})
			return
		}
		sent := websocket.BroadcastNotification(wsHub, req.UserID, req.Notification)
		if sent {
			c.JSON(http.StatusOK, gin.H{"sent": true, "method": "websocket", "user_id": req.UserID})
		} else {
			c.JSON(http.StatusOK, gin.H{"sent": false, "reason": "failed_to_send", "user_id": req.UserID})
		}
	}
}

func CheckWebSocketConnection(cfg *config.Config, wsHub *websocket.Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Query("user_id")
		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
			return
		}
		isConnected := wsHub.IsUserConnected(userID)
		c.JSON(http.StatusOK, gin.H{"connected": isConnected})
	}
}

// ListConnections godoc
// @Summary List WebSocket connections
// @Tags websocket
// @Security BearerAuth
// @Success 200 {object} object
// @Router /websocket/connections [get]
func ListConnections(cfg *config.Config, wsHub *websocket.Hub, clients *grpc.Clients) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDs := wsHub.GetConnectedUserIDs()
		connections := make([]gin.H, 0, len(userIDs))
		ctx := getAuthContext(c)
		for _, id := range userIDs {
			conn := gin.H{"user_id": id}
			if clients != nil && clients.User != nil {
				user, err := clients.User.GetUser(ctx, &userpb.GetUserRequest{Id: id})
				if err == nil && user != nil && user.Name != "" {
					conn["name"] = user.Name
				}
			}
			connections = append(connections, conn)
		}
		c.JSON(http.StatusOK, gin.H{"connections": connections, "user_ids": userIDs})
	}
}
