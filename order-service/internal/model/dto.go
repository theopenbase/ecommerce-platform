package model

// ============ 请求 DTO ============

// 购物车
type CartAddRequest struct {
	SkuID    uint64 `json:"sku_id" binding:"required"`
	Quantity int    `json:"quantity" binding:"required,gt=0"`
}

type CartUpdateRequest struct {
	Quantity int `json:"quantity" binding:"required,gte=0"`
}

type CartBatchRequest struct {
	Items []CartItemRequest `json:"items" binding:"required,min=1"`
}

type CartItemRequest struct {
	SkuID    uint64 `json:"sku_id" binding:"required"`
	Quantity int    `json:"quantity" binding:"required,gt=0"`
}

// 确认订单
type ConfirmOrderRequest struct {
	AddressID    uint64   `json:"address_id" binding:"required"`
	Items        []uint64 `json:"items" binding:"required,min=1"` // cart_item_ids
	InvoiceType  uint8    `json:"invoice_type"`
	InvoiceTitle string   `json:"invoice_title"`
	InvoiceTaxNo string   `json:"invoice_tax_no"`
	CouponIDs    []uint64 `json:"coupon_ids"`
	Remark       string   `json:"remark"`
}

// 订单查询
type OrderListQuery struct {
	Status   uint8 `form:"status"`
	Page     int   `form:"page,default=1"`
	PageSize int   `form:"page_size,default=20"`
}

// 订单操作
type CancelOrderRequest struct {
	Reason string `json:"reason"`
}

type ReceiveOrderRequest struct{}

type ApplyRefundRequest struct {
	SubOrderNo string  `json:"sub_order_no" binding:"required"`
	Type       uint8   `json:"type" binding:"required,oneof=1 2 3 4"` // 1-仅退款 2-退货退款 3-换货 4-维修
	Reason     string `json:"reason" binding:"required"`
	Amount     float64 `json:"amount"`
	Description string `json:"description"`
}

// ============ 响应 DTO ============

type CartItemResponse struct {
	CartID    uint64    `json:"cart_id"`
	SkuID     uint64    `json:"sku_id"`
	SpuID     uint64    `json:"spu_id"`
	ShopID    uint64    `json:"shop_id"`
	Title     string    `json:"title"`
	SkuAttrs  []string  `json:"sku_attrs"`
	PriceSell float64   `json:"price_sell"`
	PriceTag  float64   `json:"price_tag"`
	Stock     int       `json:"stock"`
	Quantity  int       `json:"quantity"`
	Checked   bool      `json:"checked"`
	MainImage string    `json:"main_image"`
}

type CartResponse struct {
	Items     []CartItemResponse `json:"items"`
	TotalAmount float64         `json:"total_amount"`
	TotalDiscount float64       `json:"total_discount"`
}

type OrderResponse struct {
	ParentOrder *ParentOrder   `json:"parent_order"`
	SubOrders   []SubOrderInfo `json:"sub_orders"`
	Address     *OrderAddress   `json:"address"`
}

type SubOrderInfo struct {
	SubOrder *SubOrder   `json:"sub_order"`
	Items    []OrderItem `json:"items"`
	ShopName string      `json:"shop_name"`
}

type OrderListResponse struct {
	Total int64             `json:"total"`
	Page  int               `json:"page"`
	Size  int               `json:"size"`
	Items []OrderListItem   `json:"items"`
}

type OrderListItem struct {
	OrderNo     string    `json:"order_no"`
	Status      uint8     `json:"status"`
	StatusText  string    `json:"status_text"`
	TotalAmount float64   `json:"total_amount"`
	PayAmount   float64   `json:"pay_amount"`
	ItemCount   int       `json:"item_count"`
	ShopName    string    `json:"shop_name"`
	ShopID      uint64    `json:"shop_id"`
	CreatedAt   string    `json:"created_at"`
}
