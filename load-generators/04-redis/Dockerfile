# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /build

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o load-generator ./cmd

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /build/load-generator .

LABEL obi.instrumentation.enabled="true"
LABEL obi.service.name="redis-load-generator"
LABEL obi.service.type="load-generator"

ENTRYPOINT ["/app/load-generator"]
