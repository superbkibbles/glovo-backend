package application

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mendmzury/food-delivery/pkg/config"
	"github.com/mendmzury/food-delivery/pkg/database"
	"github.com/mendmzury/food-delivery/services/delivery-service/internal/domain"
	"github.com/mendmzury/food-delivery/services/delivery-service/internal/repository"
	deliverypb "github.com/mendmzury/food-delivery/proto/delivery"
	commonpb "github.com/mendmzury/food-delivery/proto/common"
	orderpb "github.com/mendmzury/food-delivery/proto/order"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const driverLocationKeyPrefix = "driver:location:"
const driverLocationTTL = 5 * time.Minute

var statusToProto = map[domain.AssignmentStatus]deliverypb.AssignmentStatus{
	domain.AssignmentStatusAssigned:  deliverypb.AssignmentStatus_ASSIGNMENT_STATUS_ASSIGNED,
	domain.AssignmentStatusPickedUp:  deliverypb.AssignmentStatus_ASSIGNMENT_STATUS_PICKED_UP,
	domain.AssignmentStatusDelivered: deliverypb.AssignmentStatus_ASSIGNMENT_STATUS_DELIVERED,
}

var protoToStatus = map[deliverypb.AssignmentStatus]domain.AssignmentStatus{
	deliverypb.AssignmentStatus_ASSIGNMENT_STATUS_ASSIGNED:  domain.AssignmentStatusAssigned,
	deliverypb.AssignmentStatus_ASSIGNMENT_STATUS_PICKED_UP:  domain.AssignmentStatusPickedUp,
	deliverypb.AssignmentStatus_ASSIGNMENT_STATUS_DELIVERED: domain.AssignmentStatusDelivered,
}

type DeliveryService struct {
	assignmentRepo *repository.AssignmentRepository
	trackingRepo   *repository.TrackingRepository
	redisDB        *database.RedisDB
	cfg            *config.Config
}

func NewDeliveryService(assignmentRepo *repository.AssignmentRepository, trackingRepo *repository.TrackingRepository, redisDB *database.RedisDB, cfg *config.Config) *DeliveryService {
	return &DeliveryService{
		assignmentRepo: assignmentRepo,
		trackingRepo:   trackingRepo,
		redisDB:        redisDB,
		cfg:            cfg,
	}
}

func (s *DeliveryService) getOrderClient() (orderpb.OrderServiceClient, func() error, error) {
	conn, err := grpc.NewClient(s.cfg.OrderServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	return orderpb.NewOrderServiceClient(conn), conn.Close, nil
}

func (s *DeliveryService) AssignDriver(ctx context.Context, req *deliverypb.AssignDriverRequest) (*deliverypb.DeliveryAssignment, error) {
	orderID, err := uuid.Parse(req.OrderId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid order_id")
	}
	driverID, err := uuid.Parse(req.DriverId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid driver_id")
	}

	existing, err := s.assignmentRepo.GetByOrderID(ctx, orderID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if existing != nil {
		return nil, status.Error(codes.AlreadyExists, "order already has assignment")
	}

	orderClient, closeFn, err := s.getOrderClient()
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to connect to order service")
	}
	defer closeFn()

	_, err = orderClient.UpdateOrderStatus(ctx, &orderpb.UpdateOrderStatusRequest{
		Id:     req.OrderId,
		Status: orderpb.OrderStatus_ORDER_STATUS_ASSIGNED,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to update order status: "+err.Error())
	}

	a := &domain.DeliveryAssignment{
		OrderID:    orderID,
		DriverID:   driverID,
		Status:     domain.AssignmentStatusAssigned,
		AssignedAt:  time.Now(),
	}
	if err := s.assignmentRepo.Create(ctx, a); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return assignmentToProto(a, 0, 0), nil
}

func (s *DeliveryService) GetAssignment(ctx context.Context, req *deliverypb.GetAssignmentRequest) (*deliverypb.DeliveryAssignment, error) {
	var a *domain.DeliveryAssignment
	var err error

	if req.Id != "" {
		id, parseErr := uuid.Parse(req.Id)
		if parseErr != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid id")
		}
		a, err = s.assignmentRepo.GetByID(ctx, id)
	} else if req.OrderId != "" {
		orderID, parseErr := uuid.Parse(req.OrderId)
		if parseErr != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid order_id")
		}
		a, err = s.assignmentRepo.GetByOrderID(ctx, orderID)
	} else {
		return nil, status.Error(codes.InvalidArgument, "id or order_id required")
	}

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if a == nil {
		return nil, status.Error(codes.NotFound, "assignment not found")
	}

	lat, lng := s.getDriverLocation(ctx, a.DriverID.String())
	return assignmentToProto(a, lat, lng), nil
}

