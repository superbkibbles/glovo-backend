package server

import (
	"context"

	notificationpb "github.com/mendmzury/food-delivery/proto/notification"
	commonpb "github.com/mendmzury/food-delivery/proto/common"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type NotificationServer struct {
	notificationpb.UnimplementedNotificationServiceServer
}

func NewNotificationServer() *NotificationServer {
	return &NotificationServer{}
}

func (s *NotificationServer) SendNotification(ctx context.Context, req *notificationpb.SendNotificationRequest) (*notificationpb.Notification, error) {
	return nil, status.Error(codes.Unimplemented, "stub")
}

func (s *NotificationServer) GetNotification(ctx context.Context, req *notificationpb.GetNotificationRequest) (*notificationpb.Notification, error) {
	return nil, status.Error(codes.Unimplemented, "stub")
}

func (s *NotificationServer) ListNotifications(ctx context.Context, req *notificationpb.ListNotificationsRequest) (*notificationpb.ListNotificationsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "stub")
}

func (s *NotificationServer) MarkAsRead(ctx context.Context, req *notificationpb.MarkAsReadRequest) (*notificationpb.Notification, error) {
	return nil, status.Error(codes.Unimplemented, "stub")
}

func (s *NotificationServer) MarkAllAsRead(ctx context.Context, req *notificationpb.MarkAllAsReadRequest) (*commonpb.SuccessResponse, error) {
	return nil, status.Error(codes.Unimplemented, "stub")
}

func (s *NotificationServer) DeleteNotification(ctx context.Context, req *notificationpb.DeleteNotificationRequest) (*commonpb.SuccessResponse, error) {
	return nil, status.Error(codes.Unimplemented, "stub")
}

func (s *NotificationServer) GetUnreadCount(ctx context.Context, req *notificationpb.GetUnreadCountRequest) (*notificationpb.GetUnreadCountResponse, error) {
	return nil, status.Error(codes.Unimplemented, "stub")
}

func (s *NotificationServer) RegisterDeviceToken(ctx context.Context, req *notificationpb.RegisterDeviceTokenRequest) (*commonpb.SuccessResponse, error) {
	return nil, status.Error(codes.Unimplemented, "stub")
}

func (s *NotificationServer) UnregisterDeviceToken(ctx context.Context, req *notificationpb.UnregisterDeviceTokenRequest) (*commonpb.SuccessResponse, error) {
	return nil, status.Error(codes.Unimplemented, "stub")
}
