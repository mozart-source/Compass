package service

import (
	"context"

	"github.com/google/uuid"
)

type Input interface{}

type Service[T any, CreateIn Input, UpdateIn Input] interface {
	Create(ctx context.Context, input CreateIn) (*T, error)
	Get(ctx context.Context, id uuid.UUID) (*T, error)
	List(ctx context.Context, page, pageSize int) ([]T, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateIn) (*T, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type BaseService[T any, CreateIn Input, UpdateIn Input] struct {
	Service[T, CreateIn, UpdateIn]
}

func ValidatePagination(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	return page, pageSize
}
