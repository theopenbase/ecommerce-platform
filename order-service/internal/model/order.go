package model

import (
	"fmt"
	"math/rand"
	"time"
)

// ============ 购物车 ============

type Cart struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint64    `gorm:"index;not null" json:"user_id"`
	SkuID     uint64    `gorm:"not null" json:"sku_id"`
	SpuID     uint64    `gorm:"not null" json:"spu_id"`
	ShopID    uint64    `gorm:"not null" json:"shop_id"`
	Quantity  int       `gorm:"not null;default:1" json:"quantity"`
	Checked   uint8     `gorm:"type:tinyint;not null;default:1" json:"checked"` // 1-选中 0-未选中
	// SKU 快照（下单时复制商品信息，保留下单时刻价格）
	SkuCode    string  `gorm:"type:varchar(64);not null" json:"sku_code"`
	Title     string  `gorm:"type:varchar(80);not null" json:"title"`
	SkuAttrs   string  `gorm:"type:varchar(512);not null" json:"sku_attrs"` // JSON
	PriceTag  float64 `gorm:"type:decimal(12,2);not null" json:"price_tag"`
	PriceSell float64 `gorm:"type:decimal(12,2);not null" json:"price_sell"`
	CreatedAt time.Time `gorm:"type:datetime;not null;default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:datetime;not null;default:current_timestamp on update current_timestamp" json:"updated_at"`
}

func (Cart) TableName() string { return "carts" }

// ============ 父订单 ============

type ParentOrder struct {
	ID            uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderNo       string     `gorm:"type:varchar(32);uniqueIndex;not null" json:"order_no"`
	BuyerID       uint64     `gorm:"index;not null" json:"buyer_id"`
	TotalAmount   float64    `gorm:"type:decimal(14,2);not null" json:"total_amount"`
	FreightAmount float64    `gorm:"type:decimal(14,2);not null;default:0" json:"freight_amount"`
	DiscountAmt   float64    `gorm:"type:decimal(14,2);not null;default:0" json:"discount_amount"`
	PayAmount     float64    `gorm:"type:decimal(14,2);not null" json:"pay_amount"`
	Status        uint8      `gorm:"type:tinyint;not null;index" json:"status"`
	PayTime       *time.Time `gorm:"type:datetime" json:"pay_time"`
	DeliveryTime  *time.Time `gorm:"type:datetime" json:"delivery_time"`
	ReceiveTime   *time.Time `gorm:"type:datetime" json:"receive_time"`
	FinishTime    *time.Time `gorm:"type:datetime" json:"finish_time"`
	CancelTime    *time.Time `gorm:"type:datetime" json:"cancel_time"`
	CancelReason  string     `gorm:"type:varchar(256);default:''" json:"cancel_reason"`
	CreatedAt     time.Time  `gorm:"type:datetime;not null;default:current_timestamp" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"type:datetime;not null;default:current_timestamp on update current_timestamp" json:"updated_at"`
}

func (ParentOrder) TableName() string { return "parent_orders" }

// ============ 子订单 ============

type SubOrder struct {
	ID             uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	SubOrderNo     string     `gorm:"type:varchar(32);uniqueIndex;not null" json:"sub_order_no"`
	ParentOrderNo  string     `gorm:"index;not null" json:"parent_order_no"`
	BuyerID        uint64     `gorm:"index;not null" json:"buyer_id"`
	ShopID         uint64     `gorm:"index;not null" json:"shop_id"`
	Status         uint8      `gorm:"type:tinyint;not null" json:"status"`
	FreightAmount  float64    `gorm:"type:decimal(14,2);not null;default:0" json:"freight_amount"`
	DiscountAmt    float64    `gorm:"type:decimal(14,2);not null;default:0" json:"discount_amount"`
	PayAmount      float64    `gorm:"type:decimal(14,2);not null" json:"pay_amount"`
	InvoiceType    uint8      `gorm:"type:tinyint;default:0" json:"invoice_type"`
	InvoiceTitle   string     `gorm:"type:varchar(128);default:''" json:"invoice_title"`
	InvoiceTaxNo   string     `gorm:"type:varchar(32);default:''" json:"invoice_tax_no"`
	LogisticsNo    string     `gorm:"type:varchar(64);default:''" json:"logistics_no"`
	LogisticsCo    string     `gorm:"type:varchar(32);default:''" json:"logistics_company"`
	CreatedAt      time.Time  `gorm:"type:datetime;not null;default:current_timestamp" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"type:datetime;not null;default:current_timestamp on update current_timestamp" json:"updated_at"`
}

