package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"

	"github.com/ecommerce/order-service/internal/model"
)

type OrderRepository struct {
	db    *gorm.DB
	cache *redis.Client
}

func NewOrderRepository(db *gorm.DB, cache *redis.Client) *OrderRepository {
	return &OrderRepository{db: db, cache: cache}
}

// ============ 事务支持 ============

// TxFunc 事务回调函数类型
type TxFunc func(tx *gorm.DB) error

// WithTx 在事务中执行操作
func (r *OrderRepository) WithTx(ctx context.Context, fn TxFunc) error {
	return r.db.WithContext(ctx).Transaction(fn)
}

// DB 返回底层 gorm.DB（用于服务层创建事务）
func (r *OrderRepository) DB() *gorm.DB {
	return r.db
}

// ============ 购物车 ============

func (r *OrderRepository) FindCartByUserAndSku(ctx context.Context, userID, skuID uint64) (*model.Cart, error) {
	var cart model.Cart
	err := r.db.WithContext(ctx).Where("user_id = ? AND sku_id = ?", userID, skuID).First(&cart).Error
	if err != nil {
		return nil, err
	}
	return &cart, nil
}

func (r *OrderRepository) FindCartByID(ctx context.Context, id, userID uint64) (*model.Cart, error) {
	var cart model.Cart
	err := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).First(&cart).Error
	if err != nil {
		return nil, err
	}
	return &cart, nil
}

func (r *OrderRepository) FindCartsByUserID(ctx context.Context, userID uint64) ([]model.Cart, error) {
	var carts []model.Cart
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&carts).Error
	return carts, err
}

func (r *OrderRepository) FindCheckedCartsByUserID(ctx context.Context, userID uint64) ([]model.Cart, error) {
	var carts []model.Cart
	err := r.db.WithContext(ctx).Where("user_id = ? AND checked = 1", userID).Find(&carts).Error
	return carts, err
}

func (r *OrderRepository) FindCartsByIDs(ctx context.Context, ids []uint64, userID uint64) ([]model.Cart, error) {
	var carts []model.Cart
	err := r.db.WithContext(ctx).Where("id IN ? AND user_id = ?", ids, userID).Find(&carts).Error
	return carts, err
}

func (r *OrderRepository) CreateCart(ctx context.Context, cart *model.Cart) error {
	return r.db.WithContext(ctx).Create(cart).Error
}

func (r *OrderRepository) UpdateCart(ctx context.Context, cart *model.Cart) error {
	return r.db.WithContext(ctx).Save(cart).Error
}

func (r *OrderRepository) DeleteCart(ctx context.Context, id, userID uint64) error {
	return r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&model.Cart{}).Error
}

func (r *OrderRepository) DeleteCartsByIDs(ctx context.Context, ids []uint64, userID uint64) error {
	return r.db.WithContext(ctx).Where("id IN ? AND user_id = ?", ids, userID).Delete(&model.Cart{}).Error
}

func (r *OrderRepository) ClearCheckedCart(ctx context.Context, userID uint64) error {
	return r.db.WithContext(ctx).Where("user_id = ? AND checked = 1", userID).Delete(&model.Cart{}).Error
}

// ============ 父订单 ============

func (r *OrderRepository) CreateParentOrder(ctx context.Context, order *model.ParentOrder) error {
	return r.db.WithContext(ctx).Create(order).Error
}

func (r *OrderRepository) FindParentOrderByOrderNo(ctx context.Context, orderNo string) (*model.ParentOrder, error) {
	var order model.ParentOrder
	err := r.db.WithContext(ctx).Where("order_no = ?", orderNo).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepository) FindParentOrderByID(ctx context.Context, id, buyerID uint64) (*model.ParentOrder, error) {
	var order model.ParentOrder
	err := r.db.WithContext(ctx).Where("id = ? AND buyer_id = ?", id, buyerID).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepository) UpdateParentOrder(ctx context.Context, order *model.ParentOrder) error {
	return r.db.WithContext(ctx).Save(order).Error
}

func (r *OrderRepository) ListParentOrders(ctx context.Context, buyerID uint64, status *uint8, page, pageSize int) ([]model.ParentOrder, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.ParentOrder{}).Where("buyer_id = ?", buyerID)
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	var total int64
	query.Count(&total)
	offset := (page - 1) * pageSize
	var orders []model.ParentOrder
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&orders).Error
	return orders, total, err
}

// ============ 子订单 ============

func (r *OrderRepository) CreateSubOrder(ctx context.Context, order *model.SubOrder) error {
	return r.db.WithContext(ctx).Create(order).Error
}

func (r *OrderRepository) FindSubOrdersByParentOrderNo(ctx context.Context, parentOrderNo string) ([]model.SubOrder, error) {
	var orders []model.SubOrder
	err := r.db.WithContext(ctx).Where("parent_order_no = ?", parentOrderNo).Find(&orders).Error
	return orders, err
}

func (r *OrderRepository) FindSubOrderBySubOrderNo(ctx context.Context, subOrderNo string) (*model.SubOrder, error) {
	var order model.SubOrder
	err := r.db.WithContext(ctx).Where("sub_order_no = ?", subOrderNo).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepository) UpdateSubOrder(ctx context.Context, order *model.SubOrder) error {
	return r.db.WithContext(ctx).Save(order).Error
}

// ============ 订单商品项 ============

