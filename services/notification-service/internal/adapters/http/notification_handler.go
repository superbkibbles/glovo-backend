package http

import (
	"net/http"

	"glovo-backend/services/notification-service/internal/domain"
	"glovo-backend/shared/auth"
	"glovo-backend/shared/middleware"

	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	notificationService domain.NotificationService
}

func NewNotificationHandler(notificationService domain.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
	}
}

func (h *NotificationHandler) SetupRoutes(router *gin.RouterGroup) {
	// Public notification endpoints (for internal service calls)
	public := router.Group("/notifications")
	{
		public.POST("/send", h.sendNotification)
		public.POST("/send-bulk", h.sendBulkNotification)
	}

	// User notification preferences and history
	user := router.Group("/user/notifications")
	user.Use(middleware.AuthMiddleware())
	{
		// Notification history
		// user.GET("/", h.getNotificationHistory) // TODO: Add this later
		// user.PUT("/:id/read", h.markAsRead) // TODO: Add this later
		// user.PUT("/read-all", h.markAllAsRead)  // TODO: Add this later

		// Preferences
		// user.GET("/preferences", h.getPreferences)  // TODO: Add this later
		// user.PUT("/preferences", h.updatePreferences)  // TODO: Add this later

		// Device management
		user.POST("/devices", h.registerDevice)
		user.GET("/devices", h.getDevices)
		// user.PUT("/devices/:id", h.updateDevice)
		// user.DELETE("/devices/:id", h.deleteDevice)  // TODO: Add this later
	}

	// Admin notification management
	admin := router.Group("/admin/notifications")
	admin.Use(middleware.AuthMiddleware())
	admin.Use(middleware.RequireRole(auth.RoleAdmin))
	{
		// Templates
		// admin.POST("/templates", h.createTemplate)  // TODO: Add this later
		// admin.GET("/templates", h.getTemplates)  // TODO: Add this later
		admin.GET("/templates/:id", h.getTemplate)
		// admin.PUT("/templates/:id", h.updateTemplate)  // TODO: Add this later
		admin.DELETE("/templates/:id", h.deleteTemplate)

		// Broadcast notifications
		// admin.POST("/broadcast", h.broadcastNotification)

		// Analytics
		// admin.GET("/stats", h.getNotificationStats)  // TODO: Add this later
		// admin.GET("/delivery-stats", h.getDeliveryStats)  // TODO: Add this later

		// All notifications
		// admin.GET("/", h.getAllNotifications)  // TODO: Add this later
		admin.GET("/:id", h.getNotificationDetails)
	}
}

// @Summary Send notification
// @Description Send a notification to a user
// @Tags notifications
// @Accept json
// @Produce json
// @Param request body domain.SendNotificationRequest true "Notification data"
// @Success 200 {object} domain.Notification
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/notifications/send [post]
func (h *NotificationHandler) sendNotification(c *gin.Context) {
	var req domain.SendNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	notification, err := h.notificationService.SendNotification(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, notification)
}

// @Summary Send bulk notification
// @Description Send notifications to multiple users
// @Tags notifications
// @Accept json
// @Produce json
// @Param request body domain.BulkNotificationRequest true "Bulk notification data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/notifications/send-bulk [post]
func (h *NotificationHandler) sendBulkNotification(c *gin.Context) {
	var req domain.BulkNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.notificationService.SendBulkNotification(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// User endpoints

// @Summary Get notification history
// @Description Get notification history for authenticated user
// @Tags user
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit results"
// @Param offset query int false "Offset for pagination"
// @Param type query string false "Notification type filter"
// @Success 200 {array} domain.Notification
// @Failure 500 {object} map[string]string
// @Router /api/v1/user/notifications [get]
// func (h *NotificationHandler) getNotificationHistory(c *gin.Context) {
// 	userID, _ := c.Get("user_id")
// 	limitStr := c.DefaultQuery("limit", "50")
// 	offsetStr := c.DefaultQuery("offset", "0")
// 	notificationType := c.Query("type")

// 	limit, _ := strconv.Atoi(limitStr)
// 	offset, _ := strconv.Atoi(offsetStr)

// 	req := domain.NotificationHistoryRequest{
// 		UserID: userID.(string),
// 		Type:   notificationType,
// 		Limit:  limit,
// 		Offset: offset,
// 	}

// 	notifications, err := h.notificationService.GetNotificationHistory(req)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, notifications)
// }

// @Summary Mark notification as read
// @Description Mark a specific notification as read
// @Tags user
// @Produce json
// @Security BearerAuth
// @Param id path string true "Notification ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/user/notifications/{id}/read [put]
// func (h *NotificationHandler) markAsRead(c *gin.Context) {
// 	notificationID := c.Param("id")
// 	userID, _ := c.Get("user_id")

// 	err := h.notificationService.MarkAsRead(notificationID, userID.(string))
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{"message": "Notification marked as read"})
// }

// @Summary Mark all notifications as read
// @Description Mark all notifications as read for the user
// @Tags user
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/user/notifications/read-all [put]
// func (h *NotificationHandler) markAllAsRead(c *gin.Context) {
// 	userID, _ := c.Get("user_id")

// 	err := h.notificationService.MarkAllAsRead(userID.(string))
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{"message": "All notifications marked as read"})
// }

