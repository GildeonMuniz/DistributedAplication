package domain

import (
	"time"

	"github.com/google/uuid"
	apperrors "github.com/microservices/shared/errors"
)

type OrderStatus string

const (
	StatusPending    OrderStatus = "pending"
	StatusConfirmed  OrderStatus = "confirmed"
	StatusProcessing OrderStatus = "processing"
	StatusShipped    OrderStatus = "shipped"
	StatusDelivered  OrderStatus = "delivered"
	StatusCancelled  OrderStatus = "cancelled"
)

type OrderItem struct {
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	TotalPrice  float64 `json:"total_price"`
}

type Order struct {
	ID          string      `json:"id"`
	UserID      string      `json:"user_id"`
	Items       []OrderItem `json:"items"`
	TotalAmount float64     `json:"total_amount"`
	Status      OrderStatus `json:"status"`
	Notes       string      `json:"notes,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

type CreateOrderInput struct {
	UserID string              `json:"user_id" binding:"required"`
	Items  []CreateItemInput   `json:"items"   binding:"required,min=1,dive"`
	Notes  string              `json:"notes"`
}

type CreateItemInput struct {
	ProductID   string  `json:"product_id"   binding:"required"`
	ProductName string  `json:"product_name" binding:"required"`
	Quantity    int     `json:"quantity"     binding:"required,min=1"`
	UnitPrice   float64 `json:"unit_price"   binding:"required,gt=0"`
}

type UpdateStatusInput struct {
	Status OrderStatus `json:"status" binding:"required"`
}

func NewOrder(input CreateOrderInput) (*Order, error) {
	if len(input.Items) == 0 {
		return nil, apperrors.New(400, "order must have at least one item", "")
	}

	items := make([]OrderItem, 0, len(input.Items))
	var total float64

	for _, i := range input.Items {
		itemTotal := float64(i.Quantity) * i.UnitPrice
		items = append(items, OrderItem{
			ProductID:   i.ProductID,
			ProductName: i.ProductName,
			Quantity:    i.Quantity,
			UnitPrice:   i.UnitPrice,
			TotalPrice:  itemTotal,
		})
		total += itemTotal
	}

	now := time.Now()
	return &Order{
		ID:          uuid.NewString(),
		UserID:      input.UserID,
		Items:       items,
		TotalAmount: total,
		Status:      StatusPending,
		Notes:       input.Notes,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

func (o *Order) CanTransitionTo(next OrderStatus) bool {
	transitions := map[OrderStatus][]OrderStatus{
		StatusPending:    {StatusConfirmed, StatusCancelled},
		StatusConfirmed:  {StatusProcessing, StatusCancelled},
		StatusProcessing: {StatusShipped, StatusCancelled},
		StatusShipped:    {StatusDelivered},
		StatusDelivered:  {},
		StatusCancelled:  {},
	}

	for _, allowed := range transitions[o.Status] {
		if allowed == next {
			return true
		}
	}
	return false
}
