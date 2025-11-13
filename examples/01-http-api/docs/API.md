# API Reference

Complete API documentation for the HTTP REST API Product Catalog service.

## Base URL

- Local: `http://localhost:8080`
- Kubernetes: `http://<service-ip>`

## Authentication

This example does not require authentication. In production, add authentication middleware.

## Common Headers

### Request Headers
- `Content-Type: application/json` - For POST/PUT requests
- `X-Request-ID: <uuid>` - Optional, auto-generated if not provided

### Response Headers
- `X-Request-ID: <uuid>` - Unique request identifier
- `Content-Type: application/json` - All responses are JSON

## Endpoints

### Health Check

Check the service health status.

**Endpoint**: `GET /health`

**Response**: `200 OK`
```json
{
  "status": "healthy",
  "uptime": "2h15m30.5s",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

**Use Cases**:
- Kubernetes liveness probe
- Kubernetes readiness probe
- Service monitoring
- Load balancer health checks

---

### List Products

Retrieve a paginated list of all products.

**Endpoint**: `GET /products`

**Query Parameters**:
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `limit` | integer | 10 | Number of results (1-100) |
| `offset` | integer | 0 | Offset for pagination |

**Example**:
```bash
curl "http://localhost:8080/products?limit=20&offset=10"
```

**Response**: `200 OK`
```json
{
  "products": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Product Name",
      "description": "Product Description",
      "price": 99.99,
      "stock": 100,
      "created_at": "2024-01-15T10:00:00Z",
      "updated_at": "2024-01-15T10:00:00Z"
    }
  ],
  "total": 100,
  "limit": 20,
  "offset": 10
}
```

**Notes**:
- Maximum limit is 100
- Returns empty array if offset exceeds total

---

### Get Product

Retrieve a single product by ID.

**Endpoint**: `GET /products/:id`

**Path Parameters**:
| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | string (UUID) | Product ID |

**Example**:
```bash
curl "http://localhost:8080/products/550e8400-e29b-41d4-a716-446655440000"
```

**Response**: `200 OK`
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Product Name",
  "description": "Product Description",
  "price": 99.99,
  "stock": 100,
  "created_at": "2024-01-15T10:00:00Z",
  "updated_at": "2024-01-15T10:00:00Z"
}
```

**Error Responses**:
- `404 Not Found` - Product does not exist
```json
{
  "error": "Product not found"
}
```

---

### Create Product

Create a new product.

**Endpoint**: `POST /products`

**Request Body**:
```json
{
  "name": "New Product",
  "description": "Product Description",
  "price": 99.99,
  "stock": 100
}
```

**Field Validation**:
| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `name` | string | Yes | 3-100 characters |
| `description` | string | No | Any length |
| `price` | number | Yes | Greater than 0 |
| `stock` | integer | No | Greater than or equal to 0 |

**Example**:
```bash
curl -X POST "http://localhost:8080/products" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "New Product",
    "description": "Product Description",
    "price": 99.99,
    "stock": 100
  }'
```

**Response**: `201 Created`
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "New Product",
  "description": "Product Description",
  "price": 99.99,
  "stock": 100,
  "created_at": "2024-01-15T10:00:00Z",
  "updated_at": "2024-01-15T10:00:00Z"
}
```

**Error Responses**:
- `400 Bad Request` - Validation failed
```json
{
  "error": "Key: 'Product.Name' Error:Field validation for 'Name' failed on the 'required' tag"
}
```

---

### Update Product

Update an existing product.

**Endpoint**: `PUT /products/:id`

**Path Parameters**:
| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | string (UUID) | Product ID |

**Request Body**:
```json
{
  "name": "Updated Product",
  "description": "Updated Description",
  "price": 149.99,
  "stock": 50
}
```

**Example**:
```bash
curl -X PUT "http://localhost:8080/products/550e8400-e29b-41d4-a716-446655440000" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Updated Product",
    "price": 149.99,
    "stock": 50
  }'
```

**Response**: `200 OK`
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Updated Product",
  "description": "Updated Description",
  "price": 149.99,
  "stock": 50,
  "created_at": "2024-01-15T10:00:00Z",
  "updated_at": "2024-01-15T11:30:00Z"
}
```

**Error Responses**:
- `404 Not Found` - Product does not exist
- `400 Bad Request` - Validation failed

**Notes**:
- `created_at` is preserved
- `updated_at` is set to current time
- `id` cannot be changed

---

### Delete Product

Delete a product.

**Endpoint**: `DELETE /products/:id`

**Path Parameters**:
| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | string (UUID) | Product ID |

**Example**:
```bash
curl -X DELETE "http://localhost:8080/products/550e8400-e29b-41d4-a716-446655440000"
```

**Response**: `200 OK`
```json
{
  "message": "Product deleted successfully"
}
```

