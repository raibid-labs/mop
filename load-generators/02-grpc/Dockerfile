# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum* ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o load-generator ./cmd

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /build/load-generator .

# Add labels for OBI instrumentation
LABEL obi.instrumentation.enabled="true"
LABEL obi.service.name="grpc-load-generator"
LABEL obi.service.type="load-generator"

ENTRYPOINT ["/app/load-generator"]