// @Summary Get notification preferences
// @Description Get user's notification preferences
// @Tags user
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.NotificationPreference
// @Failure 500 {object} map[string]string
// @Router /api/v1/user/notifications/preferences [get]
// func (h *NotificationHandler) getPreferences(c *gin.Context) {
// 	userID, _ := c.Get("user_id")

// 	preferences, err := h.notificationService.GetPreferences(userID.(string))
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, preferences)
// }

// @Summary Update notification preferences
// @Description Update user's notification preferences
// @Tags user
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.UpdatePreferencesRequest true "Preferences data"
// @Success 200 {object} domain.NotificationPreference
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/user/notifications/preferences [put]
// func (h *NotificationHandler) updatePreferences(c *gin.Context) {
// 	var req domain.UpdatePreferencesRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	userID, _ := c.Get("user_id")
// 	req.UserID = userID.(string)

// 	preferences, err := h.notificationService.UpdatePreferences(req)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, preferences)
// }

// @Summary Register device
// @Description Register a device for push notifications
// @Tags user
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.RegisterDeviceRequest true "Device data"
// @Success 201 {object} domain.Device
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/user/notifications/devices [post]
func (h *NotificationHandler) registerDevice(c *gin.Context) {
	var req domain.RegisterDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")
	req.UserID = userID.(string)

	device, err := h.notificationService.RegisterDevice(userID.(string), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, device)
}

// @Summary Get user devices
// @Description Get all registered devices for the user
// @Tags user
// @Produce json
// @Security BearerAuth
// @Success 200 {array} domain.Device
// @Failure 500 {object} map[string]string
// @Router /api/v1/user/notifications/devices [get]
func (h *NotificationHandler) getDevices(c *gin.Context) {
	userID, _ := c.Get("user_id")

	devices, err := h.notificationService.GetUserDevices(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, devices)
}

// @Summary Update device
// @Description Update device information
// @Tags user
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Device ID"
// @Param request body domain.UpdateDeviceRequest true "Device update data"
// @Success 200 {object} domain.Device
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/user/notifications/devices/{id} [put]
// func (h *NotificationHandler) updateDevice(c *gin.Context) {
// 	deviceID := c.Param("id")
// 	userID, _ := c.Get("user_id")

// 	var req domain.UpdateDeviceRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	device, err := h.notificationService.UpdateDevice(deviceID, userID.(string), req)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, device)
// }

// @Summary Delete device
// @Description Delete a registered device
// @Tags user
// @Produce json
// @Security BearerAuth
// @Param id path string true "Device ID"
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/user/notifications/devices/{id} [delete]
// func (h *NotificationHandler) deleteDevice(c *gin.Context) {
// 	deviceID := c.Param("id")
// 	userID, _ := c.Get("user_id")

// 	err := h.notificationService.DeleteDevice(deviceID, userID.(string))
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{"message": "Device deleted successfully"})
// }

// Admin endpoints

// @Summary Create notification template
// @Description Create a new notification template (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.CreateTemplateRequest true "Template data"
// @Success 201 {object} domain.NotificationTemplate
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/notifications/templates [post]
// func (h *NotificationHandler) createTemplate(c *gin.Context) {
// 	var req domain.CreateTemplateRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	adminID, _ := c.Get("user_id")
// 	req.CreatedBy = adminID.(string)

// 	template, err := h.notificationService.CreateTemplate(req)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusCreated, template)
// }

