# Gateway service

The gateway exposes an HTTP API to the front end and connects it to the content and CAPTCHA services. It retrieves localized content over gRPC, requests CV download tokens through RabbitMQ, and forwards CAPTCHA and PDF requests over HTTP. Handlers translate transport data and errors; processes coordinate each use case; service adapters own the protocol-specific code.

## HTTP API

| Method | Path | Downstream operation |
| --- | --- | --- |
| `GET` | `/api/v1/content` | Read localized content from the content service over gRPC. |
| `POST` | `/api/v1/cv-request` | Request a CV token from the content service through RabbitMQ. |
| `GET` | `/api/v1/download/cv` | Stream a PDF from the content service over HTTP. |
| `GET` | `/api/v1/pow` | Request a signed PoW challenge from the CAPTCHA service. |
| `POST` | `/api/v1/captcha` | Submit PoW and request an image CAPTCHA. |
| `POST` | `/api/v1/captcha-verify` | Verify a CAPTCHA answer. |

Responses use HTTP status codes and distinct machine-readable error slugs. The gateway validates each downstream HTTP error slug against its status; an invalid or missing error payload becomes `502 error_bad_gateway`. The old `error_message` slug remains accepted with its HTTP status during a service rollout. The CV download path accepts a nonempty `token` and `lang`, forwards only selected response headers, and sends `Cache-Control: no-store` and `Referrer-Policy: no-referrer`.

## RabbitMQ request/reply

The gateway creates an exclusive, automatically deleted reply queue for each connection. Each request has a unique correlation ID. The client keeps pending replies in a synchronized map, publishes on a separate channel in confirm mode, and checks mandatory returns for requests without a route. It limits concurrent in-flight requests to 64, applies a ten-second request deadline, and reconnects after a connection failure. A publisher confirmation means RabbitMQ accepted the publish; the client still waits for the correlated reply.

## Configuration

The module requires Go 1.25.1 (see `go.mod`). `APP_ENV=production` selects `config/production/config.yml`; otherwise the local file is used. These environment variables can override service addresses in YAML: `CONTENT_SERVICE_GRPC_ADDR`, `CONTENT_SERVICE_HTTP_ADDR`, and `CAPTCHA_SERVICE_HTTP_ADDR`. Supply `RABBITMQ_URL` through the environment; its YAML field is intentionally empty. The YAML file also defines the HTTP port, RabbitMQ exchange and routing key, and startup retry policy.

The HTTP server bounds request headers and bodies. Content gRPC calls, CAPTCHA HTTP calls, and the RabbitMQ request/reply path have explicit deadlines. The gRPC client uses an internal insecure transport as configured in `internal/app`; deploy it only within a network boundary appropriate for that choice.

## Run locally

1. Start the content service, CAPTCHA service, and RabbitMQ. The content service also needs Redis; the CAPTCHA service and content service must share its instance for the CV flow.
2. Set `RABBITMQ_URL` in the environment. When running the gateway directly on the host, also set `CONTENT_SERVICE_GRPC_ADDR`, `CONTENT_SERVICE_HTTP_ADDR`, and `CAPTCHA_SERVICE_HTTP_ADDR` to addresses reachable from the host. The supplied local YAML uses container DNS names (`content-service` and `captcha-service`), which normally do not resolve on the host.
3. Run `go run .` from this repository root. The local HTTP port is 8082. With the content service running, check the content route:

   ```sh
   curl -i 'http://127.0.0.1:8082/api/v1/content?lang=pl'
   ```

The `Dockerfile` builds a container image, but service addresses and the RabbitMQ URL must still be reachable and configured in its runtime environment.

## Verification

```sh
go test -race ./...
```

An opt-in integration test checks RabbitMQ correlation under concurrent requests, unroutable requests, and reconnection:

```sh
go test -race -tags=integration ./internal/service/content_service/rabbitmq
```

Set `RABBIT_TEST_URL` to an isolated RabbitMQ instance before running it. The test skips when the variable is unset and creates temporary topology in the selected broker.
