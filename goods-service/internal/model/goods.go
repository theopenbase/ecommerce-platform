package model

import (
	"encoding/json"
	"time"
)

// ============ 类目 ============

type Category struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"type:varchar(32);not null" json:"name"`
	ParentID  uint64    `gorm:"not null;default:0;index" json:"parent_id"`
	Level     uint8     `gorm:"type:tinyint;not null" json:"level"`
	Path      string    `gorm:"type:varchar(128);not null;default:''" json:"path"`
	Sort      int       `gorm:"not null;default:0" json:"sort"`
	Status    uint8     `gorm:"type:tinyint;not null;default:1" json:"status"`
	IsLeaf    uint8     `gorm:"type:tinyint;not null;default:0" json:"is_leaf"`
	CreatedAt time.Time `gorm:"type:datetime;not null;default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:datetime;not null;default:current_timestamp on update current_timestamp" json:"updated_at"`
}

func (Category) TableName() string { return "categories" }

// CategoryNode 前端展示用树节点
type CategoryNode struct {
	ID       uint64          `json:"id"`
	Name     string          `json:"name"`
	ParentID uint64          `json:"parent_id"`
	Level    uint8           `json:"level"`
	Sort     int             `json:"sort"`
	Children []*CategoryNode `json:"children,omitempty"`
}

// ============ 品牌 ============

type Brand struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"type:varchar(64);not null;unique" json:"name"`
	LogoURL   string    `gorm:"type:varchar(512);default:''" json:"logo_url"`
	Status    uint8     `gorm:"type:tinyint;not null;default:1" json:"status"`
	CreatedAt time.Time `gorm:"type:datetime;not null;default:current_timestamp" json:"created_at"`
}

func (Brand) TableName() string { return "brands" }

// ============ SPU ============

type Spu struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	SpuCode     string    `gorm:"type:varchar(32);not null;unique" json:"spu_code"`
	Title       string    `gorm:"type:varchar(80);not null" json:"title"`
	ShortDesc   string    `gorm:"type:varchar(200);not null" json:"short_desc"`
	BrandID     uint64    `gorm:"not null;index" json:"brand_id"`
	CategoryID  uint64    `gorm:"not null;index" json:"category_id"`
	Unit        string    `gorm:"type:varchar(16);not null" json:"unit"`
	Origin      string    `gorm:"type:varchar(64);default:''" json:"origin"`
	Status      uint8     `gorm:"type:tinyint;not null;default:0" json:"status"` // 0-待上架 1-上架 2-下架 3-归档
	ShopID      uint64    `gorm:"not null;index" json:"shop_id"`
	AuditStatus uint8     `gorm:"type:tinyint;not null;default:0" json:"audit_status"` // 0-待审核 1-通过 2-拒绝
	CreatedAt   time.Time `gorm:"type:datetime;not null;default:current_timestamp" json:"created_at"`
	UpdatedAt   time.Time `gorm:"type:datetime;not null;default:current_timestamp on update current_timestamp" json:"updated_at"`
}

func (Spu) TableName() string { return "spus" }

// SpuExt SPU扩展（描述图文等）
type SpuExt struct {
	ID          uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	SpuID       uint64 `gorm:"uniqueIndex;not null" json:"spu_id"`
	Description string `gorm:"type:text" json:"description"` // 富文本描述
	Images      string `gorm:"type:varchar(2048);default:''" json:"images"` // JSON 数组 ["url1","url2"]
}

func (SpuExt) TableName() string { return "spu_exts" }

// ============ SKU ============

