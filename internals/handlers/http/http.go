package http

import (
	"oms-service-goc/internals/models"
	"oms-service-goc/internals/services"
	"time"

	"github.com/gin-gonic/gin"
	//"github.com/omniful/go_commons/http"
	"github.com/omniful/go_commons/log"
)

type OrderHandler struct {
	orderService *services.OrderService
}

func NewOrderHandler(orderService *services.OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

// GET ORDER BY SELLER
func (h *OrderHandler) GetOrderBySeller(c *gin.Context) {
	sellerID := c.Query("seller_id")
	if sellerID == "" {
		c.JSON(400, gin.H{
			"error": "Seller_id is required",
		})
		return
	}

	orders, err := h.orderService.GetOrdersBySellerID(c.Request.Context(), sellerID)
	if err != nil {
		log.ErrorfWithContext(c.Request.Context(), "failed to get orders by seller ID %s: %v", sellerID, err)
		c.JSON(500, gin.H{
			"error": "Unable to fetch orders",
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"orders":    orders,
			"count":     len(orders),
			"seller_id": sellerID,
		},
		"timestamp": time.Now(),
	})
}

func (h *OrderHandler) GetOrderByID(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		c.JSON(404, gin.H{
			"error": "Order ID is required",
		})
		return
	}
	order, err := h.orderService.GetOrderByID(c.Request.Context(), orderID)

	if err != nil {
		log.ErrorfWithContext(c.Request.Context(), "failed to get order by ID %s: %v", orderID, err)
		c.JSON(400, gin.H{
			"error": "Unable to fetch order",
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"order":    order,
			"order_id": orderID,
		},
		"timestamp": time.Now(),
	})
}

type UpdateStatusRequest struct {
	Status models.OrderStatus `json:"status" binding:"required"`
}

func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	orderID := c.Param("id")

	if orderID == "" {
		c.JSON(404, gin.H{
			"error": "Order ID is required",
		})
		return
	}

	var updateOrderRequest UpdateStatusRequest

	if err := c.ShouldBindJSON(&updateOrderRequest); err != nil {
		log.ErrorfWithContext(c.Request.Context(), "Invalid request format: %v", err)
		c.JSON(400, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	err := h.orderService.UpdateOrderStatus(c.Request.Context(), orderID, updateOrderRequest.Status)
	if err != nil {
		log.ErrorfWithContext(c.Request.Context(), "failed to update order %w", err)
		c.JSON(500, gin.H{
			"error": "Failed to update order status",
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "Order status updated successfully",
		"data": gin.H{
			"status": updateOrderRequest.Status,
		},
		"timestamp": time.Now(),
	})
}

func (h *OrderHandler) CreateBulkOrder(c *gin.Context) {

	var request models.BulkOrderRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		log.ErrorfWithContext(c.Request.Context(), "Invalid request format : %v", err)
		c.JSON(400, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	err := h.orderService.CreateBulkOrder(c.Request.Context(), &request)
	if err != nil {
		log.ErrorfWithContext(c.Request.Context(), "failed to create bulk order: %v", err)
		c.JSON(500, gin.H{
			"error": "Failed to create bulk order",
		})
		return
	}

	c.JSON(202, gin.H{
		"success": true,
		"message": "Bulk order request queued successfully",
		"data": gin.H{
			"file_path": request.FilePath,
			"user_id":   request.UserID,
		},
		"timestamp": time.Now(),
	})
}
