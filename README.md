# Adrian Janczenia - Gateway Service

The **Gateway Service** acts as the central entry point for the portfolio microservices ecosystem. It is responsible for aggregating requests from the front-end, orchestrating communication between internal services, and performing protocol translation.

## Service Role

This service serves as a facade, providing a standardized REST API. Its primary responsibilities include:

- **Protocol Translation**: Converting external REST HTTP requests into internal gRPC or RabbitMQ RPC calls.
- **Data Aggregation**: Fetching and merging localized content from the Content Service.
- **Error Normalization**: Advanced mapping of internal errors to correct HTTP statuses (e.g., 400 for invalid Captcha, 410 for expired session) instead of generic 500 errors.
- **Secure Proxying**: Acting as a secure intermediary for streaming PDF files with strict header verification.

## Architecture

The project follows a rigorous layered architecture to ensure high testability, scalability, and clear separation of concerns.

### Layered Pattern: Handler -> Process -> Task
1. Handler (Transport Layer): Handles HTTP-specific logic, input validation, and mapping internal domain errors to standardized JSON responses (slugs).
2. Process (Orchestration Layer): Contains the core business logic, managing the flow of data between different tasks and services.
3. Task / Service (Action Layer): Atomic technical operations, such as calling a gRPC procedure, publishing a message to a RabbitMQ queue, or performing internal HTTP requests.

## Technical Specification

- Go: 1.23+
- gRPC: High-performance synchronous communication for content retrieval.
- RabbitMQ: Asynchronous message-driven architecture for decoupled CV token requests.
- HTTP Client: Enhanced Captcha Service client with slug-based error handling.

## API Documentation

| Endpoint | Method | Description |
|----------|--------|-------------|
| /api/v1/content | GET | Retrieves localized page content. |
| /api/v1/cv-request | POST | Authenticates user and requests a CV download token via RabbitMQ. |
| /api/v1/download/cv | GET | Securely streams the CV PDF file. |
| /api/v1/pow | GET | Fetches a seed for the Proof of Work mechanism. |
| /api/v1/captcha | POST | Generates a Captcha image based on solved PoW. |
| /api/v1/captcha-verify | POST | Verifies the provided Captcha code. |

## Development and Deployment

### Build Optimized Docker Image
docker build -t gateway-service .

### Execute Unit Tests
go test -v ./...

## Error Handling Design

The system implements a unified error mapping mechanism using a custom AppError type. This ensures that every failure returned to the front-end contains a consistent HTTP status and a machine-readable slug (e.g., error_cv_auth), enabling precise error handling on the client side.

---
Adrian Janczenia