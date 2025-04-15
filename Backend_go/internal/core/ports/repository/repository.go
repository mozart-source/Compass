package repository

import (
	"context"

	"github.com/google/uuid"
)

type Repository[T any] interface {
	Create(ctx context.Context, entity *T) error
	FindAll(ctx context.Context, offset, limit int) ([]T, error)
	FindByID(ctx context.Context, id uuid.UUID) (*T, error)
	Update(ctx context.Context, entity *T) error
	Delete(ctx context.Context, id uuid.UUID) error
}