func (r *OrderRepository) CreateOrderItems(ctx context.Context, items []model.OrderItem) error {
	if len(items) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&items).Error
}

func (r *OrderRepository) FindOrderItemsBySubOrderNo(ctx context.Context, subOrderNo string) ([]model.OrderItem, error) {
	var items []model.OrderItem
	err := r.db.WithContext(ctx).Where("sub_order_no = ?", subOrderNo).Find(&items).Error
	return items, err
}

// ============ 订单地址 ============

func (r *OrderRepository) CreateOrderAddress(ctx context.Context, addr *model.OrderAddress) error {
	return r.db.WithContext(ctx).Create(addr).Error
}

func (r *OrderRepository) FindOrderAddressByOrderNo(ctx context.Context, orderNo string) (*model.OrderAddress, error) {
	var addr model.OrderAddress
	err := r.db.WithContext(ctx).Where("order_no = ?", orderNo).First(&addr).Error
	if err != nil {
		return nil, err
	}
	return &addr, nil
}

// ============ 订单操作日志 ============

func (r *OrderRepository) CreateOrderActionLog(ctx context.Context, log *model.OrderActionLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// ============ 分布式锁 ============

// AcquireLock 获取分布式锁（SETNX）
func (r *OrderRepository) AcquireLock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	lockKey := fmt.Sprintf("lock:%s", key)
	result, err := r.cache.SetNX(ctx, lockKey, "1", ttl).Result()
	return result, err
}

// ReleaseLock 释放分布式锁
func (r *OrderRepository) ReleaseLock(ctx context.Context, key string) error {
	lockKey := fmt.Sprintf("lock:%s", key)
	return r.cache.Del(ctx, lockKey).Err()
}

// ============ 库存预占（Redis）============

// FreezeStock 冻结库存（Redis Lua 原子扣减可用库存）
func (r *OrderRepository) FreezeStock(ctx context.Context, skuID uint64, quantity int) (bool, error) {
	key := fmt.Sprintf("stock:%d", skuID)
	script := `
		local stock = tonumber(redis.call('GET', KEYS[1]))
		if stock == nil or stock < tonumber(ARGV[1]) then
			return 0
		end
		redis.call('DECRBY', KEYS[1], ARGV[1])
		return 1
	`
	result, err := r.cache.Eval(ctx, script, []string{key}, quantity).Int()
	return result == 1, err
}

// UnfreezeStock 解冻库存（Redis 归还可用库存）
func (r *OrderRepository) UnfreezeStock(ctx context.Context, skuID uint64, quantity int) error {
	key := fmt.Sprintf("stock:%d", skuID)
	return r.cache.IncrBy(ctx, key, int64(quantity)).Err()
}

// DeductStock 扣减库存（真正下单成功，冻结→已售罄）
func (r *OrderRepository) DeductStock(ctx context.Context, skuID uint64, quantity int) error {
	key := fmt.Sprintf("stock:%d", skuID)
	return r.cache.DecrBy(ctx, key, int64(quantity)).Err()
}

// ============ 冻结库存记录（DB，用于超卖补偿）============

const (
	FrozenStateActive  = 0 // 冻结中（等待支付）
	FrozenStateUsed    = 1 // 已使用（支付成功，库存已真正扣减）
	FrozenStateRolled  = 2 // 已回滚（超时/取消，库存已解冻）
)

// CreateFrozenStock 创建冻结记录
func (r *OrderRepository) CreateFrozenStock(ctx context.Context, record *model.FrozenStock) error {
	return r.db.WithContext(ctx).Create(record).Error
}

// UpdateFrozenStockState 更新冻结记录状态
func (r *OrderRepository) UpdateFrozenStockState(ctx context.Context, id uint64, state uint8) error {
	return r.db.WithContext(ctx).Model(&model.FrozenStock{}).
		Where("id = ?", id).
		Update("state", state).Error
}

// FindExpiredFrozenStocks 查找超时未支付的冻结记录（用于补偿任务）
// 超时时间默认 30 分钟
func (r *OrderRepository) FindExpiredFrozenStocks(ctx context.Context, timeout time.Duration) ([]model.FrozenStock, error) {
	var records []model.FrozenStock
	cutoff := time.Now().Add(-timeout)
	err := r.db.WithContext(ctx).
		Where("state = ? AND created_at < ?", FrozenStateActive, cutoff).
		Find(&records).Error
	return records, err
}

// GetFrozenStocksByOrderNo 查询某订单的所有冻结记录
func (r *OrderRepository) GetFrozenStocksByOrderNo(ctx context.Context, orderNo string) ([]model.FrozenStock, error) {
	var records []model.FrozenStock
	err := r.db.WithContext(ctx).
		Where("order_no = ?", orderNo).
		Find(&records).Error
	return records, err
}

// ============ 幂等 ============

func (r *OrderRepository) CacheIdempotencyKey(ctx context.Context, key string, ttl time.Duration) error {
	cacheKey := fmt.Sprintf("idempotent:%s", key)
	return r.cache.Set(ctx, cacheKey, "1", ttl).Err()
}

func (r *OrderRepository) CheckIdempotencyKey(ctx context.Context, key string) (bool, error) {
	cacheKey := fmt.Sprintf("idempotent:%s", key)
	exists, err := r.cache.Exists(ctx, cacheKey).Result()
	return exists > 0, err
}
