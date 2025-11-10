# Stage 1: Build the protobuf and Go binary
FROM golang:1.25.4-alpine AS builder

# Install protoc and build dependencies
RUN apk add --no-cache \
    git \
    protobuf \
    protobuf-dev \
    make

WORKDIR /build

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Install protoc plugins
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Copy the entire source code
COPY . .

# Generate protobuf code
RUN make proto

# Build the server binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w -s' -o /grpc-server ./cmd/server

# Stage 2: Create minimal runtime image
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /grpc-server .

# Expose gRPC port
EXPOSE 9090

# Run the server
CMD ["./grpc-server"]
