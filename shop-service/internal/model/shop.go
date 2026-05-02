package model

import "time"

// 店铺类型
const (
	ShopTypeSelf     uint8 = 1 // 自营
	ShopTypeFlagship uint8 = 2 // 旗舰
	ShopTypeSpecial  uint8 = 3 // 专卖
	ShopTypeCategory uint8 = 4 // 专营
)

// 店铺状态
const (
	ShopStatusPending  uint8 = 0 // 待审核
	ShopStatusActive  uint8 = 1 // 正常
	ShopStatusBanned  uint8 = 2 // 封禁
	ShopStatusClosed  uint8 = 3 // 已清退
)

type Shop struct {
	ID           uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ShopCode     string     `gorm:"type:varchar(32);uniqueIndex;not null" json:"shop_code"`
	Name         string     `gorm:"type:varchar(64);not null" json:"name"`
	Type        uint8      `gorm:"type:tinyint;not null" json:"type"`
	OwnerID     uint64     `gorm:"not null" json:"owner_id"` // 店主用户ID
	Status      uint8      `gorm:"type:tinyint;not null;default:0" json:"status"`
	LogoURL     string     `gorm:"type:varchar(512);default:''" json:"logo_url"`
	BannerURL   string     `gorm:"type:varchar(512);default:''" json:"banner_url"`
	Description string     `gorm:"type:varchar(500);default:''" json:"description"`
	Province    string     `gorm:"type:varchar(32);default:''" json:"province"`
	City        string     `gorm:"type:varchar(32);default:''" json:"city"`
	DSRProduct  float64    `gorm:"type:decimal(3,2);default:0" json:"dsr_product"`  // 商品描述DSR
	DSRService  float64    `gorm:"type:decimal(3,2);default:0" json:"dsr_service"`  // 服务态度DSR
	DSRLogistics float64   `gorm:"type:decimal(3,2);default:0" json:"dsr_logistics"` // 物流DSR
	DSR Overall float64    `gorm:"-" json:"dsr_overall"` // 计算字段
	FollowerCount uint64   `gorm:"default:0" json:"follower_count"`
	CreatedAt   time.Time  `gorm:"type:datetime;not null;default:current_timestamp" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"type:datetime;not null;default:current_timestamp on update current_timestamp" json:"updated_at"`
}

func (Shop) TableName() string { return "shops" }

type ShopQualification struct {
	ID         uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ShopID    uint64     `gorm:"index;not null" json:"shop_id"`
	QualType  string     `gorm:"type:varchar(32);not null" json:"qual_type"`
	CertNo    string     `gorm:"type:varchar(64);default:''" json:"cert_no"`
	FrontURL  string     `gorm:"type:varchar(512);not null" json:"front_url"`
	BackURL   string     `gorm:"type:varchar(512);default:''" json:"back_url"`
	ExpiryDate *time.Time `gorm:"type:date" json:"expiry_date"`
	Status    uint8      `gorm:"type:tinyint;not null;default:0" json:"status"` // 0-待审 1-通过 2-拒绝
	AuditedAt *time.Time `gorm:"type:datetime" json:"audited_at"`
	AuditorID uint64     `gorm:"default:null" json:"auditor_id"`
}

func (ShopQualification) TableName() string { return "shop_qualifications" }

type ShopDeposit struct {
	ID          uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ShopID     uint64     `gorm:"index;not null" json:"shop_id"`
	Amount     float64    `gorm:"type:decimal(12,2);not null" json:"amount"`
	Status     uint8      `gorm:"type:tinyint;not null;default:0" json:"status"` // 0-冻结 1-可用 2-扣除中
	FreezeTime *time.Time `gorm:"type:datetime" json:"freeze_time"`
}

func (ShopDeposit) TableName() string { return "shop_deposits" }

type FreightTemplate struct {
	ID             uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ShopID        uint64    `gorm:"index;not null" json:"shop_id"`
	Name          string    `gorm:"type:varchar(32);not null" json:"name"`
	Type         uint8     `gorm:"type:tinyint;not null" json:"type"` // 1-按件 2-按重量 3-固定 4-地区计价
	IsFreeThreshold uint8   `gorm:"type:tinyint;not null;default:0" json:"is_free_threshold"`
	FreeAmount   float64   `gorm:"type:decimal(12,2);default:0" json:"free_amount"`
	FreeNum     int       `gorm:"default:0" json:"free_num"`
	CreatedAt   time.Time `gorm:"type:datetime;not null;default:current_timestamp" json:"created_at"`
}

func (FreightTemplate) TableName() string { return "freight_templates" }

type FreightRule struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	TemplateID  uint64    `gorm:"index;not null" json:"template_id"`
	ProvinceCodes string   `gorm:"type:varchar(256);not null" json:"province_codes"`
	FirstAmount float64   `gorm:"type:decimal(10,2);not null" json:"first_amount"`
	FirstNum    int       `gorm:"not null;default:1" json:"first_num"`
	AddAmount   float64   `gorm:"type:decimal(10,2);not null" json:"add_amount"`
	AddNum      int       `gorm:"not null;default:1" json:"add_num"`
}

func (FreightRule) TableName() string { return "freight_rules" }

type ShopDecoration struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ShopID   uint64    `gorm:"uniqueIndex;not null" json:"shop_id"`
	Layout   string    `gorm:"type:json;not null" json:"layout"` // 装修布局JSON
	CreatedAt time.Time `gorm:"type:datetime;not null;default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:datetime;not null;default:current_timestamp on update current_timestamp" json:"updated_at"`
}

func (ShopDecoration) TableName() string { return "shop_decorations" }
