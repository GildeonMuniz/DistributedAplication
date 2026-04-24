package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	apperrors "github.com/microservices/shared/errors"
	"github.com/microservices/user-service/internal/domain"
	"github.com/microservices/user-service/internal/service"
)

type UserHandler struct {
	svc *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

func (h *UserHandler) RegisterRoutes(r *gin.Engine) {
	v1 := r.Group("/api/v1")
	{
		v1.POST("/auth/login", h.Login)
		v1.POST("/users", h.Create)
		v1.GET("/users", h.List)
		v1.GET("/users/:id", h.GetByID)
		v1.PUT("/users/:id", h.Update)
		v1.DELETE("/users/:id", h.Delete)
	}
}

func (h *UserHandler) Create(c *gin.Context) {
	var input domain.CreateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.svc.Create(c.Request.Context(), input)
	if err != nil {
		respondAppError(c, err)
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (h *UserHandler) GetByID(c *gin.Context) {
	user, err := h.svc.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		respondAppError(c, err)
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	users, err := h.svc.List(c.Request.Context(), limit, offset)
	if err != nil {
		respondAppError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": users, "limit": limit, "offset": offset})
}

func (h *UserHandler) Update(c *gin.Context) {
	var input domain.UpdateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.svc.Update(c.Request.Context(), c.Param("id"), input)
	if err != nil {
		respondAppError(c, err)
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) Delete(c *gin.Context) {
	if err := h.svc.Delete(c.Request.Context(), c.Param("id")); err != nil {
		respondAppError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *UserHandler) Login(c *gin.Context) {
	var input domain.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	token, user, err := h.svc.Login(c.Request.Context(), input)
	if err != nil {
		respondAppError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token, "user": user})
}

func respondError(c *gin.Context, code int, msg string) {
	c.JSON(code, gin.H{"message": msg})
}

func respondAppError(c *gin.Context, err error) {
	if appErr, ok := err.(*apperrors.AppError); ok {
		c.JSON(appErr.Code, appErr)
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
}
