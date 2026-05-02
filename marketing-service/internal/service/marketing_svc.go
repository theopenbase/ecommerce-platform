package service

import (
	"context"
	"errors"
	"time"

	"github.com/ecommerce/marketing-service/internal/model"
	"github.com/ecommerce/marketing-service/internal/repository"
)

var (
	ErrCouponNotFound      = errors.New("coupon not found")
	ErrCouponDepleted     = errors.New("coupon depleted")
	ErrCouponLimitReached = errors.New("coupon limit reached for user")
	ErrPromotionNotFound  = errors.New("promotion not found")
	ErrSkuNotInPromotion  = errors.New("sku not in promotion")
	ErrStockNotEnough    = errors.New("promotion stock not enough")
)

type MarketingService struct {
	repo *repository.MarketingRepository
}

func NewMarketingService(repo *repository.MarketingRepository) *MarketingService {
	return &MarketingService{repo: repo}
}

// ============ 优惠券 ============

// CreateCoupon 创建优惠券（商家/平台运营）
func (s *MarketingService) CreateCoupon(ctx context.Context, req *model.CreateCouponRequest) (*model.Coupon, error) {
	code := generateCouponCode()
	coupon := &model.Coupon{
		CouponCode:      code,
		Name:            req.Name,
		Type:           req.Type,
		FaceValue:      req.FaceValue,
		Threshold:      req.Threshold,
		TotalCount:    req.TotalCount,
		RemainCount:    req.TotalCount,
		PerUserLimit:  req.PerUserLimit,
		ValidType:     req.ValidType,
		ValidDays:     req.ValidDays,
		ApplicableType: req.ApplicableType,
		ApplicableIDs: req.ApplicableIDs,
		ShopID:        req.ShopID,
		Status:        model.CouponStatusPending,
	}
	if req.ValidType == 1 {
		start, _ := time.Parse(time.RFC3339, req.ValidStart)
		end, _ := time.Parse(time.RFC3339, req.ValidEnd)
		coupon.ValidStart = &start
		coupon.ValidEnd = &end
	}
	if err := s.repo.CreateCoupon(ctx, coupon); err != nil {
		return nil, err
	}
	return coupon, nil
}

// PublishCoupon 发布优惠券
func (s *MarketingService) PublishCoupon(ctx context.Context, couponID uint64) error {
	coupon, err := s.repo.FindCouponByID(ctx, couponID)
	if err != nil {
		return ErrCouponNotFound
	}
	coupon.Status = model.CouponStatusActive
	return s.repo.UpdateCoupon(ctx, coupon)
}

// ReceiveCoupon 领取优惠券
func (s *MarketingService) ReceiveCoupon(ctx context.Context, userID, couponID uint64) (*model.UserCoupon, error) {
	coupon, err := s.repo.FindCouponByID(ctx, couponID)
	if err != nil {
		return nil, ErrCouponNotFound
	}
	if coupon.Status != model.CouponStatusActive || coupon.RemainCount <= 0 {
		return nil, ErrCouponDepleted
	}

	// 检查用户限领
	count, _ := s.repo.CountUserCouponsByCouponID(ctx, userID, couponID)
	if int(count) >= coupon.PerUserLimit {
		return nil, ErrCouponLimitReached
	}

	// 扣减库存
	if err := s.repo.DecrCouponRemainCount(ctx, couponID); err != nil {
		return nil, ErrCouponDepleted
	}

	// 计算过期时间
	expireDate := time.Now().AddDate(0, 0, 30) // 默认30天
	if coupon.ValidType == 2 && coupon.ValidDays > 0 {
		expireDate = time.Now().AddDate(0, 0, coupon.ValidDays)
	} else if coupon.ValidType == 1 && coupon.ValidEnd != nil {
		expireDate = *coupon.ValidEnd
	}

	uc := &model.UserCoupon{
		UserID:     userID,
		CouponID:  couponID,
		CouponCode: coupon.CouponCode,
		Status:    model.UserCouponStatusUnused,
		ReceivedAt: time.Now(),
		ExpireDate: expireDate,
	}
	if err := s.repo.CreateUserCoupon(ctx, uc); err != nil {
		s.repo.IncrCouponRemainCount(ctx, couponID)
		return nil, err
	}
	return uc, nil
}

// ListUserCoupons 获取用户优惠券列表
func (s *MarketingService) ListUserCoupons(ctx context.Context, userID uint64, status *uint8, page, size int) ([]model.UserCoupon, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	return s.repo.ListUserCoupons(ctx, userID, status, page, size)
}

// ListCoupons 获取优惠券列表（商城展示）
func (s *MarketingService) ListCoupons(ctx context.Context, shopID *uint64, page, size int) ([]model.Coupon, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	return s.repo.ListCoupons(ctx, shopID, page, size)
}

