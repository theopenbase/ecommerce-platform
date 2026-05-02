package model

import "time"

// 优惠券类型
const (
	CouponTypePlatform uint8 = 1 // 平台券
	CouponTypeShop    uint8 = 2 // 店铺券
	CouponTypeGoods  uint8 = 3 // 商品券
	CouponTypeInternal uint8 = 4 // 内部券
)

// 优惠券状态
const (
	CouponStatusPending  uint8 = 1 // 待发放
	CouponStatusActive   uint8 = 2 // 发放中
	CouponStatusDepleted uint8 = 3 // 已领完
	CouponStatusExpired  uint8 = 4 // 已过期
)

// 用户优惠券状态
const (
	UserCouponStatusUnused   uint8 = 0 // 未使用
	UserCouponStatusUsed     uint8 = 1 // 已使用
	UserCouponStatusExpired  uint8 = 2 // 已过期
	UserCouponStatusRefunded uint8 = 3 // 已退款
)

// 促销活动类型
const (
	PromoTypeFullReduce  = "full_reduce"  // 满减
	PromoTypeFlashSeckill = "flash_seckill" // 秒杀
	PromoTypeGroupBuy   = "group_buy"   // 拼团
	PromoTypeTierPrice  = "tier_price"  // 阶梯价
	PromoTypePreSell   = "pre_sell"    // 预售
	PromoTypeDiscount  = "discount"    // 折扣
)

type Coupon struct {
	ID              uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	CouponCode      string     `gorm:"type:varchar(32);uniqueIndex;not null" json:"coupon_code"`
	Name            string     `gorm:"type:varchar(32);not null" json:"name"`
	Type           uint8      `gorm:"type:tinyint;not null" json:"type"`
	FaceValue      float64     `gorm:"type:decimal(10,2);not null" json:"face_value"`
	Threshold      float64     `gorm:"type:decimal(10,2);not null;default:0" json:"threshold"`
	TotalCount     int         `gorm:"not null" json:"total_count"`
	RemainCount    int         `gorm:"not null" json:"remain_count"`
	PerUserLimit   int         `gorm:"not null;default:1" json:"per_user_limit"`
	ValidType      uint8      `gorm:"type:tinyint;not null" json:"valid_type"` // 1-绝对时间 2-相对领取日
	ValidStart     *time.Time `gorm:"type:datetime" json:"valid_start"`
	ValidEnd       *time.Time `gorm:"type:datetime" json:"valid_end"`
	ValidDays      int         `gorm:"default:null" json:"valid_days"` // 领取后N天有效
	ApplicableType uint8      `gorm:"type:tinyint;not null;default:0" json:"applicable_type"` // 0-全品 1-指定类目 2-指定商品 3-指定店铺
	ApplicableIDs  string     `gorm:"type:varchar(512);default:''" json:"applicable_ids"`
	ShopID        *uint64    `gorm:"default:null" json:"shop_id"`
	Status        uint8      `gorm:"type:tinyint;not null;default:1" json:"status"`
	CreatedAt     time.Time  `gorm:"type:datetime;not null;default:current_timestamp" json:"created_at"`
}

func (Coupon) TableName() string { return "coupons" }

type UserCoupon struct {
	ID         uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint64     `gorm:"index;not null" json:"user_id"`
	CouponID  uint64     `gorm:"not null" json:"coupon_id"`
	CouponCode string    `gorm:"type:varchar(32);not null" json:"coupon_code"`
	Status    uint8      `gorm:"type:tinyint;not null;default:0" json:"status"`
	ReceivedAt time.Time `gorm:"type:datetime;not null" json:"received_at"`
	UsedAt    *time.Time `gorm:"type:datetime" json:"used_at"`
	UsedOrderNo string   `gorm:"type:varchar(32);default:''" json:"used_order_no"`
	ExpireDate  time.Time `gorm:"type:date;not null" json:"expire_date"`
}

func (UserCoupon) TableName() string { return "user_coupons" }

type Promotion struct {
	ID        uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string     `gorm:"type:varchar(64);not null" json:"name"`
	Type      string     `gorm:"type:varchar(16);not null" json:"type"` // full_reduce/flash_seckill/group_buy/tier_price/pre_sell/discount
	StartTime time.Time  `gorm:"type:datetime;not null" json:"start_time"`
	EndTime   time.Time  `gorm:"type:datetime;not null" json:"end_time"`
	Status    uint8      `gorm:"type:tinyint;not null;default:0" json:"status"` // 0-未开始 1-进行中 2-已结束
	Rules     string     `gorm:"type:json;not null" json:"rules"` // JSON rules
	CreatedAt time.Time  `gorm:"type:datetime;not null;default:current_timestamp" json:"created_at"`
}

func (Promotion) TableName() string { return "promotions" }

type PromotionSku struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	PromotionID  uint64    `gorm:"index;not null" json:"promotion_id"`
	SkuID        uint64    `gorm:"index;not null" json:"sku_id"`
	ShopID       uint64    `gorm:"not null" json:"shop_id"`
	PromoPrice   float64   `gorm:"type:decimal(12,2);not null" json:"promo_price"`
	StockLimit   int       `gorm:"not null" json:"stock_limit"`
	SoldCount    int       `gorm:"not null;default:0" json:"sold_count"`
}

func (PromotionSku) TableName() string { return "promotion_skus" }
