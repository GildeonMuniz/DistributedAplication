package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/microservices/shared/database"
	apperrors "github.com/microservices/shared/errors"
	"github.com/microservices/order-service/internal/domain"
)

type OrderRepository struct {
	db *database.DB
}

func NewOrderRepository(db *database.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(ctx context.Context, o *domain.Order) error {
	itemsJSON, err := json.Marshal(o.Items)
	if err != nil {
		return fmt.Errorf("marshal items: %w", err)
	}

	query := `
		INSERT INTO orders (id, user_id, items, total_amount, status, notes, created_at, updated_at)
		VALUES (@id, @user_id, @items, @total_amount, @status, @notes, @created_at, @updated_at)`

	_, err = r.db.ExecContext(ctx, query,
		sql.Named("id", o.ID),
		sql.Named("user_id", o.UserID),
		sql.Named("items", string(itemsJSON)),
		sql.Named("total_amount", o.TotalAmount),
		sql.Named("status", string(o.Status)),
		sql.Named("notes", o.Notes),
		sql.Named("created_at", o.CreatedAt),
		sql.Named("updated_at", o.UpdatedAt),
	)
	return err
}

func (r *OrderRepository) FindByID(ctx context.Context, id string) (*domain.Order, error) {
	query := `
		SELECT id, user_id, items, total_amount, status, notes, created_at, updated_at
		FROM orders WHERE id = @id`

	row := r.db.QueryRowContext(ctx, query, sql.Named("id", id))
	return scanOrder(row)
}

func (r *OrderRepository) ListByUser(ctx context.Context, userID string, limit, offset int) ([]*domain.Order, error) {
	query := `
		SELECT id, user_id, items, total_amount, status, notes, created_at, updated_at
		FROM orders WHERE user_id = @user_id
		ORDER BY created_at DESC
		OFFSET @offset ROWS FETCH NEXT @limit ROWS ONLY`

	rows, err := r.db.QueryContext(ctx, query,
		sql.Named("user_id", userID),
		sql.Named("limit", limit),
		sql.Named("offset", offset),
	)
	if err != nil {
		return nil, fmt.Errorf("list orders: %w", err)
	}
	defer rows.Close()

	var orders []*domain.Order
	for rows.Next() {
		o, err := scanOrderRow(rows)
		if err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, rows.Err()
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, id string, status domain.OrderStatus) error {
	query := `UPDATE orders SET status = @status, updated_at = GETDATE() WHERE id = @id`
	_, err := r.db.ExecContext(ctx, query,
		sql.Named("status", string(status)),
		sql.Named("id", id),
	)
	return err
}

func scanOrder(row *sql.Row) (*domain.Order, error) {
	o := &domain.Order{}
	var itemsJSON string
	err := row.Scan(&o.ID, &o.UserID, &itemsJSON, &o.TotalAmount, &o.Status, &o.Notes, &o.CreatedAt, &o.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, apperrors.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan order: %w", err)
	}
	if err := json.Unmarshal([]byte(itemsJSON), &o.Items); err != nil {
		return nil, fmt.Errorf("unmarshal items: %w", err)
	}
	return o, nil
}

func scanOrderRow(rows *sql.Rows) (*domain.Order, error) {
	o := &domain.Order{}
	var itemsJSON string
	err := rows.Scan(&o.ID, &o.UserID, &itemsJSON, &o.TotalAmount, &o.Status, &o.Notes, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("scan order: %w", err)
	}
	if err := json.Unmarshal([]byte(itemsJSON), &o.Items); err != nil {
		return nil, fmt.Errorf("unmarshal items: %w", err)
	}
	return o, nil
}
