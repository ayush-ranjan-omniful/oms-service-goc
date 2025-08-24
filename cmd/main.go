package main

import (
	"context"
	"log"
	"oms-service-goc/internals/configs"
	"oms-service-goc/internals/handlers/http"
	"oms-service-goc/internals/repositories"
	"oms-service-goc/internals/services"
	"oms-service-goc/routes"
	"os"

	"github.com/omniful/go_commons/db/nosql/mongodm"
	"github.com/omniful/go_commons/sqs"
)

// Mock SQS Publisher for testing - implements services.SQSPublisher interface
type MockSQSPublisher struct{}

func (m *MockSQSPublisher) Publish(ctx context.Context, message *sqs.Message) error {
	log.Printf("Mock SQS: Publishing message - GroupID: %s, DeduplicationID: %s", message.GroupId, message.DeduplicationId)
	return nil
}

func main() {
	log.Println("Starting OMS Service...")

	// Load configuration
	cfg := configs.LoadConfig()

	// Initialize MongoDB connection
	db := mongodm.NewDatabase(cfg.MongoDB)
	log.Println("Connected to MongoDB")

	// Initialize SQS publisher
	var sqsPublisher services.SQSPublisher
	env := os.Getenv("ENVIRONMENT")
	if env == "" || env == "local" {
		// Use mock for local development
		sqsPublisher = &MockSQSPublisher{}
		log.Println("Using Mock SQS Publisher for local development")
	} else {
		// For production, you would initialize real SQS client here
		// sqsClient := sqs.NewClient(cfg.SQS)
		// sqsPublisher = sqsClient
		// For now, using mock in all environments
		sqsPublisher = &MockSQSPublisher{}
		log.Println("Using Mock SQS Publisher")
	}

	// Initialize repository
	orderRepo, err := repositories.NewOrderRepository(db)
	if err != nil {
		log.Fatalf("Failed to initialize order repository: %v", err)
	}

	// Initialize service
	orderService := services.NewOrderService(orderRepo, sqsPublisher)

	// Initialize handler
	orderHandler := http.NewOrderHandler(orderService)

	// Setup routes
	router := routes.SetupRoutes(orderHandler)

	// Start server
	log.Printf("Starting server on port %s", cfg.Server.Port)
	if err := router.Run(cfg.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
