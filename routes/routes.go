package routes

import (
	"oms-service-goc/internals/handlers/http"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(orderHandler *http.OrderHandler) *gin.Engine {
	router := gin.Default()

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "oms-service",
		})
	})

	// API v1 group
	v1 := router.Group("/api/v1")
	{
		orders := v1.Group("/orders")
		{
			orders.GET("", orderHandler.GetOrderBySeller)             // GET /api/v1/orders?seller_id=xxx
			orders.GET("/:id", orderHandler.GetOrderByID)             // GET /api/v1/orders/{id}
			orders.PUT("/:id/status", orderHandler.UpdateOrderStatus) // PUT /api/v1/orders/{id}/status
			orders.POST("/bulk", orderHandler.CreateBulkOrder)        // POST /api/v1/orders/bulk
		}
	}

	return router
}
