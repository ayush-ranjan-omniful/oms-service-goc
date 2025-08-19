package models

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type OrderStatus string

const (
	OrderStatusOnHold   OrderStatus = "on_hold"
	OrderStatusNewOrder OrderStatus = "new_order"
)

type Order struct {
	ID        bson.ObjectID `bson:"_id,omitempty"  json:"id,omitempty"`
	TenantID  string        `bson:"tenant_id" json:"tenant_id"`
	SellerID  string        `bson:"seller_id" json:"seller_id"`
	HubID     string        `bson:"hub_id" json:"hub_id"`
	Status    OrderStatus   `bson:"status" json:"status"`
	Items     []OrderItem   `bson:"items,omitempty" json:"items,omitempty"`
	CreatedAt time.Time     `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time     `bson:"updated_at" json:"updated_at"`
}

type OrderItem struct {
	SKUCode  string `bson:"sku_code" json:"sku_code"`
	Quantity int    `bson:"quantity" json:"quantity"`
}

func (o *Order) GetID() bson.ObjectID     { return o.ID }
func (o *Order) SetID(id bson.ObjectID)   { o.ID = id }
func (o *Order) SetCreatedAt(t time.Time) { o.CreatedAt = t }
func (o *Order) SetUpdatedAt(t time.Time) { o.UpdatedAt = t }

func (o *Order) Initialise(ctx context.Context) {
	if o.Status == "" {
		o.Status = OrderStatusOnHold
	}
}

func (o *Order) Validate(ctx context.Context) error {
	return nil
}

// // Scoping methods for Tenant
// func (o *Order) GetTenantID() string         { return o.TenantID }
// func (o *Order) GetTenantKey() string        { return "tenant_id" }
// func (o *Order) SetTenantID(tenantID string) { o.TenantID = tenantID }

// // Scoping methods for Seller
// func (o *Order) GetSellerID() string          { return o.SellerID }
// func (o *Order) GetSellerKey() string         { return "seller_id" }
// func (o *Order) SetSellerKey(sellerID string) { o.SellerID = sellerID }

// // Scoping methods for Hub

// func (o *Order) GetHubKey() string      { return "hub_id" }
// func (o *Order) GetHubID() string       { return o.HubID }
// func (o *Order) SetHubKey(hubID string) { o.HubID = hubID }

// // Scoping interface implementations
// func (o *Order) IsTenantScoped() bool { return true }
// func (o *Order) IsSellerScoped() bool { return true }
// func (o *Order) IsHubScoped() bool    { return true }

// CSV Models

type OrderCSVRow struct {
	TenantID string `csv:"tenant_id"`
	SellerID string `csv:"seller_id"`
	HubID    string `csv:"hub_id"`
	SKUCode  string `csv:"sku_code"`
	Quantity int    `csv:"quantity"`
}

type BulkOrderRequest struct {
	FilePath string `json:"file_path"`
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
}

// SQS Event Models - filepath, Userid, username
type CreateBulkOrderEvent struct {
	FilePath string `json:"file_path"`
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
}

// Kafka Event models - orderid, tenantid, sellerid, hubid, items, createdat
type OrderCreatedEvent struct {
	OrderID   string      `json:"order_id"`
	TenantID  string      `json:"tenant_id"`
	SellerID  string      `json:"seller_id"`
	HubId     string      `json:"hub_id"`
	Items     []OrderItem `json:"items"`
	CreatedAt time.Time   `json:"created_at"`
}
