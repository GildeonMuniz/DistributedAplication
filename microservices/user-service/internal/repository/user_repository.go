package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/microservices/shared/database"
	apperrors "github.com/microservices/shared/errors"
	"github.com/microservices/user-service/internal/domain"
)

type UserRepository struct {
	db *database.DB
}

func NewUserRepository(db *database.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, u *domain.User) error {
	query := `
		INSERT INTO users (id, name, email, password_hash, role, active, created_at, updated_at)
		VALUES (@id, @name, @email, @password_hash, @role, @active, @created_at, @updated_at)`

	_, err := r.db.ExecContext(ctx, query,
		sql.Named("id", u.ID),
		sql.Named("name", u.Name),
		sql.Named("email", u.Email),
		sql.Named("password_hash", u.PasswordHash),
		sql.Named("role", string(u.Role)),
		sql.Named("active", u.Active),
		sql.Named("created_at", u.CreatedAt),
		sql.Named("updated_at", u.UpdatedAt),
	)
	if err != nil {
		return fmt.Errorf("insert user: %w", err)
	}
	return nil
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT id, name, email, password_hash, role, active, created_at, updated_at
		FROM users WHERE id = @id AND active = 1`

	row := r.db.QueryRowContext(ctx, query, sql.Named("id", id))
	return scanUser(row)
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, name, email, password_hash, role, active, created_at, updated_at
		FROM users WHERE email = @email AND active = 1`

	row := r.db.QueryRowContext(ctx, query, sql.Named("email", email))
	return scanUser(row)
}

func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]*domain.User, error) {
	query := `
		SELECT id, name, email, password_hash, role, active, created_at, updated_at
		FROM users WHERE active = 1
		ORDER BY created_at DESC
		OFFSET @offset ROWS FETCH NEXT @limit ROWS ONLY`

	rows, err := r.db.QueryContext(ctx, query,
		sql.Named("limit", limit),
		sql.Named("offset", offset),
	)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		u := &domain.User{}
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.Active, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *UserRepository) Update(ctx context.Context, u *domain.User) error {
	query := `
		UPDATE users SET name = @name, email = @email, updated_at = @updated_at
		WHERE id = @id`

	_, err := r.db.ExecContext(ctx, query,
		sql.Named("name", u.Name),
		sql.Named("email", u.Email),
		sql.Named("updated_at", u.UpdatedAt),
		sql.Named("id", u.ID),
	)
	return err
}

func (r *UserRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE users SET active = 0, updated_at = GETDATE() WHERE id = @id`
	_, err := r.db.ExecContext(ctx, query, sql.Named("id", id))
	return err
}

func scanUser(row *sql.Row) (*domain.User, error) {
	u := &domain.User{}
	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.Active, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, apperrors.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan user: %w", err)
	}
	return u, nil
}