func (s *DeliveryService) ListAssignments(ctx context.Context, req *deliverypb.ListAssignmentsRequest) (*deliverypb.ListAssignmentsResponse, error) {
	page, pageSize := int64(1), int64(10)
	if req.Pagination != nil {
		if req.Pagination.Page > 0 {
			page = req.Pagination.Page
		}
		if req.Pagination.PageSize > 0 {
			pageSize = req.Pagination.PageSize
		}
	}
	if pageSize > 100 {
		pageSize = 100
	}

	var driverID *uuid.UUID
	if req.DriverId != "" {
		id, err := uuid.Parse(req.DriverId)
		if err == nil {
			driverID = &id
		}
	}

	assignmentStatus := domain.AssignmentStatus("")
	if req.Status != 0 {
		assignmentStatus = protoToStatus[req.Status]
	}

	assignments, total, err := s.assignmentRepo.List(ctx, driverID, assignmentStatus, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	protoAssignments := make([]*deliverypb.DeliveryAssignment, len(assignments))
	for i, a := range assignments {
		lat, lng := s.getDriverLocation(ctx, a.DriverID.String())
		protoAssignments[i] = assignmentToProto(a, lat, lng)
	}

	totalPages := int64(0)
	if total > 0 && pageSize > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}

	return &deliverypb.ListAssignmentsResponse{
		Assignments: protoAssignments,
		Pagination: &commonpb.PaginationMeta{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
			HasNext:    page < totalPages,
			HasPrev:    page > 1,
		},
	}, nil
}

func (s *DeliveryService) UpdateLocation(ctx context.Context, req *deliverypb.UpdateLocationRequest) (*commonpb.SuccessResponse, error) {
	if req.DriverId == "" {
		return nil, status.Error(codes.InvalidArgument, "driver_id required")
	}
	key := driverLocationKeyPrefix + req.DriverId
	value := fmt.Sprintf("%f,%f", req.Lat, req.Lng)
	if err := s.redisDB.Set(ctx, key, value, driverLocationTTL); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	driverID, err := uuid.Parse(req.DriverId)
	if err == nil {
		if a, _ := s.assignmentRepo.GetActiveByDriverID(ctx, driverID); a != nil {
			_ = s.trackingRepo.Create(ctx, &domain.DeliveryLocationHistory{
				AssignmentID: a.ID,
				DriverID:     driverID,
				Lat:          req.Lat,
				Lng:          req.Lng,
				RecordedAt:   time.Now(),
			})
		}
	}
	return &commonpb.SuccessResponse{Success: true}, nil
}

func (s *DeliveryService) CompleteDelivery(ctx context.Context, req *deliverypb.CompleteDeliveryRequest) (*deliverypb.DeliveryAssignment, error) {
	id, err := uuid.Parse(req.AssignmentId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid assignment_id")
	}
	a, err := s.assignmentRepo.GetByID(ctx, id)
	if err != nil || a == nil {
		return nil, status.Error(codes.NotFound, "assignment not found")
	}

	now := time.Now()
	a.Status = domain.AssignmentStatusDelivered
	a.DeliveredAt = &now
	if err := s.assignmentRepo.Update(ctx, a); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	orderClient, closeFn, err := s.getOrderClient()
	if err == nil {
		defer closeFn()
		orderClient.UpdateOrderStatus(ctx, &orderpb.UpdateOrderStatusRequest{
			Id:     a.OrderID.String(),
			Status: orderpb.OrderStatus_ORDER_STATUS_DELIVERED,
		})
	}

	lat, lng := s.getDriverLocation(ctx, a.DriverID.String())
	return assignmentToProto(a, lat, lng), nil
}

