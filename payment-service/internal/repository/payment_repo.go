package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"

	"github.com/ecommerce/payment-service/internal/model"
)

type PaymentRepository struct {
	db    *gorm.DB
	cache *redis.Client
}

func NewPaymentRepository(db *gorm.DB, cache *redis.Client) *PaymentRepository {
	return &PaymentRepository{db: db, cache: cache}
}

func (r *PaymentRepository) CreateTrans(ctx context.Context, t *model.PaymentTransaction) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *PaymentRepository) FindTransByTransNo(ctx context.Context, transNo string) (*model.PaymentTransaction, error) {
	var t model.PaymentTransaction
	err := r.db.WithContext(ctx).Where("trans_no = ?", transNo).First(&t).Error
	return &t, err
}

func (r *PaymentRepository) FindTransByOrderNo(ctx context.Context, orderNo string) (*model.PaymentTransaction, error) {
	var t model.PaymentTransaction
	err := r.db.WithContext(ctx).Where("order_no = ?", orderNo).First(&t).Error
	return &t, err
}

func (r *PaymentRepository) UpdateTrans(ctx context.Context, t *model.PaymentTransaction) error {
	return r.db.WithContext(ctx).Save(t).Error
}

func (r *PaymentRepository) FindAccountByUserID(ctx context.Context, userID uint64, userType uint8) (*model.Account, error) {
	var a model.Account
	err := r.db.WithContext(ctx).Where("user_id = ? AND user_type = ?", userID, userType).First(&a).Error
	return &a, err
}

func (r *PaymentRepository) CreateAccount(ctx context.Context, a *model.Account) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *PaymentRepository) UpdateAccount(ctx context.Context, a *model.Account) error {
	return r.db.WithContext(ctx).Save(a).Error
}

func (r *PaymentRepository) CreateAccountLog(ctx context.Context, l *model.AccountLog) error {
	return r.db.WithContext(ctx).Create(l).Error
}

func (r *PaymentRepository) FindAccountLogs(ctx context.Context, accountID uint64, limit int) ([]model.AccountLog, error) {
	var logs []model.AccountLog
	err := r.db.WithContext(ctx).Where("account_id = ?", accountID).Order("created_at DESC").Limit(limit).Find(&logs).Error
	return logs, err
}

func (r *PaymentRepository) CreateRefund(ctx context.Context, rfd *model.RefundRecord) error {
	return r.db.WithContext(ctx).Create(rfd).Error
}

func (r *PaymentRepository) FindRefundByRefundNo(ctx context.Context, refundNo string) (*model.RefundRecord, error) {
	var rfd model.RefundRecord
	err := r.db.WithContext(ctx).Where("refund_no = ?", refundNo).First(&rfd).Error
	return &rfd, err
}

func (r *PaymentRepository) UpdateRefund(ctx context.Context, rfd *model.RefundRecord) error {
	return r.db.WithContext(ctx).Save(rfd).Error
}

func (r *PaymentRepository) AcquireLock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return r.cache.SetNX(ctx, fmt.Sprintf("lock:%s", key), "1", ttl).Result()
}

func (r *PaymentRepository) ReleaseLock(ctx context.Context, key string) error {
	return r.cache.Del(ctx, fmt.Sprintf("lock:%s", key)).Err()
}
