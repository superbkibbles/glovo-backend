package server

import (
	"context"

	"github.com/mendmzury/food-delivery/services/order-service/internal/application"
	orderpb "github.com/mendmzury/food-delivery/proto/order"
)

type OrderServer struct {
	orderpb.UnimplementedOrderServiceServer
	service *application.OrderService
}

func NewOrderServer(service *application.OrderService) *OrderServer {
	return &OrderServer{service: service}
}

func (s *OrderServer) CreateOrder(ctx context.Context, req *orderpb.CreateOrderRequest) (*orderpb.Order, error) {
	return s.service.CreateOrder(ctx, req)
}

func (s *OrderServer) GetOrder(ctx context.Context, req *orderpb.GetOrderRequest) (*orderpb.Order, error) {
	return s.service.GetOrder(ctx, req)
}

func (s *OrderServer) ListOrders(ctx context.Context, req *orderpb.ListOrdersRequest) (*orderpb.ListOrdersResponse, error) {
	return s.service.ListOrders(ctx, req)
}

func (s *OrderServer) UpdateOrderStatus(ctx context.Context, req *orderpb.UpdateOrderStatusRequest) (*orderpb.Order, error) {
	return s.service.UpdateOrderStatus(ctx, req)
}

func (s *OrderServer) CancelOrder(ctx context.Context, req *orderpb.CancelOrderRequest) (*orderpb.Order, error) {
	return s.service.CancelOrder(ctx, req)
}

func (s *OrderServer) AcceptOrder(ctx context.Context, req *orderpb.AcceptOrderRequest) (*orderpb.Order, error) {
	return s.service.AcceptOrder(ctx, req)
}