func (s *DeliveryService) getDriverLocation(ctx context.Context, driverID string) (float64, float64) {
	key := driverLocationKeyPrefix + driverID
	val, err := s.redisDB.Get(ctx, key)
	if err != nil || val == "" {
		return 0, 0
	}
	parts := strings.Split(val, ",")
	if len(parts) != 2 {
		return 0, 0
	}
	lat, _ := strconv.ParseFloat(parts[0], 64)
	lng, _ := strconv.ParseFloat(parts[1], 64)
	return lat, lng
}

func (s *DeliveryService) GetTrackingHistory(ctx context.Context, req *deliverypb.GetTrackingHistoryRequest) (*deliverypb.GetTrackingHistoryResponse, error) {
	if req.AssignmentId == "" {
		return nil, status.Error(codes.InvalidArgument, "assignment_id required")
	}
	assignmentID, err := uuid.Parse(req.AssignmentId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid assignment_id")
	}
	points, err := s.trackingRepo.ListByAssignmentID(ctx, assignmentID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	protoPoints := make([]*deliverypb.DeliveryLocationPoint, len(points))
	for i, p := range points {
		protoPoints[i] = &deliverypb.DeliveryLocationPoint{
			AssignmentId: p.AssignmentID.String(),
			DriverId:     p.DriverID.String(),
			Lat:          p.Lat,
			Lng:          p.Lng,
			RecordedAt:   timestamppb.New(p.RecordedAt),
		}
	}
	return &deliverypb.GetTrackingHistoryResponse{Points: protoPoints}, nil
}

func (s *DeliveryService) ListDriverLocations(ctx context.Context, req *deliverypb.ListDriverLocationsRequest) (*deliverypb.ListDriverLocationsResponse, error) {
	keys, err := s.redisDB.Keys(ctx, driverLocationKeyPrefix+"*")
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	var locations []*deliverypb.DriverLocation
	for _, key := range keys {
		driverID := strings.TrimPrefix(key, driverLocationKeyPrefix)
		val, err := s.redisDB.Get(ctx, key)
		if err != nil || val == "" {
			continue
		}
		parts := strings.Split(val, ",")
		if len(parts) != 2 {
			continue
		}
		lat, _ := strconv.ParseFloat(parts[0], 64)
		lng, _ := strconv.ParseFloat(parts[1], 64)
		locations = append(locations, &deliverypb.DriverLocation{
			DriverId:  driverID,
			Lat:       lat,
			Lng:       lng,
			UpdatedAt: timestamppb.Now(),
		})
	}
	return &deliverypb.ListDriverLocationsResponse{Locations: locations}, nil
}

func assignmentToProto(a *domain.DeliveryAssignment, lat, lng float64) *deliverypb.DeliveryAssignment {
	if a == nil {
		return nil
	}
	res := &deliverypb.DeliveryAssignment{
		Id:         a.ID.String(),
		OrderId:    a.OrderID.String(),
		DriverId:   a.DriverID.String(),
		DriverName: a.DriverName,
		Status:     statusToProto[a.Status],
		DriverLat:  a.DriverLat,
		DriverLng:  a.DriverLng,
		AssignedAt: timestamppb.New(a.AssignedAt),
	}
	if lat != 0 || lng != 0 {
		res.DriverLat = lat
		res.DriverLng = lng
	}
	if a.PickedUpAt != nil {
		res.PickedUpAt = timestamppb.New(*a.PickedUpAt)
	}
	if a.DeliveredAt != nil {
		res.DeliveredAt = timestamppb.New(*a.DeliveredAt)
	}
	return res
}
