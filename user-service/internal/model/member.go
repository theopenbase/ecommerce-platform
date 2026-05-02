package model

import (
	"time"
)

type MemberLevel struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Level     uint8     `gorm:"uniqueIndex;not null" json:"level"`
	Name      string    `gorm:"type:varchar(16);not null" json:"name"`
	Threshold float64   `gorm:"type:decimal(12,2);not null" json:"threshold"`
	Rights    string    `gorm:"type:json" json:"rights"` // JSON string of rights
	CreatedAt time.Time `gorm:"type:datetime;not null;default:current_timestamp" json:"created_at"`
}

func (MemberLevel) TableName() string {
	return "member_levels"
}

type Member struct {
	ID             uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID         uint64    `gorm:"uniqueIndex;not null" json:"user_id"`
	Level          uint8     `gorm:"type:tinyint;not null;default:0" json:"level"`
	TotalSpend     float64   `gorm:"type:decimal(14,2);not null;default:0" json:"total_spend"`
	GrowthValue    int64     `gorm:"not null;default:0" json:"growth_value"`
	GraceEndDate   *Date     `gorm:"type:date" json:"grace_end_date"`
	UpgradedAt     *time.Time `gorm:"type:datetime" json:"upgraded_at"`
	CreatedAt      time.Time `gorm:"type:datetime;not null;default:current_timestamp" json:"created_at"`
	UpdatedAt      time.Time `gorm:"type:datetime;not null;default:current_timestamp on update current_timestamp" json:"updated_at"`
}

func (Member) TableName() string {
	return "members"
}

type Date struct {
	time.Time
}

func (d *Date) Scan(value interface{}) error {
	if v, ok := value.(time.Time); ok {
		d.Time = v
		return nil
	}
	return nil
}

type MemberProfile struct {
	UserID      uint64  `json:"user_id"`
	Level       uint8   `json:"level"`
	LevelName   string  `json:"level_name"`
	TotalSpend  float64 `json:"total_spend"`
	GrowthValue int64   `json:"growth_value"`
	NextThreshold float64 `json:"next_threshold"`
	NextGrowthNeeded int64 `json:"next_growth_needed"`
}

// PointsAccount 积分账户
type PointsAccount struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      uint64    `gorm:"uniqueIndex;not null" json:"user_id"`
	Balance     int64     `gorm:"not null;default:0" json:"balance"`
	TotalEarned int64     `gorm:"not null;default:0" json:"total_earned"`
	TotalSpent  int64     `gorm:"not null;default:0" json:"total_spent"`
	ExpireDate  *Date     `gorm:"type:date" json:"expire_date"`
	CreatedAt   time.Time `gorm:"type:datetime;not null;default:current_timestamp" json:"created_at"`
	UpdatedAt   time.Time `gorm:"type:datetime;not null;default:current_timestamp on update current_timestamp" json:"updated_at"`
}

func (PointsAccount) TableName() string {
	return "points_accounts"
}

// PointsLog 积分流水
type PointsLog struct {
	ID            uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID        uint64    `gorm:"index;not null" json:"user_id"`
	OrderNo       string    `gorm:"type:varchar(32);index;default:''" json:"order_no"`
	Type          string    `gorm:"type:varchar(16);not null" json:"type"` // order/sign/eval/redeem/expire/refund
	Points        int64     `gorm:"not null" json:"points"`
	BalanceBefore int64     `gorm:"not null" json:"balance_before"`
	BalanceAfter  int64     `gorm:"not null" json:"balance_after"`
	ExpireDate    *Date     `gorm:"type:date" json:"expire_date"`
	CreatedAt     time.Time `gorm:"type:datetime;not null;default:current_timestamp" json:"created_at"`
}

func (PointsLog) TableName() string {
	return "points_logs"
}

// ReceiverAddress 收货地址
type ReceiverAddress struct {
	ID            uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID        uint64    `gorm:"index;not null" json:"user_id"`
	Receiver      string    `gorm:"type:varchar(32);not null" json:"receiver"`
	Mobile        string    `gorm:"type:varchar(20);not null" json:"mobile"`
	ProvinceCode  string    `gorm:"type:varchar(10);not null" json:"province_code"`
	CityCode      string    `gorm:"type:varchar(10);not null" json:"city_code"`
	DistrictCode  string    `gorm:"type:varchar(10);not null" json:"district_code"`
	ProvinceName  string    `gorm:"type:varchar(32);not null" json:"province_name"`
	CityName      string    `gorm:"type:varchar(32);not null" json:"city_name"`
	DistrictName  string    `gorm:"type:varchar(32);not null" json:"district_name"`
	Detail        string    `gorm:"type:varchar(256);not null" json:"detail"`
	Tag           string    `gorm:"type:varchar(16);default:''" json:"tag"`
	IsDefault     uint8     `gorm:"type:tinyint;not null;default:0" json:"is_default"`
	CreatedAt     time.Time `gorm:"type:datetime;not null;default:current_timestamp" json:"created_at"`
	UpdatedAt     time.Time `gorm:"type:datetime;not null;default:current_timestamp on update current_timestamp" json:"updated_at"`
}

func (ReceiverAddress) TableName() string {
	return "receiver_addresses"
}