type Sku struct {
	ID            uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	SkuCode       string    `gorm:"type:varchar(64);not null;unique" json:"sku_code"`
	SpuID         uint64    `gorm:"not null;index" json:"spu_id"`
	ShopID        uint64    `gorm:"not null;index" json:"shop_id"`
	PriceTag      float64   `gorm:"type:decimal(12,2);not null" json:"price_tag"`      // 吊牌价
	PriceSell     float64   `gorm:"type:decimal(12,2);not null" json:"price_sell"`    // 销售价
	PriceCost     float64   `gorm:"type:decimal(12,2)" json:"price_cost"`             // 成本价
	Stock         int       `gorm:"not null;default:0" json:"stock"`                  // 库存
	StockWarn     int       `gorm:"default:0" json:"stock_warn"`                     // 库存预警值
	FreightID     uint64    `gorm:"not null" json:"freight_id"`                      // 运费模板
	DeliveryRegion string   `gorm:"type:varchar(512);not null" json:"delivery_region"`
	DeliveryTime   uint8     `gorm:"type:tinyint;not null" json:"delivery_time"`      // 发货时间（小时）
	Status        uint8     `gorm:"type:tinyint;not null;default:0" json:"status"`   // 0-待上架 1-上架 2-下架 3-售罄 4-归档
	CreatedAt     time.Time `gorm:"type:datetime;not null;default:current_timestamp" json:"created_at"`
	UpdatedAt     time.Time `gorm:"type:datetime;not null;default:current_timestamp on update current_timestamp" json:"updated_at"`
}

func (Sku) TableName() string { return "skus" }

// ============ SKU 属性 ============

type SkuAttr struct {
	ID         uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	SkuID      uint64 `gorm:"not null;index" json:"sku_id"`
	AttrName   string `gorm:"type:varchar(32);not null" json:"attr_name"`
	AttrValue  string `gorm:"type:varchar(128);not null" json:"attr_value"`
}

func (SkuAttr) TableName() string { return "sku_attrs" }

// ============ SKU 图片 ============

type SkuImage struct {
	ID     uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	SkuID  uint64 `gorm:"not null;index" json:"sku_id"`
	URL    string `gorm:"type:varchar(512);not null" json:"url"`
	IsMain uint8  `gorm:"type:tinyint;not null;default:0" json:"is_main"` // 1-主图
	Sort   uint8  `gorm:"not null;default:0" json:"sort"`
}

func (SkuImage) TableName() string { return "sku_images" }

// ============ 类目属性模板 ============

type CategoryAttrTemplate struct {
	ID          uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	CategoryID  uint64 `gorm:"not null;index" json:"category_id"`
	AttrName    string `gorm:"type:varchar(32);not null" json:"attr_name"`
	AttrType    string `gorm:"type:varchar(16);not null" json:"attr_type"` // text/number/boolean/multi
	IsRequired  uint8  `gorm:"type:tinyint;not null;default:0" json:"is_required"`
	Sort        int    `gorm:"not null;default:0" json:"sort"`
}

func (CategoryAttrTemplate) TableName() string { return "category_attr_templates" }

// ============ 销售属性名（SPU 共享）============

type SpuAttrName struct {
	ID         uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	SpuID      uint64 `gorm:"not null;index" json:"spu_id"`
	AttrName   string `gorm:"type:varchar(32);not null" json:"attr_name"`
	Sort       uint8  `gorm:"not null;default:0" json:"sort"`
}

func (SpuAttrName) TableName() string { return "spu_attr_names" }

// 销售属性值（跨 SKU 共享）
type SpuAttrValue struct {
	ID         uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	AttrNameID uint64 `gorm:"not null;index" json:"attr_name_id"`
	AttrValue  string `gorm:"type:varchar(128);not null" json:"attr_value"`
	Sort       uint8  `gorm:"not null;default:0" json:"sort"`
}

func (SpuAttrValue) TableName() string { return "spu_attr_values" }

// ============ 搜索索引结构 ============

type GoodsSearchDoc struct {
	SpuID       uint64  `json:"spu_id"`
	SkuID       uint64  `json:"sku_id"`
	Title       string  `json:"title"`
	ShortDesc   string  `json:"short_desc"`
	BrandID     uint64  `json:"brand_id"`
	BrandName   string  `json:"brand_name"`
	CategoryID  uint64  `json:"category_id"`
	CategoryPath string  `json:"category_path"`
	ShopID      uint64  `json:"shop_id"`
	PriceSell   float64 `json:"price_sell"`
	PriceTag    float64 `json:"price_tag"`
	Stock       int     `json:"stock"`
	SalesCount  int     `json:"sales_count"`
	Status      uint8   `json:"status"`
	MainImage   string  `json:"main_image"`
	Attrs       string  `json:"attrs"` // 规格属性 JSON
}

// ToJSON serializes search doc
func (d *GoodsSearchDoc) ToJSON() ([]byte, error) {
	return json.Marshal(d)
}
