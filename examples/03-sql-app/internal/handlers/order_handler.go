package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/raibid-labs/mop/examples/03-sql-app/internal/models"
	"github.com/raibid-labs/mop/examples/03-sql-app/internal/repository"
)

// OrderHandler handles order-related HTTP requests
type OrderHandler struct {
	repo *repository.OrderRepository
}

// NewOrderHandler creates a new order handler
func NewOrderHandler(repo *repository.OrderRepository) *OrderHandler {
	return &OrderHandler{repo: repo}
}

// Create creates a new order
func (h *OrderHandler) Create(c *gin.Context) {
	var req models.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := h.repo.Create(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create order"})
		return
	}

	c.JSON(http.StatusCreated, order)
}

// GetByID retrieves an order by ID
func (h *OrderHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID"})
		return
	}

	// Check if we should include customer details
	includeCustomer := c.Query("include_customer") == "true"

	var order *models.OrderWithItems
	if includeCustomer {
		order, err = h.repo.GetByIDWithCustomer(c.Request.Context(), id)
	} else {
		order, err = h.repo.GetByID(c.Request.Context(), id)
	}

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}

	c.JSON(http.StatusOK, order)
}

// ListByCustomer retrieves all orders for a customer
func (h *OrderHandler) ListByCustomer(c *gin.Context) {
	customerID, err := strconv.ParseInt(c.Param("customer_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer ID"})
		return
	}

	limit := 20
	offset := 0

	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	orders, err := h.repo.ListByCustomer(c.Request.Context(), customerID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list orders"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"orders":      orders,
		"customer_id": customerID,
		"limit":       limit,
		"offset":      offset,
	})
}

// UpdateStatus updates the status of an order
func (h *OrderHandler) UpdateStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID"})
		return
	}

	var req models.UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.UpdateStatus(c.Request.Context(), id, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update order status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "order status updated"})
}

// GetStats retrieves order statistics for a customer
func (h *OrderHandler) GetStats(c *gin.Context) {
	customerID, err := strconv.ParseInt(c.Param("customer_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer ID"})
		return
	}

	stats, err := h.repo.GetOrderStats(c.Request.Context(), customerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get order stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// SimulateSlowQuery demonstrates a slow query pattern for OBI testing
func (h *OrderHandler) SimulateSlowQuery(c *gin.Context) {
	customerID, err := strconv.ParseInt(c.Param("customer_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer ID"})
		return
	}

	orders, err := h.repo.SimulateSlowQuery(c.Request.Context(), customerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get orders"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"orders":      orders,
		"customer_id": customerID,
		"warning":     "This endpoint uses N+1 queries for demonstration purposes",
	})
}
