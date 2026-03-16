package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/mendmzury/food-delivery/pkg/config"
	"github.com/mendmzury/food-delivery/services/order-service/internal/domain"
	"github.com/mendmzury/food-delivery/services/order-service/internal/repository"
	orderpb "github.com/mendmzury/food-delivery/proto/order"
	commonpb "github.com/mendmzury/food-delivery/proto/common"
	restaurantpb "github.com/mendmzury/food-delivery/proto/restaurant"
	settingspb "github.com/mendmzury/food-delivery/proto/settings"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var statusToProto = map[domain.OrderStatus]orderpb.OrderStatus{
	domain.OrderStatusPending:   orderpb.OrderStatus_ORDER_STATUS_PENDING,
	domain.OrderStatusAccepted:  orderpb.OrderStatus_ORDER_STATUS_ACCEPTED,
	domain.OrderStatusPreparing: orderpb.OrderStatus_ORDER_STATUS_PREPARING,
	domain.OrderStatusReady:     orderpb.OrderStatus_ORDER_STATUS_READY,
	domain.OrderStatusAssigned:  orderpb.OrderStatus_ORDER_STATUS_ASSIGNED,
	domain.OrderStatusPickedUp:  orderpb.OrderStatus_ORDER_STATUS_PICKED_UP,
	domain.OrderStatusDelivered: orderpb.OrderStatus_ORDER_STATUS_DELIVERED,
	domain.OrderStatusCancelled: orderpb.OrderStatus_ORDER_STATUS_CANCELLED,
}

var protoToStatus = map[orderpb.OrderStatus]domain.OrderStatus{
	orderpb.OrderStatus_ORDER_STATUS_PENDING:   domain.OrderStatusPending,
	orderpb.OrderStatus_ORDER_STATUS_ACCEPTED:  domain.OrderStatusAccepted,
	orderpb.OrderStatus_ORDER_STATUS_PREPARING: domain.OrderStatusPreparing,
	orderpb.OrderStatus_ORDER_STATUS_READY:     domain.OrderStatusReady,
	orderpb.OrderStatus_ORDER_STATUS_ASSIGNED:  domain.OrderStatusAssigned,
	orderpb.OrderStatus_ORDER_STATUS_PICKED_UP: domain.OrderStatusPickedUp,
	orderpb.OrderStatus_ORDER_STATUS_DELIVERED: domain.OrderStatusDelivered,
	orderpb.OrderStatus_ORDER_STATUS_CANCELLED: domain.OrderStatusCancelled,
}

type OrderService struct {
	orderRepo *repository.OrderRepository
	cfg       *config.Config
}

func NewOrderService(orderRepo *repository.OrderRepository, cfg *config.Config) *OrderService {
	return &OrderService{orderRepo: orderRepo, cfg: cfg}
}