// @Summary Get notification templates
// @Description Get all notification templates (admin only)
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param type query string false "Template type filter"
// @Success 200 {array} domain.NotificationTemplate
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/notifications/templates [get]
// func (h *NotificationHandler) getTemplates(c *gin.Context) {
// 	templateType := c.Query("type")

// 	templates, err := h.notificationService.GetTemplates(templateType)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, templates)
// }

// @Summary Get notification template
// @Description Get a specific notification template (admin only)
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param id path string true "Template ID"
// @Success 200 {object} domain.NotificationTemplate
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/notifications/templates/{id} [get]
func (h *NotificationHandler) getTemplate(c *gin.Context) {
	templateID := c.Param("id")

	template, err := h.notificationService.GetTemplate(templateID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, template)
}

// @Summary Update notification template
// @Description Update a notification template (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Template ID"
// @Param request body domain.UpdateTemplateRequest true "Template update data"
// @Success 200 {object} domain.NotificationTemplate
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/notifications/templates/{id} [put]
// func (h *NotificationHandler) updateTemplate(c *gin.Context) {
// 	templateID := c.Param("id")

// 	var req domain.UpdateTemplateRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	adminID, _ := c.Get("user_id")
// 	req.UpdatedBy = adminID.(string)

// 	template, err := h.notificationService.UpdateTemplate(templateID, req)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, template)
// }

// @Summary Delete notification template
// @Description Delete a notification template (admin only)
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param id path string true "Template ID"
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/notifications/templates/{id} [delete]
func (h *NotificationHandler) deleteTemplate(c *gin.Context) {
	templateID := c.Param("id")

	err := h.notificationService.DeleteTemplate(templateID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Template deleted successfully"})
}

// @Summary Broadcast notification
// @Description Send a broadcast notification to multiple users (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.BroadcastRequest true "Broadcast data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/notifications/broadcast [post]
// func (h *NotificationHandler) broadcastNotification(c *gin.Context) {
// 	var req domain.BroadcastRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	adminID, _ := c.Get("user_id")
// 	req.SentBy = adminID.(string)

// 	result, err := h.notificationService.BroadcastNotification(req)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, result)
// }

// @Summary Get notification statistics
// @Description Get notification delivery statistics (admin only)
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.NotificationStats
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/notifications/stats [get]
// func (h *NotificationHandler) getNotificationStats(c *gin.Context) {
// 	stats, err := h.notificationService.GetNotificationStats()
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, stats)
// }

// @Summary Get delivery statistics
// @Description Get notification delivery success/failure statistics (admin only)
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param start_date query string false "Start date filter (YYYY-MM-DD)"
// @Param end_date query string false "End date filter (YYYY-MM-DD)"
// @Success 200 {object} domain.DeliveryStats
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/notifications/delivery-stats [get]
// func (h *NotificationHandler) getDeliveryStats(c *gin.Context) {
// 	startDate := c.Query("start_date")
// 	endDate := c.Query("end_date")

// 	req := domain.DeliveryStatsRequest{
// 		StartDate: startDate,
// 		EndDate:   endDate,
// 	}

// 	stats, err := h.notificationService.GetDeliveryStats(req)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, stats)
// }

// @Summary Get all notifications
// @Description Get all notifications in the system (admin only)
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit results"
// @Param offset query int false "Offset for pagination"
// @Param status query string false "Status filter"
// @Param type query string false "Type filter"
// @Success 200 {array} domain.Notification
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/notifications [get]
// func (h *NotificationHandler) getAllNotifications(c *gin.Context) {
// 	limitStr := c.DefaultQuery("limit", "100")
// 	offsetStr := c.DefaultQuery("offset", "0")
// 	status := c.Query("status")
// 	notificationType := c.Query("type")

// 	limit, _ := strconv.Atoi(limitStr)
// 	offset, _ := strconv.Atoi(offsetStr)

// 	req := domain.AdminNotificationRequest{
// 		Limit:  limit,
// 		Offset: offset,
// 		Status: status,
// 		Type:   notificationType,
// 	}

// 	notifications, err := h.notificationService.GetAllNotifications(req)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, notifications)
// }

// @Summary Get notification details
// @Description Get detailed information about a specific notification (admin only)
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param id path string true "Notification ID"
// @Success 200 {object} domain.Notification
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/notifications/{id} [get]
func (h *NotificationHandler) getNotificationDetails(c *gin.Context) {
	notificationID := c.Param("id")

	notification, err := h.notificationService.GetNotification(notificationID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, notification)
}
