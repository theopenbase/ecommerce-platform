package repository

import (
	"context"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"

	"github.com/ecommerce/marketing-service/internal/model"
)

type MarketingRepository struct {
	db    *gorm.DB
	cache *redis.Client
}

func NewMarketingRepository(db *gorm.DB, cache *redis.Client) *MarketingRepository {
	return &MarketingRepository{db: db, cache: cache}
}

func (r *MarketingRepository) CreateCoupon(ctx context.Context, c *model.Coupon) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *MarketingRepository) FindCouponByID(ctx context.Context, id uint64) (*model.Coupon, error) {
	var c model.Coupon
	err := r.db.WithContext(ctx).First(&c, id).Error
	return &c, err
}

func (r *MarketingRepository) FindActiveCoupons(ctx context.Context, applicableType uint8, applicableID uint64) ([]model.Coupon, error) {
	var coupons []model.Coupon
	query := r.db.WithContext(ctx).Where("status = ?", model.CouponStatusActive).
		Where("remain_count > 0")
	if applicableType == 1 {
		query = query.Where("applicable_type = 0 OR applicable_type = 1")
	}
	err := query.Find(&coupons).Error
	return coupons, err
}

func (r *MarketingRepository) ListCoupons(ctx context.Context, shopID *uint64, page, size int) ([]model.Coupon, int64, error) {
	var coupons []model.Coupon
	query := r.db.WithContext(ctx).Model(&model.Coupon{})
	if shopID != nil {
		query = query.Where("shop_id = ? OR shop_id IS NULL", *shopID)
	}
	var total int64
	query.Count(&total)
	offset := (page - 1) * size
	err := query.Order("created_at DESC").Offset(offset).Limit(size).Find(&coupons).Error
	return coupons, total, err
}

func (r *MarketingRepository) UpdateCoupon(ctx context.Context, c *model.Coupon) error {
	return r.db.WithContext(ctx).Save(c).Error
}

func (r *MarketingRepository) DecrCouponRemainCount(ctx context.Context, couponID uint64) error {
	return r.db.WithContext(ctx).Exec("UPDATE coupons SET remain_count = remain_count - 1 WHERE id = ? AND remain_count > 0", couponID).Error
}

func (r *MarketingRepository) IncrCouponRemainCount(ctx context.Context, couponID uint64) error {
	return r.db.WithContext(ctx).Exec("UPDATE coupons SET remain_count = remain_count + 1 WHERE id = ?", couponID).Error
}

// UserCoupon
func (r *MarketingRepository) CreateUserCoupon(ctx context.Context, uc *model.UserCoupon) error {
	return r.db.WithContext(ctx).Create(uc).Error
}

func (r *MarketingRepository) FindUserCouponByID(ctx context.Context, id uint64) (*model.UserCoupon, error) {
	var uc model.UserCoupon
	err := r.db.WithContext(ctx).First(&uc, id).Error
	return &uc, err
}

func (r *MarketingRepository) CountUserCouponsByCouponID(ctx context.Context, userID, couponID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.UserCoupon{}).
		Where("user_id = ? AND coupon_id = ?", userID, couponID).Count(&count).Error
	return count, err
}

func (r *MarketingRepository) ListUserCoupons(ctx context.Context, userID uint64, status *uint8, page, size int) ([]model.UserCoupon, int64, error) {
	var ucs []model.UserCoupon
	query := r.db.WithContext(ctx).Where("user_id = ?", userID)
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	var total int64
	query.Model(&model.UserCoupon{}).Count(&total)
	offset := (page - 1) * size
	err := query.Order("received_at DESC").Offset(offset).Limit(size).Find(&ucs).Error
	return ucs, total, err
}

func (r *MarketingRepository) UpdateUserCoupon(ctx context.Context, uc *model.UserCoupon) error {
	return r.db.WithContext(ctx).Save(uc).Error
}

// Promotion
func (r *MarketingRepository) CreatePromotion(ctx context.Context, p *model.Promotion) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *MarketingRepository) FindPromotionByID(ctx context.Context, id uint64) (*model.Promotion, error) {
	var p model.Promotion
	err := r.db.WithContext(ctx).First(&p, id).Error
	return &p, err
}

func (r *MarketingRepository) ListActivePromotions(ctx context.Context) ([]model.Promotion, error) {
	var ps []model.Promotion
	err := r.db.WithContext(ctx).Where("status = 1").Find(&ps).Error
	return ps, err
}

func (r *MarketingRepository) UpdatePromotion(ctx context.Context, p *model.Promotion) error {
	return r.db.WithContext(ctx).Save(p).Error
}

// PromotionSku
func (r *MarketingRepository) CreatePromotionSku(ctx context.Context, ps *model.PromotionSku) error {
	return r.db.WithContext(ctx).Create(ps).Error
}

func (r *MarketingRepository) FindPromotionSkuBySkuID(ctx context.Context, skuID uint64) (*model.PromotionSku, error) {
	var ps model.PromotionSku
	err := r.db.WithContext(ctx).Where("sku_id = ?", skuID).First(&ps).Error
	return &ps, err
}

func (r *MarketingRepository) ListPromotionSkus(ctx context.Context, promotionID uint64) ([]model.PromotionSku, error) {
	var ps []model.PromotionSku
	err := r.db.WithContext(ctx).Where("promotion_id = ?", promotionID).Find(&ps).Error
	return ps, err
}

func (r *MarketingRepository) IncrPromotionSoldCount(ctx context.Context, id uint64, delta int) error {
	return r.db.WithContext(ctx).Exec("UPDATE promotion_skus SET sold_count = sold_count + ? WHERE id = ?", delta, id).Error
}
