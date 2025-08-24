# OMS Service - Order Management Service

A microservice built with GoCommons framework for managing orders with SQS integration.

## Project Structure

```
├── cmd/
│   └── main.go                 # Application entry point
├── internals/
│   ├── configs/                # Configuration management
│   ├── handlers/http/          # HTTP API handlers  
│   ├── models/                 # Data models and entities
│   ├── repositories/           # Database access layer
│   └── services/               # Business logic layer
├── routes/                     # HTTP route definitions
├── go.mod                      # Go module dependencies
└── bin/                        # Compiled binaries
```

## Features

- **Order Management**: Create, read, update orders
- **Bulk Order Processing**: Process multiple orders via SQS
- **MongoDB Integration**: Using GoCommons mongodm for data persistence
- **SQS Integration**: Asynchronous message processing (with LocalStack support)
- **Clean Architecture**: Repository pattern with dependency injection
- **Comprehensive Testing**: Unit tests for all layers

## API Endpoints

### Health Check
- `GET /health` - Service health status

### Orders
- `GET /api/v1/orders?seller_id={id}` - Get orders by seller ID
- `GET /api/v1/orders/{id}` - Get order by ID
- `PUT /api/v1/orders/{id}/status` - Update order status
- `POST /api/v1/orders/bulk` - Create bulk orders (queues via SQS)

### Running the Service

1. **Start MongoDB:**
   ```bash
   docker run --rm -d -p 27017:27017 --name mongodb mongo:latest
   ```

2. **Start LocalStack (optional, for SQS testing):**
   ```bash
   docker run --rm -d -p 4566:4566 --name localstack localstack/localstack
   ```

3. **Build and run the service:**
   ```bash
   go build -o bin/oms cmd/main.go
   ./bin/oms
   ```

The service will start on port `:8080`

### Testing

**Run all tests:**
```bash
go test ./... -v
```

**Test individual components:**
```bash
# Repository tests
go test ./internals/repositories -v

# Service tests  
go test ./internals/services -v
```

### API Calls

**Health check:**
```bash
curl http://localhost:8080/health
```

**Get orders by seller:**
```bash
curl "http://localhost:8080/api/v1/orders?seller_id=seller123"
```

**Create bulk orders:**
```bash
curl -X POST http://localhost:8080/api/v1/orders/bulk \
  -H "Content-Type: application/json" \
  -d '{
    "seller_id": "seller123",
    "orders": [
      {
        "customer_name": "John Doe",
        "customer_email": "john@example.com", 
        "total_amount": 100.50,
        "items": [
          {
            "product_id": "prod1",
            "product_name": "Widget A",
            "quantity": 2,
            "unit_price": 50.25
          }
        ]
      }
    ]
  }'
```

## Configuration

The service supports both local and production environments:

- **Local**: Uses MongoDB at `localhost:27017` and LocalStack SQS at `localhost:4566`
- **Production**: Uses environment variables for configuration

### Environment Variables (Production)
- `ENVIRONMENT`: Set to "production" for production mode
- `PORT`: Server port (default: ":8080")
- `MONGODB_URI`: MongoDB connection string
- `MONGODB_DATABASE`: MongoDB database name
- `SQS_ACCOUNT`: AWS account ID
- `SQS_REGION`: AWS region
- `SQS_ENDPOINT`: SQS endpoint URL

