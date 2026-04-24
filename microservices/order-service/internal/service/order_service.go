package service

import (
	"context"
	"log/slog"

	apperrors "github.com/microservices/shared/errors"
	"github.com/microservices/order-service/internal/domain"
	"github.com/microservices/order-service/internal/events"
	"github.com/microservices/order-service/internal/repository"
)

type OrderService struct {
	repo      *repository.OrderRepository
	publisher *events.Publisher
	log       *slog.Logger
}

func NewOrderService(repo *repository.OrderRepository, publisher *events.Publisher, log *slog.Logger) *OrderService {
	return &OrderService{repo: repo, publisher: publisher, log: log}
}

func (s *OrderService) Create(ctx context.Context, input domain.CreateOrderInput) (*domain.Order, error) {
	order, err := domain.NewOrder(input)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, order); err != nil {
		return nil, apperrors.Wrap(err, "create order")
	}

	go s.publisher.OrderCreated(context.Background(), order)

	s.log.Info("order created", "id", order.ID, "user_id", order.UserID, "total", order.TotalAmount)
	return order, nil
}

func (s *OrderService) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *OrderService) ListByUser(ctx context.Context, userID string, limit, offset int) ([]*domain.Order, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.repo.ListByUser(ctx, userID, limit, offset)
}

func (s *OrderService) UpdateStatus(ctx context.Context, id string, input domain.UpdateStatusInput) (*domain.Order, error) {
	order, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if !order.CanTransitionTo(input.Status) {
		return nil, apperrors.New(422, "invalid status transition",
			string(order.Status)+" -> "+string(input.Status))
	}

	oldStatus := order.Status

	if err := s.repo.UpdateStatus(ctx, id, input.Status); err != nil {
		return nil, apperrors.Wrap(err, "update order status")
	}

	order.Status = input.Status

	go s.publisher.StatusChanged(context.Background(), events.StatusChangedPayload{
		OrderID:   order.ID,
		UserID:    order.UserID,
		OldStatus: oldStatus,
		NewStatus: input.Status,
	})

	s.log.Info("order status updated", "id", order.ID, "from", oldStatus, "to", input.Status)
	return order, nil
}
