package service

import (
	"context"
	"log/slog"

	"github.com/microservices/shared/config"
	apperrors "github.com/microservices/shared/errors"
	"github.com/microservices/shared/middleware"
	"github.com/microservices/user-service/internal/domain"
	"github.com/microservices/user-service/internal/events"
	"github.com/microservices/user-service/internal/repository"
)

type UserService struct {
	repo      *repository.UserRepository
	publisher *events.Publisher
	jwt       config.JWTConfig
	log       *slog.Logger
}

func NewUserService(
	repo *repository.UserRepository,
	publisher *events.Publisher,
	jwt config.JWTConfig,
	log *slog.Logger,
) *UserService {
	return &UserService{repo: repo, publisher: publisher, jwt: jwt, log: log}
}

func (s *UserService) Create(ctx context.Context, input domain.CreateUserInput) (*domain.User, error) {
	if existing, _ := s.repo.FindByEmail(ctx, input.Email); existing != nil {
		return nil, apperrors.ErrConflict
	}

	user, err := domain.NewUser(input)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, apperrors.Wrap(err, "create user")
	}

	go s.publisher.UserCreated(context.Background(), events.UserPayload{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Role:  string(user.Role),
	})

	s.log.Info("user created", "id", user.ID, "email", user.Email)
	return user, nil
}

func (s *UserService) GetByID(ctx context.Context, id string) (*domain.User, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *UserService) List(ctx context.Context, limit, offset int) ([]*domain.User, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.repo.List(ctx, limit, offset)
}

func (s *UserService) Update(ctx context.Context, id string, input domain.UpdateUserInput) (*domain.User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	user.Update(input)

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, apperrors.Wrap(err, "update user")
	}

	go s.publisher.UserUpdated(context.Background(), events.UserPayload{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Role:  string(user.Role),
	})

	return user, nil
}

func (s *UserService) Delete(ctx context.Context, id string) error {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return apperrors.Wrap(err, "delete user")
	}

	go s.publisher.UserDeleted(context.Background(), id)
	return nil
}

func (s *UserService) Login(ctx context.Context, input domain.LoginInput) (string, *domain.User, error) {
	user, err := s.repo.FindByEmail(ctx, input.Email)
	if err != nil {
		return "", nil, apperrors.ErrUnauthorized
	}

	if !user.CheckPassword(input.Password) {
		return "", nil, apperrors.ErrUnauthorized
	}

	token, err := middleware.GenerateToken(user.ID, user.Email, string(user.Role), s.jwt.Secret, s.jwt.Expiration)
	if err != nil {
		return "", nil, apperrors.Wrap(err, "generate token")
	}

	return token, user, nil
}
