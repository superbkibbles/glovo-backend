package middleware

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Food delivery permissions
const (
	PermUsersView   = "users.view"
	PermUsersSelect = "users.select"
	PermUsersCreate = "users.create"
	PermUsersEdit   = "users.edit"
	PermUsersDelete = "users.delete"

	PermRolesView   = "roles.view"
	PermRolesSelect = "roles.select"
	PermRolesCreate = "roles.create"
	PermRolesEdit   = "roles.edit"
	PermRolesDelete = "roles.delete"

	PermRestaurantsView   = "restaurants.view"
	PermRestaurantsSelect = "restaurants.select"
	PermRestaurantsCreate = "restaurants.create"
	PermRestaurantsEdit   = "restaurants.edit"
	PermRestaurantsDelete = "restaurants.delete"

	PermOrdersView   = "orders.view"
	PermOrdersCreate = "orders.create"
	PermOrdersEdit   = "orders.edit"
	PermOrdersDelete = "orders.delete"

	PermDeliveryView   = "delivery.view"
	PermDeliveryManage = "delivery.manage"

	PermFilesView   = "files.view"
	PermFilesUpload = "files.upload"
	PermFilesDelete = "files.delete"

	PermNotificationsView = "notifications.view"

	PermWebSocketViewSend = "websocket.view_send"

	PermSharedDriveView   = "shared_drive.view"
	PermSharedDriveCreate = "shared_drive.create"
	PermSharedDriveEdit   = "shared_drive.edit"
	PermSharedDriveDelete = "shared_drive.delete"
)

// AllPermissions returns all available permissions
func AllPermissions() []string {
	return []string{
		PermUsersView, PermUsersSelect, PermUsersCreate, PermUsersEdit, PermUsersDelete,
		PermRolesView, PermRolesSelect, PermRolesCreate, PermRolesEdit, PermRolesDelete,
		PermRestaurantsView, PermRestaurantsSelect, PermRestaurantsCreate, PermRestaurantsEdit, PermRestaurantsDelete,
		PermOrdersView, PermOrdersCreate, PermOrdersEdit, PermOrdersDelete,
		PermDeliveryView, PermDeliveryManage,
		PermFilesView, PermFilesUpload, PermFilesDelete,
		PermNotificationsView,
		PermWebSocketViewSend,
		PermSharedDriveView, PermSharedDriveCreate, PermSharedDriveEdit, PermSharedDriveDelete,
	}
}

// CustomerPermissions returns permissions for customer role
func CustomerPermissions() []string {
	return []string{
		PermRestaurantsView, PermRestaurantsSelect,
		PermOrdersView, PermOrdersCreate,
	}
}

// PermissionChecker provides permission checking functionality
type PermissionChecker struct{}

// NewPermissionChecker creates a new permission checker
func NewPermissionChecker() *PermissionChecker {
	return &PermissionChecker{}
}

// HasPermission checks if the user has the required permission
func (p *PermissionChecker) HasPermission(ctx context.Context, requiredPermission string) error {
	claims, err := GetUserFromContext(ctx)
	if err != nil {
		return err
	}
	if claims.IsSuperadmin {
		return nil
	}
	for _, perm := range claims.Permissions {
		if perm == requiredPermission {
			return nil
		}
	}
	return status.Errorf(codes.PermissionDenied, "missing required permission: %s", requiredPermission)
}

// HasAnyPermission checks if the user has any of the required permissions
func (p *PermissionChecker) HasAnyPermission(ctx context.Context, requiredPermissions ...string) error {
	claims, err := GetUserFromContext(ctx)
	if err != nil {
		return err
	}
	if claims.IsSuperadmin {
		return nil
	}
	for _, required := range requiredPermissions {
		for _, perm := range claims.Permissions {
			if perm == required {
				return nil
			}
		}
	}
	return status.Error(codes.PermissionDenied, "missing required permissions")
}

// HasAllPermissions checks if the user has all of the required permissions
func (p *PermissionChecker) HasAllPermissions(ctx context.Context, requiredPermissions ...string) error {
	claims, err := GetUserFromContext(ctx)
	if err != nil {
		return err
	}
	if claims.IsSuperadmin {
		return nil
	}
	permSet := make(map[string]struct{}, len(claims.Permissions))
	for _, perm := range claims.Permissions {
		permSet[perm] = struct{}{}
	}
	for _, required := range requiredPermissions {
		if _, ok := permSet[required]; !ok {
			return status.Errorf(codes.PermissionDenied, "missing required permission: %s", required)
		}
	}
	return nil
}

// IsSuperadmin checks if the user is a superadmin
func (p *PermissionChecker) IsSuperadmin(ctx context.Context) bool {
	claims, err := GetUserFromContext(ctx)
	if err != nil {
		return false
	}
	return claims.IsSuperadmin
}

// RequireSuperadmin ensures the user is a superadmin
func (p *PermissionChecker) RequireSuperadmin(ctx context.Context) error {
	if !p.IsSuperadmin(ctx) {
		return status.Error(codes.PermissionDenied, "superadmin access required")
	}
	return nil
}
