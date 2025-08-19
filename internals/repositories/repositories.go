package repositories

import (
	"context"
	"fmt"
	"oms-service-goc/internals/models"
	"time"

	"github.com/omniful/go_commons/db/nosql/mongodm"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type OrderRepository interface {
	Create(ctx context.Context, order *models.Order) (*models.Order, error)
	FindByID(ctx context.Context, id string) (*models.Order, error)
	FindByFilters(ctx context.Context, filters OrderFilters) ([]*models.Order, error)
	UpdateStatus(ctx context.Context, id string, status models.OrderStatus) error
}

type OrderFilters struct {
	TenantID  string
	SellerID  string
	Status    string
	StartDate *time.Time
	EndDate   *time.Time
}

type orderRepository struct {
	collection *mongo.Collection
}

func NewOrderRepository(db mongodm.Database) (OrderRepository, error) {
	collection := db.GetWriteDB().Collection("orders")
	return &orderRepository{
		collection: collection,
	}, nil
}

func (r *orderRepository) Create(ctx context.Context, order *models.Order) (*models.Order, error) {
	if order.ID.IsZero() {
		order.ID = bson.NewObjectID()
	}
	order.SetCreatedAt(time.Now())
	order.SetUpdatedAt(time.Now())
	order.Initialise(ctx)
	_, err := r.collection.InsertOne(ctx, order)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}
	return order, nil
}

func (r *orderRepository) FindByID(ctx context.Context, id string) (*models.Order, error) {
	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid order ID: %w", err)
	}
	var order models.Order
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&order)
	if err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}
	return &order, nil
}

func (r *orderRepository) FindByFilters(ctx context.Context, filters OrderFilters) ([]*models.Order, error) {
	filter := bson.M{}

	if filters.TenantID != "" {
		filter["tenant_id"] = filters.TenantID
	}
	if filters.SellerID != "" {
		filter["seller_id"] = filters.SellerID
	}
	if filters.Status != "" {
		filter["status"] = filters.Status
	}

	if filters.StartDate != nil || filters.EndDate != nil {
		dateFilter := bson.M{}
		if filters.StartDate != nil {
			dateFilter["$gte"] = *filters.StartDate
		}
		if filters.EndDate != nil {
			dateFilter["$lte"] = *filters.EndDate
		}
		filter["created_at"] = dateFilter
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find orders: %w", err)
	}

	defer cursor.Close(ctx)
	var orders []*models.Order
	for cursor.Next(ctx) {
		var order models.Order
		if err := cursor.Decode(&order); err != nil {
			return nil, fmt.Errorf("failed to decode order: %w", err)
		}
		orders = append(orders, &order)
	}
	return orders, cursor.Err()
}

func (r *orderRepository) UpdateStatus(ctx context.Context, id string, status models.OrderStatus) error {
	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid order ID: %w", err)
	}

	filter := bson.M{"_id": objID}
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	_, err = r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}
	return nil
}
