# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build producer
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o producer ./cmd/producer

# Build consumer
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o consumer ./cmd/consumer

# Producer stage
FROM alpine:latest AS producer

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /build/producer .

ENV KAFKA_BROKERS=localhost:9092
ENV LOG_LEVEL=info

CMD ["./producer"]

# Consumer stage
FROM alpine:latest AS consumer

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /build/consumer .

ENV KAFKA_BROKERS=localhost:9092
ENV CONSUMER_GROUP=default-group
ENV CONSUMER_TYPE=notifications
ENV LOG_LEVEL=info

CMD ["./consumer"]