**Error Responses**:
- `404 Not Found` - Product does not exist

---

### Search Products

Search products by name or description.

**Endpoint**: `GET /search`

**Query Parameters**:
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `q` | string | Yes | Search query |
| `limit` | integer | No | Number of results (default: 10, max: 100) |
| `offset` | integer | No | Offset for pagination (default: 0) |

**Search Behavior**:
- Case-insensitive substring match
- Searches both `name` and `description` fields
- Returns products matching either field

**Example**:
```bash
curl "http://localhost:8080/search?q=laptop&limit=10"
```

**Response**: `200 OK`
```json
{
  "products": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Apple MacBook",
      "description": "Laptop computer",
      "price": 1999.99,
      "stock": 5,
      "created_at": "2024-01-15T10:00:00Z",
      "updated_at": "2024-01-15T10:00:00Z"
    }
  ],
  "total": 1,
  "limit": 10,
  "offset": 0
}
```

**Error Responses**:
- `400 Bad Request` - Missing query parameter
```json
{
  "error": "Query parameter 'q' is required"
}
```

---

### Testing Endpoints

#### Slow Endpoint

Simulates a slow endpoint with 1-3 second latency.

**Endpoint**: `GET /slow`

**Example**:
```bash
curl "http://localhost:8080/slow"
```

**Response**: `200 OK` (after 1-3 seconds)
```json
{
  "message": "Slow endpoint response",
  "delay_ms": 2347
}
```

**Use Cases**:
- Testing OBI latency tracking
- Load testing with realistic delays
- Timeout testing

---

#### Error Endpoint

Always returns a 500 Internal Server Error.

**Endpoint**: `GET /error`

**Example**:
```bash
curl "http://localhost:8080/error"
```

**Response**: `500 Internal Server Error`
```json
{
  "error": "Simulated internal server error",
  "code": "TESTING_ERROR"
}
```

**Use Cases**:
- Testing OBI error detection
- Error rate alerting
- Client error handling

---

## Rate Limiting

All endpoints are rate limited per IP address:

- **Rate**: 100 requests per second per IP
- **Response**: `429 Too Many Requests`

**Rate Limit Response**:
```json
{
  "error": "Rate limit exceeded",
  "message": "Too many requests from this IP address"
}
```

**Configuration**:
Set `RATE_LIMIT_RPS` environment variable to change the limit.

---

## Error Handling

All errors return JSON responses with appropriate HTTP status codes.

### Error Response Format
```json
{
  "error": "Error message",
  "message": "Detailed error description"
}
```

### Common Status Codes

| Code | Description |
|------|-------------|
| `200` | Success |
| `201` | Created |
| `400` | Bad Request (validation error) |
| `404` | Not Found |
| `429` | Too Many Requests (rate limited) |
| `500` | Internal Server Error |
| `504` | Gateway Timeout (request timeout) |

---

## Examples

### Create and Update Flow

```bash
# 1. Create a product
PRODUCT_ID=$(curl -s -X POST "http://localhost:8080/products" \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Product","price":99.99,"stock":100}' \
  | jq -r '.id')

echo "Created product: $PRODUCT_ID"

# 2. Get the product
curl "http://localhost:8080/products/$PRODUCT_ID" | jq

# 3. Update the product
curl -X PUT "http://localhost:8080/products/$PRODUCT_ID" \
  -H "Content-Type: application/json" \
  -d '{"name":"Updated Product","price":149.99,"stock":50}' | jq

# 4. Delete the product
curl -X DELETE "http://localhost:8080/products/$PRODUCT_ID" | jq
```

### Pagination Example

```bash
# Get first page
curl "http://localhost:8080/products?limit=10&offset=0" | jq

# Get second page
curl "http://localhost:8080/products?limit=10&offset=10" | jq

# Get third page
curl "http://localhost:8080/products?limit=10&offset=20" | jq
```

### Search Example

```bash
# Search for "Apple" products
curl "http://localhost:8080/search?q=Apple" | jq

# Search with pagination
curl "http://localhost:8080/search?q=laptop&limit=5&offset=0" | jq
```

---

## Best Practices

1. **Always include request IDs**: Use `X-Request-ID` header for distributed tracing
2. **Handle rate limiting**: Implement exponential backoff for 429 responses
3. **Validate input**: Client-side validation improves UX
4. **Use pagination**: Always specify reasonable limits for list operations
5. **Handle errors**: Check status codes and parse error messages
6. **Set timeouts**: Configure client timeouts (30s recommended)

---

## Performance Tips

- Use pagination for large datasets
- Cache frequent queries client-side
- Use search instead of listing + filtering
- Batch operations when possible
- Monitor rate limit headers

---

For more information, see the main [README](../README.md).