// UseCoupon 核销优惠券（下单时调用）
func (s *MarketingService) UseCoupon(ctx context.Context, userCouponID uint64, orderNo string) error {
	uc, err := s.repo.FindUserCouponByID(ctx, userCouponID)
	if err != nil {
		return ErrCouponNotFound
	}
	if uc.Status != model.UserCouponStatusUnused {
		return errors.New("coupon already used or expired")
	}
	if time.Now().After(uc.ExpireDate) {
		return errors.New("coupon expired")
	}
	now := time.Now()
	uc.Status = model.UserCouponStatusUsed
	uc.UsedAt = &now
	uc.UsedOrderNo = orderNo
	return s.repo.UpdateUserCoupon(ctx, uc)
}

// ReturnCoupon 退货退还优惠券
func (s *MarketingService) ReturnCoupon(ctx context.Context, userCouponID uint64) error {
	uc, err := s.repo.FindUserCouponByID(ctx, userCouponID)
	if err != nil {
		return ErrCouponNotFound
	}
	if uc.Status != model.UserCouponStatusUsed {
		return nil
	}
	uc.Status = model.UserCouponStatusRefunded
	return s.repo.UpdateUserCoupon(ctx, uc)
}

// ============ 促销活动 ============

// CreatePromotion 创建促销活动
func (s *MarketingService) CreatePromotion(ctx context.Context, req *model.CreatePromotionRequest) (*model.Promotion, error) {
	start, _ := time.Parse(time.RFC3339, req.StartTime)
	end, _ := time.Parse(time.RFC3339, req.EndTime)
	status := uint8(0)
	if time.Now().After(start) && time.Now().Before(end) {
		status = 1
	}
	promo := &model.Promotion{
		Name:      req.Name,
		Type:      req.Type,
		StartTime: start,
		EndTime:   end,
		Status:    status,
		Rules:     req.Rules,
	}
	if err := s.repo.CreatePromotion(ctx, promo); err != nil {
		return nil, err
	}
	return promo, nil
}

// AddPromotionSku 添加活动商品
func (s *MarketingService) AddPromotionSku(ctx context.Context, promotionID uint64, req *model.AddPromotionSkuRequest) (*model.PromotionSku, error) {
	promo, err := s.repo.FindPromotionByID(ctx, promotionID)
	if err != nil {
		return nil, ErrPromotionNotFound
	}
	if promo.Status != 1 {
		return nil, errors.New("promotion not active")
	}
	ps := &model.PromotionSku{
		PromotionID: promotionID,
		SkuID:      req.SkuID,
		ShopID:     req.ShopID,
		PromoPrice: req.PromoPrice,
		StockLimit: req.StockLimit,
		SoldCount:  0,
	}
	if err := s.repo.CreatePromotionSku(ctx, ps); err != nil {
		return nil, err
	}
	return ps, nil
}

// GetPromotionDetail 获取活动详情
func (s *MarketingService) GetPromotionDetail(ctx context.Context, promotionID uint64) (*model.PromotionDetailResponse, error) {
	promo, err := s.repo.FindPromotionByID(ctx, promotionID)
	if err != nil {
		return nil, ErrPromotionNotFound
	}
	skus, _ := s.repo.ListPromotionSkus(ctx, promotionID)
	return &model.PromotionDetailResponse{Promotion: promo, Skus: skus}, nil
}

// GetSkuPromotion 获取商品正在进行的促销活动
func (s *MarketingService) GetSkuPromotion(ctx context.Context, skuID uint64) (*model.PromotionSku, error) {
	ps, err := s.repo.FindPromotionSkuBySkuID(ctx, skuID)
	if err != nil {
		return nil, ErrSkuNotInPromotion
	}
	promo, err := s.repo.FindPromotionByID(ctx, ps.PromotionID)
	if err != nil || promo.Status != 1 {
		return nil, ErrSkuNotInPromotion
	}
	return ps, nil
}

// DeductPromotionStock 扣减活动库存（下单成功时）
func (s *MarketingService) DeductPromotionStock(ctx context.Context, promotionSkuID uint64, quantity int) error {
	ps, err := s.repo.FindPromotionSkuBySkuID(ctx, promotionSkuID)
	if err != nil {
		return ErrSkuNotInPromotion
	}
	if ps.SoldCount+quantity > ps.StockLimit {
		return ErrStockNotEnough
	}
	return s.repo.IncrPromotionSoldCount(ctx, promotionSkuID, quantity)
}

func (s *MarketingService) FindPromotionSkuByID(ctx context.Context, id uint64) (*model.PromotionSku, error) {
	return s.repo.FindPromotionSkuBySkuID(ctx, id)
}

func generateCouponCode() string {
	return time.Now().Format("20060102150405") + randomString(8)
}

func randomString(n int) string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}
