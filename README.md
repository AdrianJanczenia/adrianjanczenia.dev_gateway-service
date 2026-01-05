# Adrian Janczenia - Gateway Service

The **Gateway Service** acts as the central entry point (API Gateway) for the portfolio microservices ecosystem. It is responsible for aggregating requests from the front-end, orchestrating communication between internal services, and performing protocol translation.

## Service Role

This service serves as a facade, providing a standardized REST API. Its primary responsibilities include:

- **Protocol Translation**: Converting external REST HTTP requests into internal gRPC or RabbitMQ RPC calls.
- **Data Aggregation**: Fetching and merging localized content from the Content Service.
- **Secure Proxying**: Acting as a secure intermediary for streaming PDF files with strict header verification (Allow-List).
- **Context Propagation**: Ensuring full context.Context propagation to handle distributed operation cancellation and timeouts correctly.

## Architecture

The project follows a rigorous layered architecture to ensure high testability, scalability, and clear separation of concerns.

### Layered Pattern: Handler -> Process -> Task

1. Handler (Transport Layer): Handles HTTP-specific logic. It is responsible for input validation, query parameter extraction, and mapping internal domain errors to standardized JSON responses.
2. Process (Orchestration Layer): Contains the core business logic. It manages the flow of data between different tasks and services without being aware of the underlying transport protocol.
3. Task / Service (Action Layer): Atomic technical operations, such as calling a gRPC procedure, publishing a message to a RabbitMQ queue, or performing specialized data mapping.

## Technical Specification

- Go: 1.23+ (utilizing structured logging and enhanced context features).
- gRPC: High-performance, low-latency synchronous communication for content retrieval.
- RabbitMQ: Asynchronous message-driven architecture for decoupled CV token requests.
- Docker: Optimized containerization using multi-stage builds on Alpine Linux for minimal image size and high security.

## API Documentation

| Endpoint | Method | Description |
|----------|--------|-------------|
| /api/v1/content | GET | Retrieves localized page content based on language query parameters. |
| /api/v1/cv-request | POST | Authenticates user and requests a temporary CV download token via MQ. |
| /api/v1/download/cv | GET | Securely streams the CV PDF file while stripping sensitive server headers. |

## Environment Configuration

The service utilizes a configuration system based on YAML files and environment variables, following the fail-fast principle.

| Variable | Description | Default |
|----------|-------------|---------|
| APP_ENV | Runtime environment (local/production) | local |
| CONTENT_SERVICE_GRPC_ADDR | Address of the Content Service gRPC server | localhost:50051 |
| CONTENT_SERVICE_HTTP_ADDR | Base URL for Content Service HTTP API | http://localhost:8081 |
| RABBITMQ_URL | Connection string for RabbitMQ broker | amqp://... |

## Development and Deployment

### Build Optimized Docker Image
docker build -t gateway-service .

### Execute Unit Tests
go test -v ./...

## Error Handling Design

The system implements a unified error mapping mechanism using a custom AppError type. This ensures that every failure returned to the front-end contains a consistent HTTP status and a machine-readable slug (e.g., error_cv_auth), enabling precise error handling on the client side.

---
Adrian Janczenia