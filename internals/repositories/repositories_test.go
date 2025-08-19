package repositories

import (
	"context"
	"oms-service-goc/internals/models"
	"testing"
	"time"

	"github.com/omniful/go_commons/db/nosql/mongodm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB() mongodm.Database {
	cfg := mongodm.Config{
		Database:        "test_oms",
		URI:             "mongodb://localhost:27017",
		ReadPreference:  mongodm.ReadPrefPrimary,
		DefaultTimeout:  5 * time.Second,
		MaxPoolSize:     5,
		MinPoolSize:     1,
		MaxConnIdleTime: 1 * time.Minute,
	}
	return mongodm.NewDatabase(cfg)
}

func TestOrderRepository(t *testing.T) {
	db := setupTestDB()
	defer db.Client().Disconnect(context.Background())
	repo, err := NewOrderRepository(db)
	require.NoError(t, err)

	order := &models.Order{
		TenantID: "tenant1",
		SellerID: "seller1",
		HubID:    "hub1",
		Status:   models.OrderStatusOnHold,
		Items: []models.OrderItem{
			{SKUCode: "SKU001", Quantity: 10},
		},
	}

	created, err := repo.Create(context.Background(), order)

	assert.NoError(t, err)
	assert.NotEmpty(t, created.ID)
	assert.Equal(t, models.OrderStatusOnHold, created.Status)

	assert.NotZero(t, created.CreatedAt)
}

func TestOrderRepository_FindByFilter(t *testing.T) {
	db := setupTestDB()
	defer db.Client().Disconnect(context.Background())

	repo, err := NewOrderRepository(db)
	require.NoError(t, err)

	order := &models.Order{
		TenantID: "tenant1",
		SellerID: "seller1",
		HubID:    "hub1",
		Status:   models.OrderStatusNewOrder,
	}

	created, err := repo.Create(context.Background(), order)
	assert.NoError(t, err)

	// Test Fltering
	filters := OrderFilters{
		TenantID: "tenant1",
		Status:   string(models.OrderStatusNewOrder),
	}

	orders, err := repo.FindByFilters(context.Background(), filters)

	assert.NoError(t, err)
	assert.Len(t, orders, 1)
	assert.Equal(t, created.ID, orders[0].ID)
}
