package service

import (
	"context"
	"time"

	"github.com/ecommerce/order-service/internal/model"
	"github.com/stretchr/testify/mock"
)

// MockOrderRepository mock 实现，用于单元测试
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

func (m *MockOrderRepository) FindCartsByUserID(ctx context.Context, userID uint64) ([]model.Cart, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]model.Cart), args.Error(1)
}

func (m *MockOrderRepository) FindCartsByIDs(ctx context.Context, ids []uint64, userID uint64) ([]model.Cart, error) {
	args := m.Called(ctx, ids, userID)
	return args.Get(0).([]model.Cart), args.Error(1)
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

func (m *MockOrderRepository) DeleteCartsByIDs(ctx context.Context, ids []uint64, userID uint64) error {
	args := m.Called(ctx, ids, userID)
	return args.Error(0)
}

func (m *MockOrderRepository) CreateParentOrder(ctx context.Context, order *model.ParentOrder) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrderRepository) FindParentOrderByOrderNo(ctx context.Context, orderNo string) (*model.ParentOrder, error) {
	args := m.Called(ctx, orderNo)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ParentOrder), args.Error(1)
}

func (m *MockOrderRepository) UpdateParentOrder(ctx context.Context, order *model.ParentOrder) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrderRepository) ListParentOrders(ctx context.Context, buyerID uint64, status *uint8, page, pageSize int) ([]model.ParentOrder, int64, error) {
	args := m.Called(ctx, buyerID, status, page, pageSize)
	return args.Get(0).([]model.ParentOrder), args.Get(1).(int64), args.Error(2)
}

func (m *MockOrderRepository) CreateSubOrder(ctx context.Context, order *model.SubOrder) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrderRepository) FindSubOrdersByParentOrderNo(ctx context.Context, parentOrderNo string) ([]model.SubOrder, error) {
	args := m.Called(ctx, parentOrderNo)
	return args.Get(0).([]model.SubOrder), args.Error(1)
}

func (m *MockOrderRepository) FindSubOrderBySubOrderNo(ctx context.Context, subOrderNo string) (*model.SubOrder, error) {
	args := m.Called(ctx, subOrderNo)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.SubOrder), args.Error(1)
}

func (m *MockOrderRepository) UpdateSubOrder(ctx context.Context, order *model.SubOrder) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrderRepository) CreateOrderItems(ctx context.Context, items []model.OrderItem) error {
	args := m.Called(ctx, items)
	return args.Error(0)
}

func (m *MockOrderRepository) FindOrderItemsBySubOrderNo(ctx context.Context, subOrderNo string) ([]model.OrderItem, error) {
	args := m.Called(ctx, subOrderNo)
	return args.Get(0).([]model.OrderItem), args.Error(1)
}

func (m *MockOrderRepository) FindOrderAddressByOrderNo(ctx context.Context, orderNo string) (*model.OrderAddress, error) {
	args := m.Called(ctx, orderNo)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.OrderAddress), args.Error(1)
}

func (m *MockOrderRepository) CreateOrderAddress(ctx context.Context, addr *model.OrderAddress) error {
	args := m.Called(ctx, addr)
	return args.Error(0)
}

func (m *MockOrderRepository) CreateOrderActionLog(ctx context.Context, log *model.OrderActionLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockOrderRepository) CacheIdempotencyKey(ctx context.Context, key string, ttl time.Duration) error {
	args := m.Called(ctx, key, ttl)
	return args.Error(0)
}

func (m *MockOrderRepository) CheckIdempotencyKey(ctx context.Context, key string) (bool, error) {
	args := m.Called(ctx, key)
	return args.Bool(0), args.Error(1)
}

func (m *MockOrderRepository) FreezeStock(ctx context.Context, skuID uint64, quantity int) (bool, error) {
	args := m.Called(ctx, skuID, quantity)
	return args.Bool(0), args.Error(1)
}

func (m *MockOrderRepository) UnfreezeStock(ctx context.Context, skuID uint64, quantity int) error {
	args := m.Called(ctx, skuID, quantity)
	return args.Error(0)
}

func (m *MockOrderRepository) CreateFrozenStock(ctx context.Context, record *model.FrozenStock) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}

func (m *MockOrderRepository) UpdateFrozenStockState(ctx context.Context, id uint64, state uint8) error {
	args := m.Called(ctx, id, state)
	return args.Error(0)
}

func (m *MockOrderRepository) FindExpiredFrozenStocks(ctx context.Context, timeout time.Duration) ([]model.FrozenStock, error) {
	args := m.Called(ctx, timeout)
	return args.Get(0).([]model.FrozenStock), args.Error(1)
}

func (m *MockOrderRepository) GetFrozenStocksByOrderNo(ctx context.Context, orderNo string) ([]model.FrozenStock, error) {
	args := m.Called(ctx, orderNo)
	return args.Get(0).([]model.FrozenStock), args.Error(1)
}

// RepositoryInterface matches OrderService's repository expectations
type RepositoryInterface interface {
	FindCartByUserAndSku(ctx context.Context, userID, skuID uint64) (*model.Cart, error)
	FindCartByID(ctx context.Context, id, userID uint64) (*model.Cart, error)
	FindCartsByUserID(ctx context.Context, userID uint64) ([]model.Cart, error)
	FindCartsByIDs(ctx context.Context, ids []uint64, userID uint64) ([]model.Cart, error)
	CreateCart(ctx context.Context, cart *model.Cart) error
	UpdateCart(ctx context.Context, cart *model.Cart) error
	DeleteCart(ctx context.Context, id, userID uint64) error
	DeleteCartsByIDs(ctx context.Context, ids []uint64, userID uint64) error
	CreateParentOrder(ctx context.Context, order *model.ParentOrder) error
	FindParentOrderByOrderNo(ctx context.Context, orderNo string) (*model.ParentOrder, error)
	UpdateParentOrder(ctx context.Context, order *model.ParentOrder) error
	ListParentOrders(ctx context.Context, buyerID uint64, status *uint8, page, pageSize int) ([]model.ParentOrder, int64, error)
	CreateSubOrder(ctx context.Context, order *model.SubOrder) error
	FindSubOrdersByParentOrderNo(ctx context.Context, parentOrderNo string) ([]model.SubOrder, error)
	FindSubOrderBySubOrderNo(ctx context.Context, subOrderNo string) (*model.SubOrder, error)
	UpdateSubOrder(ctx context.Context, order *model.SubOrder) error
	CreateOrderItems(ctx context.Context, items []model.OrderItem) error
	FindOrderItemsBySubOrderNo(ctx context.Context, subOrderNo string) ([]model.OrderItem, error)
	FindOrderAddressByOrderNo(ctx context.Context, orderNo string) (*model.OrderAddress, error)
	CreateOrderAddress(ctx context.Context, addr *model.OrderAddress) error
	CreateOrderActionLog(ctx context.Context, log *model.OrderActionLog) error
	CacheIdempotencyKey(ctx context.Context, key string, ttl time.Duration) error
	CheckIdempotencyKey(ctx context.Context, key string) (bool, error)
	FreezeStock(ctx context.Context, skuID uint64, quantity int) (bool, error)
	UnfreezeStock(ctx context.Context, skuID uint64, quantity int) error
	CreateFrozenStock(ctx context.Context, record *model.FrozenStock) error
	UpdateFrozenStockState(ctx context.Context, id uint64, state uint8) error
	FindExpiredFrozenStocks(ctx context.Context, timeout time.Duration) ([]model.FrozenStock, error)
	GetFrozenStocksByOrderNo(ctx context.Context, orderNo string) ([]model.FrozenStock, error)
}
