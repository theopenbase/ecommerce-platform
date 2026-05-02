package model

// ============ 请求 ============

type CreateCouponRequest struct {
	Name           string  `json:"name" binding:"required"`
	Type          uint8   `json:"type" binding:"required,oneof=1 2 3 4"`
	FaceValue     float64 `json:"face_value" binding:"required,gt=0"`
	Threshold     float64 `json:"threshold" binding:"gte=0"`
	TotalCount    int     `json:"total_count" binding:"required,gt=0"`
	PerUserLimit  int     `json:"per_user_limit" binding:"required,gt=0"`
	ValidType     uint8   `json:"valid_type" binding:"required,oneof=1 2"`
	ValidStart    string  `json:"valid_start"` // ISO datetime
	ValidEnd      string  `json:"valid_end"`   // ISO datetime
	ValidDays     int     `json:"valid_days"`
	ApplicableType uint8  `json:"applicable_type" binding:"required,oneof=0 1 2 3"`
	ApplicableIDs string  `json:"applicable_ids"`
	ShopID        *uint64 `json:"shop_id"`
}

type CreatePromotionRequest struct {
	Name      string `json:"name" binding:"required"`
	Type      string `json:"type" binding:"required,oneof=full_reduce flash_seckill group_buy tier_price pre_sell discount"`
	StartTime string `json:"start_time" binding:"required"`
	EndTime   string `json:"end_time" binding:"required"`
	Rules     string `json:"rules" binding:"required"` // JSON
}

type AddPromotionSkuRequest struct {
	SkuID      uint64  `json:"sku_id" binding:"required"`
	ShopID     uint64  `json:"shop_id" binding:"required"`
	PromoPrice float64 `json:"promo_price" binding:"required,gt=0"`
	StockLimit int     `json:"stock_limit" binding:"required,gt=0"`
}

// ============ 响应 ============

type CouponListResponse struct {
	Total int64     `json:"total"`
	Items []Coupon  `json:"items"`
}

type UserCouponListResponse struct {
	Total int64        `json:"total"`
	Items []UserCoupon `json:"items"`
}

type PromotionDetailResponse struct {
	Promotion *Promotion     `json:"promotion"`
	Skus      []PromotionSku `json:"skus"`
}
