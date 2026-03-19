package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mendmzury/food-delivery/pkg/logger"
	"github.com/mendmzury/food-delivery/services/gateway/internal/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func successResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}

func createdResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": data})
}

func errorResponse(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, gin.H{"success": false, "error": message})
}

func paginatedResponse(c *gin.Context, data interface{}, total int64, page, pageSize int64) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if total < 0 {
		total = 0
	}
	var totalPages int64
	if total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}
	hasNextPage := page < totalPages
	hasPrevPage := page > 1
	pagination := gin.H{
		"page":           int(page),
		"page_size":      int(pageSize),
		"total":          int(total),
		"total_pages":    int(totalPages),
		"has_next_page":  hasNextPage,
		"has_prev_page":  hasPrevPage,
		"next_page":      nil,
		"prev_page":      nil,
	}
	if hasNextPage {
		pagination["next_page"] = int(page + 1)
	}
	if hasPrevPage {
		pagination["prev_page"] = int(page - 1)
	}
	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"data":       data,
		"pagination": pagination,
	})
}

func handleGRPCError(c *gin.Context, err error) {
	st, ok := status.FromError(err)
	if !ok {
		logger.Error().Err(err).Msg("Unknown error")
		errorResponse(c, http.StatusInternalServerError, "internal server error")
		return
	}
	switch st.Code() {
	case codes.Unauthenticated:
		errorResponse(c, http.StatusUnauthorized, st.Message())
	case codes.PermissionDenied:
		errorResponse(c, http.StatusForbidden, st.Message())
	case codes.InvalidArgument:
		errorResponse(c, http.StatusBadRequest, st.Message())
	case codes.NotFound:
		errorResponse(c, http.StatusNotFound, st.Message())
	case codes.AlreadyExists:
		errorResponse(c, http.StatusConflict, st.Message())
	case codes.Unavailable:
		logger.Error().Err(err).Msg("gRPC service unavailable")
		errorResponse(c, http.StatusServiceUnavailable, "service temporarily unavailable")
	default:
		errorResponse(c, http.StatusInternalServerError, "internal server error")
	}
}

func getAuthContext(c *gin.Context) context.Context {
	token, _ := c.Get("token")
	if token != nil {
		return grpc.WithAuth(c.Request.Context(), token.(string))
	}
	return c.Request.Context()
}

// requirePermission checks if the authenticated user has the given permission.
// Returns true and aborts with 403 if the permission is missing.
func requirePermission(c *gin.Context, permission string) bool {
	isSuperadmin, _ := c.Get("is_superadmin")
	if v, ok := isSuperadmin.(bool); ok && v {
		return false
	}
	permissions, _ := c.Get("permissions")
	if perms, ok := permissions.([]string); ok {
		for _, p := range perms {
			if p == permission {
				return false
			}
		}
	}
	errorResponse(c, http.StatusForbidden, "missing required permission: "+permission)
	c.Abort()
	return true
}

func getPaginationParams(c *gin.Context) (int64, int64) {
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	pageSize, _ := strconv.ParseInt(c.DefaultQuery("page_size", "10"), 10, 64)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	return page, pageSize
}
