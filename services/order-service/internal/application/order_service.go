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
	userpb "github.com/mendmzury/food-delivery/proto/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var statusToProto = map[domain.OrderStatus]orderpb.OrderStatus{
	domain.OrderStatusPending:   orderpb.OrderStatus_ORDER_STATUS_PENDING,
	domain.OrderStatusAccepted:  orderpb.OrderStatus_ORDER_STATUS_ACCEPTED,
	domain.OrderStatusPreparing: orderpb.OrderStatus_ORDER_STATUS_PREPARING,
	domain.OrderStatusReady:     orderpb.OrderStatus_ORDER_STATUS_READY,
	domain.OrderStatusAssigned:  orderpb.OrderStatus_ORDER_STATUS_ASSIGNED,
	domain.OrderStatusOnTheWay:  orderpb.OrderStatus_ORDER_STATUS_ON_THE_WAY,
	domain.OrderStatusPickedUp:  orderpb.OrderStatus_ORDER_STATUS_PICKED_UP,
	domain.OrderStatusDelivered: orderpb.OrderStatus_ORDER_STATUS_DELIVERED,
	domain.OrderStatusCancelled: orderpb.OrderStatus_ORDER_STATUS_CANCELLED,
}

var protoToStatus = map[orderpb.OrderStatus]domain.OrderStatus{
	orderpb.OrderStatus_ORDER_STATUS_PENDING:    domain.OrderStatusPending,
	orderpb.OrderStatus_ORDER_STATUS_ACCEPTED:   domain.OrderStatusAccepted,
	orderpb.OrderStatus_ORDER_STATUS_PREPARING: domain.OrderStatusPreparing,
	orderpb.OrderStatus_ORDER_STATUS_READY:     domain.OrderStatusReady,
	orderpb.OrderStatus_ORDER_STATUS_ASSIGNED:  domain.OrderStatusAssigned,
	orderpb.OrderStatus_ORDER_STATUS_ON_THE_WAY: domain.OrderStatusOnTheWay,
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

func (s *OrderService) getUserClient() (userpb.UserServiceClient, func() error, error) {
	conn, err := grpc.NewClient(s.cfg.UserServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	return userpb.NewUserServiceClient(conn), conn.Close, nil
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
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to connect to restaurant service")
	}
	defer closeFn()
	if r, err := restClient.GetRestaurant(ctx, &restaurantpb.GetRestaurantRequest{Id: req.RestaurantId}); err == nil {
		restaurantName = r.Name
	}
	for _, item := range req.Items {
		m, err := restClient.GetMenuItem(ctx, &restaurantpb.GetMenuItemRequest{Id: item.MenuItemId})
		if err != nil {
			return nil, status.Error(codes.NotFound, "menu item not found: "+item.MenuItemId)
		}
		menuItemNames[item.MenuItemId] = m.Name
		menuItemPrices[item.MenuItemId] = m.Price
	}

	var subtotal float64
	items := make([]*domain.OrderItem, len(req.Items))
	for i, it := range req.Items {
		menuItemID, err := uuid.Parse(it.MenuItemId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid menu_item_id")
		}
		unitPrice := menuItemPrices[it.MenuItemId]
		if unitPrice <= 0 {
			return nil, status.Error(codes.InvalidArgument, "invalid price for menu item: "+it.MenuItemId)
		}
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

	order, orderItems, err := s.orderRepo.GetByID(ctx, order.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to retrieve created order: "+err.Error())
	}
	if order == nil {
		return nil, status.Error(codes.Internal, "failed to retrieve created order")
	}
	return s.orderToProto(ctx, order, orderItems, ""), nil
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
	return s.orderToProto(ctx, order, items, ""), nil
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

	var statuses []domain.OrderStatus
	if len(req.Statuses) > 0 {
		for _, st := range req.Statuses {
			if domainSt, ok := protoToStatus[st]; ok {
				statuses = append(statuses, domainSt)
			}
		}
	} else if req.Status != orderpb.OrderStatus_ORDER_STATUS_PENDING {
		if st, ok := protoToStatus[req.Status]; ok {
			statuses = append(statuses, st)
		}
	}

	orders, total, err := s.orderRepo.List(ctx, customerID, restaurantID, driverID, statuses, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	protoOrders := make([]*orderpb.Order, len(orders))
	for i, o := range orders {
		items, _ := s.orderRepo.GetItemsByOrderID(ctx, o.ID)
		protoOrders[i] = s.orderToProto(ctx, o, items, "")
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
	// Accept driver_id via gRPC metadata (set by delivery-service on assignment)
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if driverIDs := md.Get("driver-id"); len(driverIDs) > 0 {
			if driverID, parseErr := uuid.Parse(driverIDs[0]); parseErr == nil {
				order.DriverID = driverID
			}
		}
	}
	if err := s.orderRepo.Update(ctx, order); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return s.orderToProto(ctx, order, items, ""), nil
}

func (s *OrderService) AcceptOrder(ctx context.Context, req *orderpb.AcceptOrderRequest) (*orderpb.Order, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}
	var callerDriver uuid.UUID
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if v := md.Get("driver-id"); len(v) > 0 {
			if d, perr := uuid.Parse(v[0]); perr == nil {
				callerDriver = d
			}
		}
	}
	if callerDriver == uuid.Nil {
		return nil, status.Error(codes.Unauthenticated, "driver-id required")
	}
	order, items, err := s.orderRepo.GetByID(ctx, id)
	if err != nil || order == nil {
		return nil, status.Error(codes.NotFound, "order not found")
	}
	if order.Status != domain.OrderStatusAssigned {
		return nil, status.Error(codes.FailedPrecondition, "order must be assigned to accept")
	}
	if order.DriverID == uuid.Nil || order.DriverID != callerDriver {
		return nil, status.Error(codes.PermissionDenied, "not the assigned driver for this order")
	}
	order.Status = domain.OrderStatusOnTheWay
	if err := s.orderRepo.Update(ctx, order); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return s.orderToProto(ctx, order, items, ""), nil
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
	return s.orderToProto(ctx, order, items, ""), nil
}

func (s *OrderService) orderToProto(ctx context.Context, o *domain.Order, items []*domain.OrderItem, driverName string) *orderpb.Order {
	if o == nil {
		return nil
	}
	protoItems := make([]*orderpb.OrderItem, len(items))
	for i, it := range items {
		protoItems[i] = &orderpb.OrderItem{
			Id:           it.ID.String(),
			MenuItemId:   it.MenuItemID.String(),
			MenuItemName: it.MenuItemName,
			Quantity:     it.Quantity,
			UnitPrice:    it.UnitPrice,
			Options:      it.Options,
		}
	}

	customerName := ""
	userCl, closeUser, err := s.getUserClient()
	if err == nil {
		defer closeUser()
		if u, err := userCl.GetUser(ctx, &userpb.GetUserRequest{Id: o.CustomerID.String()}); err == nil && u != nil {
			customerName = u.Name
		}
		if driverName == "" && o.DriverID != uuid.Nil {
			if d, err := userCl.GetUser(ctx, &userpb.GetUserRequest{Id: o.DriverID.String()}); err == nil && d != nil {
				driverName = d.Name
			}
		}
	}

	pickupAddr := ""
	var pickupLat, pickupLng float64
	restaurantName := o.RestaurantName
	restCl, closeRest, err := s.getRestaurantClient()
	if err == nil {
		defer closeRest()
		if r, err := restCl.GetRestaurant(ctx, &restaurantpb.GetRestaurantRequest{Id: o.RestaurantID.String()}); err == nil && r != nil {
			pickupAddr = r.Address
			pickupLat = r.Lat
			pickupLng = r.Lng
			if restaurantName == "" {
				restaurantName = r.Name
			}
		}
	}

	st, ok := statusToProto[o.Status]
	if !ok {
		st = orderpb.OrderStatus_ORDER_STATUS_PENDING
	}

	driverIDStr := ""
	if o.DriverID != uuid.Nil {
		driverIDStr = o.DriverID.String()
	}

	return &orderpb.Order{
		Id:                    o.ID.String(),
		CustomerId:            o.CustomerID.String(),
		CustomerName:          customerName,
		RestaurantId:          o.RestaurantID.String(),
		RestaurantName:        restaurantName,
		Status:                st,
		Total:                 o.Total,
		DeliveryAddress:       o.DeliveryAddress,
		DeliveryLat:           o.DeliveryLat,
		DeliveryLng:           o.DeliveryLng,
		Notes:                 o.Notes,
		DriverId:              driverIDStr,
		DriverName:            driverName,
		Items:                 protoItems,
		CreatedAt:             timestamppb.New(o.CreatedAt),
		UpdatedAt:             timestamppb.New(o.UpdatedAt),
		PickupAddress:         pickupAddr,
		PickupLatitude:        pickupLat,
		PickupLongitude:       pickupLng,
		DestinationLatitude:   o.DeliveryLat,
		DestinationLongitude:  o.DeliveryLng,
	}
}
