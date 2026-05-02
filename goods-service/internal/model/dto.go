package model

// ============ 请求 DTO ============

// 创建/编辑 SPU
type CreateSpuRequest struct {
	SpuCode    string            `json:"spu_code" binding:"required"`
	Title      string            `json:"title" binding:"required,max=80"`
	ShortDesc  string            `json:"short_desc" binding:"required,max=200"`
	BrandID    uint64            `json:"brand_id" binding:"required"`
	CategoryID uint64            `json:"category_id" binding:"required"`
	Unit       string            `json:"unit" binding:"required"`
	Origin     string            `json:"origin"`
	Attrs      []SpuAttrNameDTO  `json:"attrs"`
	Description string           `json:"description"`
	Images     []string         `json:"images"`
}

type SpuAttrNameDTO struct {
	AttrName  string             `json:"attr_name"`
	AttrValues []string          `json:"attr_values"`
}

// 创建/编辑 SKU
type CreateSkuRequest struct {
	SkuCode      string          `json:"sku_code" binding:"required"`
	PriceTag     float64         `json:"price_tag" binding:"required,gte=0"`
	PriceSell    float64         `json:"price_sell" binding:"required,gte=0"`
	PriceCost    float64         `json:"price_cost"`
	Stock        int             `json:"stock" binding:"gte=0"`
	StockWarn    int             `json:"stock_warn"`
	FreightID    uint64          `json:"freight_id" binding:"required"`
	DeliveryRegion string        `json:"delivery_region" binding:"required"`
	DeliveryTime uint8           `json:"delivery_time" binding:"required"`
	Attrs        []SkuAttrDTO    `json:"attrs"`
	Images       []SkuImageDTO   `json:"images"`
}

type SkuAttrDTO struct {
	AttrName  string `json:"attr_name"`
	AttrValue string `json:"attr_value"`
}

type SkuImageDTO struct {
	URL    string `json:"url"`
	IsMain bool   `json:"is_main"`
}

// 上下架
type UpdateSkuStatusRequest struct {
	Status uint8 `json:"status" binding:"required,oneof=0 1 2 3 4"`
}

// 商品列表查询
type GoodsListQuery struct {
	CategoryID   uint64   `form:"category_id"`
	BrandID      uint64   `form:"brand_id"`
	Keyword      string   `form:"keyword"`
	PriceMin     float64  `form:"price_min"`
	PriceMax     float64  `form:"price_max"`
	Status       uint8    `form:"status"`
	ShopID       uint64   `form:"shop_id"`
	Sort         string   `form:"sort"` // price_asc/price_desc/sales/new
	Page         int      `form:"page,default=1"`
	PageSize     int      `form:"page_size,default=20"`
}

// ============ 响应 DTO ============

type PageResponse struct {
	Total int64       `json:"total"`
	Page  int         `json:"page"`
	Size  int         `json:"size"`
	Items interface{} `json:"items"`
}

type SpuDetailResponse struct {
	Spu       *Spu         `json:"spu"`
	SpuExt    *SpuExt      `json:"spu_ext,omitempty"`
	Brand     *Brand       `json:"brand,omitempty"`
	Category  *Category    `json:"category,omitempty"`
	Skus      []SkuDetail  `json:"skus"`
	AttrNames []SpuAttrName `json:"attr_names"`
}

type SkuDetail struct {
	Sku    *Sku        `json:"sku"`
	Attrs  []SkuAttr   `json:"attrs"`
	Images []SkuImage  `json:"images"`
}

type GoodsListItem struct {
	SpuID       uint64  `json:"spu_id"`
	Title       string  `json:"title"`
	ShortDesc   string  `json:"short_desc"`
	BrandID     uint64  `json:"brand_id"`
	BrandName   string  `json:"brand_name"`
	CategoryID  uint64  `json:"category_id"`
	MinPrice    float64 `json:"min_price"`
	MaxPrice    float64 `json:"max_price"`
	MainImage   string  `json:"main_image"`
	Stock       int     `json:"stock"`
	SalesCount  int     `json:"sales_count"`
}

// SkuSnapshot SKU 快照（供其他服务下单时使用）
type SkuSnapshot struct {
	SkuID     uint64  `json:"id"`
	SpuID     uint64  `json:"spu_id"`
	ShopID    uint64  `json:"shop_id"`
	SkuCode   string  `json:"sku_code"`
	Title     string  `json:"title"`
	SkuAttrs  string  `json:"sku_attrs"`
	PriceTag  float64 `json:"price_tag"`
	PriceSell float64 `json:"price_sell"`
}
