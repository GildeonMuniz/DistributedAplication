package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	apperrors "github.com/microservices/shared/errors"
	"github.com/microservices/order-service/internal/domain"
	"github.com/microservices/order-service/internal/service"
)

type OrderHandler struct {
	svc *service.OrderService
}

func NewOrderHandler(svc *service.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

func (h *OrderHandler) RegisterRoutes(r *gin.Engine) {
	v1 := r.Group("/api/v1")
	{
		v1.POST("/orders", h.Create)
		v1.GET("/orders/:id", h.GetByID)
		v1.GET("/users/:user_id/orders", h.ListByUser)
		v1.PATCH("/orders/:id/status", h.UpdateStatus)
	}
}

func (h *OrderHandler) Create(c *gin.Context) {
	var input domain.CreateOrderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	order, err := h.svc.Create(c.Request.Context(), input)
	if err != nil {
		respondAppError(c, err)
		return
	}
	c.JSON(http.StatusCreated, order)
}

func (h *OrderHandler) GetByID(c *gin.Context) {
	order, err := h.svc.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		respondAppError(c, err)
		return
	}
	c.JSON(http.StatusOK, order)
}

func (h *OrderHandler) ListByUser(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	orders, err := h.svc.ListByUser(c.Request.Context(), c.Param("user_id"), limit, offset)
	if err != nil {
		respondAppError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": orders, "limit": limit, "offset": offset})
}

func (h *OrderHandler) UpdateStatus(c *gin.Context) {
	var input domain.UpdateStatusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	order, err := h.svc.UpdateStatus(c.Request.Context(), c.Param("id"), input)
	if err != nil {
		respondAppError(c, err)
		return
	}
	c.JSON(http.StatusOK, order)
}

func respondAppError(c *gin.Context, err error) {
	if appErr, ok := err.(*apperrors.AppError); ok {
		c.JSON(appErr.Code, appErr)
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
}
