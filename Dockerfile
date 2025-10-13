FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git openssh-client && \
    mkdir -p -m 0700 ~/.ssh && \
    ssh-keyscan github.com >> ~/.ssh/known_hosts && \
    git config --global url."git@github.com:".insteadOf "https://github.com/"

WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=ssh go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /app/gateway-service ./main.go

# ---

FROM gcr.io/distroless/static-debian11

WORKDIR /app

COPY --from=builder /app/config ./config
COPY --from=builder /app/gateway-service .

USER nonroot:nonroot

CMD ["/app/gateway-service"]