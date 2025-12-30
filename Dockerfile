FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /app/gateway-service ./main.go

FROM gcr.io/distroless/static-debian11

WORKDIR /app

COPY --from=builder /app/config ./config
COPY --from=builder /app/gateway-service .
USER nonroot:nonroot

CMD ["/app/gateway-service"]