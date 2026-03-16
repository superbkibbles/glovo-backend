package server

import (
	"context"

	"github.com/mendmzury/food-delivery/services/delivery-service/internal/application"
	deliverypb "github.com/mendmzury/food-delivery/proto/delivery"
	commonpb "github.com/mendmzury/food-delivery/proto/common"
)

type DeliveryServer struct {
	deliverypb.UnimplementedDeliveryServiceServer
	service *application.DeliveryService
}

func NewDeliveryServer(service *application.DeliveryService) *DeliveryServer {
	return &DeliveryServer{service: service}
}

func (s *DeliveryServer) AssignDriver(ctx context.Context, req *deliverypb.AssignDriverRequest) (*deliverypb.DeliveryAssignment, error) {
	return s.service.AssignDriver(ctx, req)
}

func (s *DeliveryServer) GetAssignment(ctx context.Context, req *deliverypb.GetAssignmentRequest) (*deliverypb.DeliveryAssignment, error) {
	return s.service.GetAssignment(ctx, req)
}

func (s *DeliveryServer) ListAssignments(ctx context.Context, req *deliverypb.ListAssignmentsRequest) (*deliverypb.ListAssignmentsResponse, error) {
	return s.service.ListAssignments(ctx, req)
}

func (s *DeliveryServer) UpdateLocation(ctx context.Context, req *deliverypb.UpdateLocationRequest) (*commonpb.SuccessResponse, error) {
	return s.service.UpdateLocation(ctx, req)
}

func (s *DeliveryServer) CompleteDelivery(ctx context.Context, req *deliverypb.CompleteDeliveryRequest) (*deliverypb.DeliveryAssignment, error) {
	return s.service.CompleteDelivery(ctx, req)
}

func (s *DeliveryServer) GetTrackingHistory(ctx context.Context, req *deliverypb.GetTrackingHistoryRequest) (*deliverypb.GetTrackingHistoryResponse, error) {
	return s.service.GetTrackingHistory(ctx, req)
}

func (s *DeliveryServer) ListDriverLocations(ctx context.Context, req *deliverypb.ListDriverLocationsRequest) (*deliverypb.ListDriverLocationsResponse, error) {
	return s.service.ListDriverLocations(ctx, req)
}
