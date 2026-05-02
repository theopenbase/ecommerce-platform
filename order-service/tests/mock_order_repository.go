package service

import (
	"context"

	"github.com/ecommerce/order-service/internal/model"
	"github.com/stretchr/testify/mock"
)

type MockOrderRepository struct {
	mock.Mock
}

func (m *MockOrderRepository) FindCartByUserAndSku(ctx context.Context, userID, skuID uint64) (*model.Cart, error) {
	args := m.Called(ctx, userID, skuID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Cart), args.Error(1)
}

func (m *MockOrderRepository) FindCartByID(ctx context.Context, id, userID uint64) (*model.Cart, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Cart), args.Error(1)
}

func (m *MockOrderRepository) CreateCart(ctx context.Context, cart *model.Cart) error {
	args := m.Called(ctx, cart)
	return args.Error(0)
}

func (m *MockOrderRepository) UpdateCart(ctx context.Context, cart *model.Cart) error {
	args := m.Called(ctx, cart)
	return args.Error(0)
}

func (m *MockOrderRepository) DeleteCart(ctx context.Context, id, userID uint64) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func (m *MockOrderRepository) FreezeStock(ctx context.Context, skuID uint64, quantity int) (bool, error) {
	args := m.Called(ctx, skuID, quantity)
	return args.Bool(0), args.Error(1)
}

func (m *MockOrderRepository) UnfreezeStock(ctx context.Context, skuID uint64, quantity int) error {
	args := m.Called(ctx, skuID, quantity)
	return args.Error(0)
}

func (m *MockOrderRepository) CheckIdempotencyKey(ctx context.Context, key string) (bool, error) {
	args := m.Called(ctx, key)
	return args.Bool(0), args.Error(1)
}

func (m *MockOrderRepository) CacheIdempotencyKey(ctx context.Context, key string, ttl interface{}) error {
	args := m.Called(ctx, key, ttl)
	return args.Error(0)
}