func (s *OrderService) getRestaurantClient() (restaurantpb.RestaurantServiceClient, func() error, error) {
	conn, err := grpc.NewClient(s.cfg.RestaurantServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	return restaurantpb.NewRestaurantServiceClient(conn), conn.Close, nil
}

func (s *OrderService) getCommissionPercent(ctx context.Context) float64 {
	conn, err := grpc.NewClient(s.cfg.SettingsServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return 0
	}
	defer conn.Close()
	client := settingspb.NewSettingsServiceClient(conn)
	resp, err := client.GetCommission(ctx, &settingspb.GetCommissionRequest{})
	if err != nil {
		return 0
	}
	return resp.CommissionPercent
}

func (s *OrderService) CreateOrder(ctx context.Context, req *orderpb.CreateOrderRequest) (*orderpb.Order, error) {
	customerID, err := uuid.Parse(req.CustomerId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid customer_id")
	}
	restaurantID, err := uuid.Parse(req.RestaurantId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid restaurant_id")
	}

	restaurantName := ""
	menuItemNames := make(map[string]string)
	menuItemPrices := make(map[string]float64)

	restClient, closeFn, err := s.getRestaurantClient()
	if err == nil {
		defer closeFn()
		if r, err := restClient.GetRestaurant(ctx, &restaurantpb.GetRestaurantRequest{Id: req.RestaurantId}); err == nil {
			restaurantName = r.Name
		}
		for _, item := range req.Items {
			if m, err := restClient.GetMenuItem(ctx, &restaurantpb.GetMenuItemRequest{Id: item.MenuItemId}); err == nil {
				menuItemNames[item.MenuItemId] = m.Name
				menuItemPrices[item.MenuItemId] = m.Price
			}
		}
	}

	var subtotal float64
	items := make([]*domain.OrderItem, len(req.Items))
	for i, it := range req.Items {
		menuItemID, err := uuid.Parse(it.MenuItemId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid menu_item_id")
		}
		unitPrice := menuItemPrices[it.MenuItemId]
		subtotal += unitPrice * float64(it.Quantity)
		items[i] = &domain.OrderItem{
			MenuItemID:   menuItemID,
			MenuItemName: menuItemNames[it.MenuItemId],
			Quantity:     it.Quantity,
			UnitPrice:    unitPrice,
			Options:      it.Options,
		}
	}
	commissionPercent := s.getCommissionPercent(ctx)
	total := subtotal * (1 + commissionPercent/100)

	order := &domain.Order{
		CustomerID:      customerID,
		RestaurantID:    restaurantID,
		RestaurantName:  restaurantName,
		Status:          domain.OrderStatusPending,
		Total:           total,
		DeliveryAddress: req.DeliveryAddress,
		DeliveryLat:     req.DeliveryLat,
		DeliveryLng:     req.DeliveryLng,
		Notes:           req.Notes,
	}
	if err := s.orderRepo.Create(ctx, order, items); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	order, orderItems, _ := s.orderRepo.GetByID(ctx, order.ID)
	return orderToProto(order, orderItems, ""), nil
}

func (s *OrderService) GetOrder(ctx context.Context, req *orderpb.GetOrderRequest) (*orderpb.Order, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}
	order, items, err := s.orderRepo.GetByID(ctx, id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if order == nil {
		return nil, status.Error(codes.NotFound, "order not found")
	}
	return orderToProto(order, items, ""), nil
}

func (s *OrderService) ListOrders(ctx context.Context, req *orderpb.ListOrdersRequest) (*orderpb.ListOrdersResponse, error) {
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

	var customerID, restaurantID, driverID *uuid.UUID
	if req.CustomerId != "" {
		id, err := uuid.Parse(req.CustomerId)
		if err == nil {
			customerID = &id
		}
	}
	if req.RestaurantId != "" {
		id, err := uuid.Parse(req.RestaurantId)
		if err == nil {
			restaurantID = &id
		}
	}
	if req.DriverId != "" {
		id, err := uuid.Parse(req.DriverId)
		if err == nil {
			driverID = &id
		}
	}

	orderStatus := domain.OrderStatus("")
	if req.Status != 0 {
		orderStatus = protoToStatus[req.Status]
	}

	orders, total, err := s.orderRepo.List(ctx, customerID, restaurantID, driverID, orderStatus, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	protoOrders := make([]*orderpb.Order, len(orders))
	for i, o := range orders {
		items, _ := s.orderRepo.GetItemsByOrderID(ctx, o.ID)
		protoOrders[i] = orderToProto(o, items, "")
	}

	totalPages := int64(0)
	if total > 0 && pageSize > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}

	return &orderpb.ListOrdersResponse{
		Orders: protoOrders,
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

func (s *OrderService) UpdateOrderStatus(ctx context.Context, req *orderpb.UpdateOrderStatusRequest) (*orderpb.Order, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}
	order, items, err := s.orderRepo.GetByID(ctx, id)
	if err != nil || order == nil {
		return nil, status.Error(codes.NotFound, "order not found")
	}
	if st, ok := protoToStatus[req.Status]; ok {
		order.Status = st
	} else {
		return nil, status.Error(codes.InvalidArgument, "invalid status")
	}
	if err := s.orderRepo.Update(ctx, order); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return orderToProto(order, items, ""), nil
}

func (s *OrderService) CancelOrder(ctx context.Context, req *orderpb.CancelOrderRequest) (*orderpb.Order, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}
	order, items, err := s.orderRepo.GetByID(ctx, id)
	if err != nil || order == nil {
		return nil, status.Error(codes.NotFound, "order not found")
	}
	if order.Status == domain.OrderStatusDelivered || order.Status == domain.OrderStatusCancelled {
		return nil, status.Error(codes.FailedPrecondition, "cannot cancel order in current status")
	}
	order.Status = domain.OrderStatusCancelled
	if err := s.orderRepo.Update(ctx, order); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return orderToProto(order, items, ""), nil
}

func orderToProto(o *domain.Order, items []*domain.OrderItem, driverName string) *orderpb.Order {
	if o == nil {
		return nil
	}
	protoItems := make([]*orderpb.OrderItem, len(items))
	for i, it := range items {
		protoItems[i] = &orderpb.OrderItem{
			Id:            it.ID.String(),
			MenuItemId:    it.MenuItemID.String(),
			MenuItemName:  it.MenuItemName,
			Quantity:      it.Quantity,
			UnitPrice:     it.UnitPrice,
			Options:       it.Options,
		}
	}
	return &orderpb.Order{
		Id:              o.ID.String(),
		CustomerId:      o.CustomerID.String(),
		RestaurantId:    o.RestaurantID.String(),
		RestaurantName:  o.RestaurantName,
		Status:          statusToProto[o.Status],
		Total:           o.Total,
		DeliveryAddress: o.DeliveryAddress,
		DeliveryLat:     o.DeliveryLat,
		DeliveryLng:     o.DeliveryLng,
		Notes:           o.Notes,
		DriverId:        o.DriverID.String(),
		DriverName:      driverName,
		Items:          protoItems,
		CreatedAt:      timestamppb.New(o.CreatedAt),
		UpdatedAt:      timestamppb.New(o.UpdatedAt),
	}
}
