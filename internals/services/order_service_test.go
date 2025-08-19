package services

import (
	"context"
	"encoding/json"
	"fmt"
	"oms-service-goc/internals/models"
	"oms-service-goc/internals/repositories"
	"testing"
	"time"

	"github.com/omniful/go_commons/db/nosql/mongodm"
	"github.com/omniful/go_commons/sqs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// Mock SQS Publisher - implements the SQSPublisher interface
type MockSQSPublisher struct {
	mock.Mock
}

func (m *MockSQSPublisher) Publish(ctx context.Context, message *sqs.Message) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

// Wrapper for the real SQS Publisher to implement our interface
type RealSQSPublisher struct {
	publisher *sqs.Publisher
}

func (r *RealSQSPublisher) Publish(ctx context.Context, message *sqs.Message) error {
	return r.publisher.Publish(ctx, message)
}

// Test setup
func setupOrderServiceTest(t *testing.T) (*OrderService, repositories.OrderRepository, *MockSQSPublisher) {
	// Setup test database
	cfg := mongodm.Config{
		Database:        "test_oms_service",
		URI:             "mongodb://localhost:27017",
		ReadPreference:  mongodm.ReadPrefPrimary,
		DefaultTimeout:  5 * time.Second,
		MaxPoolSize:     5,
		MinPoolSize:     1,
		MaxConnIdleTime: 1 * time.Minute,
	}

	db := mongodm.NewDatabase(cfg)
	repo, err := repositories.NewOrderRepository(db)
	require.NoError(t, err)

	// Create mock SQS publisher
	mockSQS := &MockSQSPublisher{}

	// Create service - now mockSQS implements SQSPublisher interface
	service := NewOrderService(repo, mockSQS)

	return service, repo, mockSQS
}

// Add cleanup helper
func cleanupTestData(t *testing.T, db mongodm.Database) {
	err := db.GetWriteDB().Collection("orders").Drop(context.Background())
	if err != nil {
		t.Logf("Warning: failed to cleanup test data: %v", err)
	}
}

func TestOrderService_GetOrdersBySellerId(t *testing.T) {
	service, repo, _ := setupOrderServiceTest(t)

	// Create test orders
	testOrders := []*models.Order{
		{
			TenantID: "tenant1",
			SellerID: "seller123",
			HubID:    "hub1",
			Status:   models.OrderStatusOnHold,
			Items: []models.OrderItem{
				{SKUCode: "SKU001", Quantity: 10},
			},
		},
		{
			TenantID: "tenant1",
			SellerID: "seller123",
			HubID:    "hub1",
			Status:   models.OrderStatusNewOrder,
			Items: []models.OrderItem{
				{SKUCode: "SKU002", Quantity: 5},
			},
		},
		{
			TenantID: "tenant1",
			SellerID: "different_seller",
			HubID:    "hub1",
			Status:   models.OrderStatusOnHold,
			Items: []models.OrderItem{
				{SKUCode: "SKU003", Quantity: 15},
			},
		},
	}

	// Insert test data
	for _, order := range testOrders {
		_, err := repo.Create(context.Background(), order)
		require.NoError(t, err)
	}

	// Test the service method
	result, err := service.GetOrdersBySellerId(context.Background(), "seller123")

	// Assertions
	assert.NoError(t, err)
	assert.Len(t, result, 2) // Should return only orders for seller123

	for _, order := range result {
		assert.Equal(t, "seller123", order.SellerID)
	}

	// Cleanup
	// t.Cleanup(func() {
	//     cleanupTestData(t, service.orderRepo.(*repositories.orderReporository).collection.Database())
	// })
}

func TestOrderService_CreateBulkOrder(t *testing.T) {
	service, _, mockSQS := setupOrderServiceTest(t)

	// Setup mock expectations
	mockSQS.On("Publish", mock.Anything, mock.MatchedBy(func(msg *sqs.Message) bool {
		// Verify the message structure
		assert.Equal(t, "bulk-orders", msg.GroupId)
		assert.Contains(t, msg.DeduplicationId, "bulk-")

		// Verify the message content
		var request models.BulkOrderRequest
		err := json.Unmarshal(msg.Value, &request)
		assert.NoError(t, err)
		assert.Equal(t, "/test/orders.csv", request.FilePath)
		assert.Equal(t, "testuser", request.UserName)

		return true
	})).Return(nil)

	// Test data
	request := &models.BulkOrderRequest{
		FilePath: "/test/orders.csv",
		UserID:   "user123",
		UserName: "testuser",
	}

	// Test the service method
	err := service.CreateBulkOrder(context.Background(), request)

	// Assertions
	assert.NoError(t, err)
	mockSQS.AssertExpectations(t)
}

func TestOrderService_GetOrderByID(t *testing.T) {
	service, repo, _ := setupOrderServiceTest(t)

	// Create test order
	testOrder := &models.Order{
		TenantID: "tenant1",
		SellerID: "seller123",
		HubID:    "hub1",
		Status:   models.OrderStatusOnHold,
		Items: []models.OrderItem{
			{SKUCode: "SKU001", Quantity: 10},
		},
	}

	created, err := repo.Create(context.Background(), testOrder)
	require.NoError(t, err)

	// Test the service method
	result, err := service.GetOrderByID(context.Background(), created.ID.Hex())

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, created.ID, result.ID)
	assert.Equal(t, created.SellerID, result.SellerID)
	assert.Equal(t, created.Status, result.Status)
}

func TestOrderService_UpdateOrderStatus(t *testing.T) {
	service, repo, _ := setupOrderServiceTest(t)

	// Create test order
	testOrder := &models.Order{
		TenantID: "tenant1",
		SellerID: "seller123",
		HubID:    "hub1",
		Status:   models.OrderStatusOnHold,
		Items: []models.OrderItem{
			{SKUCode: "SKU001", Quantity: 10},
		},
	}

	created, err := repo.Create(context.Background(), testOrder)
	require.NoError(t, err)

	// Test the service method
	err = service.UpdateOrderStatus(context.Background(), created.ID.Hex(), models.OrderStatusNewOrder)
	assert.NoError(t, err)

	// Verify the update
	updated, err := repo.FindByID(context.Background(), created.ID.Hex())
	require.NoError(t, err)
	assert.Equal(t, models.OrderStatusNewOrder, updated.Status)
}

func TestOrderService_GetOrdersBySellerId_EmptyResult(t *testing.T) {
	service, _, _ := setupOrderServiceTest(t)

	// Test with non-existent seller
	result, err := service.GetOrdersBySellerId(context.Background(), "nonexistent_seller")

	// Should return empty slice, not error
	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestOrderService_GetOrderByID_NotFound(t *testing.T) {
	service, _, _ := setupOrderServiceTest(t)

	// Test with non-existent order ID
	_, err := service.GetOrderByID(context.Background(), bson.NewObjectID().Hex())

	// Should return error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "order not found")
}

// Add test for SQS publishing failure
func TestOrderService_CreateBulkOrder_SQSFailure(t *testing.T) {
	service, _, mockSQS := setupOrderServiceTest(t)

	// Setup mock to return error
	mockSQS.On("Publish", mock.Anything, mock.Anything).Return(fmt.Errorf("SQS connection failed"))

	request := &models.BulkOrderRequest{
		FilePath: "/test/orders.csv",
		UserID:   "user123",
		UserName: "testuser",
	}

	// Test the service method
	err := service.CreateBulkOrder(context.Background(), request)

	// Should return error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to queue bulk order")
	mockSQS.AssertExpectations(t)
}