func (SubOrder) TableName() string { return "sub_orders" }

// ============ 订单商品项 ============

type OrderItem struct {
	ID            uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	SubOrderNo    string    `gorm:"index;not null" json:"sub_order_no"`
	SkuID         uint64    `gorm:"not null" json:"sku_id"`
	SpuID         uint64    `gorm:"not null" json:"spu_id"`
	SkuCode       string    `gorm:"type:varchar(64);not null" json:"sku_code"`
	Title         string    `gorm:"type:varchar(80);not null" json:"title"`
	SkuAttrs      string    `gorm:"type:varchar(512);not null" json:"sku_attrs"` // JSON
	PriceTag      float64   `gorm:"type:decimal(12,2);not null" json:"price_tag"`
	PriceSell     float64   `gorm:"type:decimal(12,2);not null" json:"price_sell"`
	Quantity      int       `gorm:"not null" json:"quantity"`
	DiscountAmt   float64   `gorm:"type:decimal(12,2);not null;default:0" json:"discount_amount"`
	ItemTotal     float64   `gorm:"type:decimal(14,2);not null" json:"item_total"`
	RefundStatus  uint8     `gorm:"type:tinyint;not null;default:0" json:"refund_status"`
	CreatedAt     time.Time `gorm:"type:datetime;not null;default:current_timestamp" json:"created_at"`
}

func (OrderItem) TableName() string { return "order_items" }

// ============ 订单地址快照 ============

type OrderAddress struct {
	ID           uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderNo      string `gorm:"uniqueIndex;not null" json:"order_no"`
	Receiver     string `gorm:"type:varchar(32);not null" json:"receiver"`
	Mobile       string `gorm:"type:varchar(20);not null" json:"mobile"`
	ProvinceCode string `gorm:"type:varchar(10);not null" json:"province_code"`
	CityCode     string `gorm:"type:varchar(10);not null" json:"city_code"`
	DistrictCode string `gorm:"type:varchar(10);not null" json:"district_code"`
	ProvinceName string `gorm:"type:varchar(32);not null" json:"province_name"`
	CityName     string `gorm:"type:varchar(32);not null" json:"city_name"`
	DistrictName string `gorm:"type:varchar(32);not null" json:"district_name"`
	Detail       string `gorm:"type:varchar(256);not null" json:"detail"`
}

func (OrderAddress) TableName() string { return "order_addresses" }

// ============ 订单操作日志 ============

type OrderActionLog struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderNo   string    `gorm:"index;not null" json:"order_no"`
	Action    string    `gorm:"type:varchar(32);not null" json:"action"`
	Operator  string    `gorm:"type:varchar(32);default:''" json:"operator"`
	Note      string    `gorm:"type:varchar(256);default:''" json:"note"`
	CreatedAt time.Time `gorm:"type:datetime;not null;default:current_timestamp" json:"created_at"`
}

func (OrderActionLog) TableName() string { return "order_action_logs" }

// ============ 常量 ============

const (
	OrderStatusPendingPayment uint8 = 0 // 待付款
	OrderStatusCancelled     uint8 = 1 // 已取消
	OrderStatusPaid         uint8 = 2 // 已付款/待发货
	OrderStatusDelivered    uint8 = 3 // 已发货/待收货
	OrderStatusReceived     uint8 = 4 // 已收货/待评价
	OrderStatusCompleted    uint8 = 5 // 已完成
	OrderStatusDispute      uint8 = 6 // 维权中
)

var OrderStatusText = map[uint8]string{
	0: "待付款",
	1: "已取消",
	2: "待发货",
	3: "待收货",
	4: "已收货",
	5: "已完成",
	6: "维权中",
}

const (
	RefundStatusNone    uint8 = 0 // 无退款
	RefundStatusPartial uint8 = 1 // 部分退款
	RefundStatusFull   uint8 = 2 // 全额退款
)

// GenOrderNo 生成订单号：8位日期+6位随机+2位校验
func GenOrderNo() string {
	prefix := time.Now().Format("20060102")
	seq := time.Now().UnixNano() % 1000000
	r := rand.Intn(100)
	return fmt.Sprintf("%s%06d%02d", prefix, seq, r)
}
