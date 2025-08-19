package services

import (
	"context"
	"encoding/json"
	"fmt"
	"oms-service-goc/internals/models"
	"oms-service-goc/internals/repositories"

	"github.com/omniful/go_commons/log"
	"github.com/omniful/go_commons/sqs"
)

type SQSPublisher interface {
	Publish(ctx context.Context, message *sqs.Message) error
}

type OrderService struct {
	orderRepo    repositories.OrderRepository
	sqsPublisher SQSPublisher
}

func NewOrderService(orderRepo repositories.OrderRepository, sqsPublisher SQSPublisher) *OrderService {
	return &OrderService{
		orderRepo:    orderRepo,
		sqsPublisher: sqsPublisher,
	}
}

func (s *OrderService) GetOrdersBySellerId(ctx context.Context, sellerID string) ([]*models.Order, error) {
	filters := repositories.OrderFilters{
		SellerID: sellerID,
	}
	orders, err := s.orderRepo.FindByFilters(ctx, filters)
	if err != nil {
		log.ErrorfWithContext(ctx, "failed to get orders by seller ID %s: %v", sellerID, err)
		return nil, fmt.Errorf("unable to fetch orders : %w", err)
	}
	return orders, nil
}

// CREATE BULK ORDER
func (s *OrderService) CreateBulkOrder(ctx context.Context, request *models.BulkOrderRequest) error {
	eventData, err := json.Marshal(request)
	if err != nil {
		log.ErrorfWithContext(ctx, "failed to marshal bulk order request: %w", err)
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	message := &sqs.Message{
		GroupId:         "bulk-orders",
		Value:           eventData,
		DeduplicationId: fmt.Sprintf("bulk-%s-%s", request.UserName, request.FilePath),
	}

	err = s.sqsPublisher.Publish(ctx, message)
	if err != nil {
		log.ErrorfWithContext(ctx, "failed to publish bulk order message: %v", err)
		return fmt.Errorf("failed to queue bulk order message: %w", err)
	}

	log.Info("Bulk order request queued successfully")
	return nil
}

// GET ORDER BY ID
func (s *OrderService) GetOrderByID(ctx context.Context, orderID string) (*models.Order, error) {
	order, err := s.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		log.ErrorfWithContext(ctx, "failed to get order %s: %v", orderID, err)
		return nil, fmt.Errorf("order not found : %w", err)
	}
	return order, nil
}

func (s *OrderService) UpdateOrderStatus(ctx context.Context, orderID string, status models.OrderStatus) error {
	err := s.orderRepo.UpdateStatus(ctx, orderID, status)

	if err != nil {
		log.ErrorfWithContext(ctx, "failed to update order status for %s: %v", orderID, err)
		return fmt.Errorf("failed to update order status: %w", err)
	}

	log.Info("Order %s status updated to %s", orderID, status)
	return nil
}
