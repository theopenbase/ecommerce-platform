package model

import "time"

// FrozenStock 冻结库存记录（用于超卖补偿和超时回滚）
type FrozenStock struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	SkuID     uint64    `gorm:"not null;index" json:"sku_id"`
	OrderNo   string    `gorm:"type:varchar(32);not null" json:"order_no"`
	Quantity  int       `gorm:"not null" json:"quantity"`
	State     uint8     `gorm:"type:tinyint;not null;default:0" json:"state"` // 0=冻结中 1=已扣减(下单成功) 2=已回滚(超时/取消)
	CreatedAt time.Time `gorm:"type:datetime;not null;default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:datetime;not null;default:current_timestamp on update current_timestamp" json:"updated_at"`
}

func (FrozenStock) TableName() string { return "frozen_stocks" }
