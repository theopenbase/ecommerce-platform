package model

import "time"

// 支付渠道
const (
	ChannelAlipay  = "alipay"
	ChannelWechat  = "wechat"
	ChannelBank   = "bank"
	ChannelBalance = "balance"
)

// 支付流水状态
const (
	PayStatusPending  uint8 = 0 // 待支付
	PayStatusProcessing uint8 = 1 // 支付中
	PayStatusSuccess   uint8 = 2 // 成功
	PayStatusFailed    uint8 = 3 // 失败
	PayStatusRefunded  uint8 = 4 // 已退款
)

// 账户类型
const (
	AccountTypeBuyer  uint8 = 1
	AccountTypeSeller uint8 = 2
	AccountTypePlatform uint8 = 3
)

// 账务流水类型
const (
	AccountLogRecharge  = "recharge"
	AccountLogPay       = "pay"
	AccountLogRefund    = "refund"
	AccountLogSettle    = "settle"
	AccountLogFreeze    = "freeze"
	AccountLogUnfreeze  = "unfreeze"
)

type PaymentTransaction struct {
	ID              uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TransNo         string     `gorm:"type:varchar(64);uniqueIndex;not null" json:"trans_no"`
	OrderNo         string     `gorm:"type:varchar(32);index;not null" json:"order_no"`
	SubOrderNo      string     `gorm:"type:varchar(32);default:''" json:"sub_order_no"`
	BuyerID         uint64     `gorm:"index;not null" json:"buyer_id"`
	Channel         string     `gorm:"type:varchar(16);not null" json:"channel"`
	ChannelTransNo  string     `gorm:"type:varchar(128);default:''" json:"channel_trans_no"`
	Amount          float64    `gorm:"type:decimal(12,2);not null" json:"amount"`
	Status          uint8      `gorm:"type:tinyint;not null;index" json:"status"`
	PayTime         *time.Time `gorm:"type:datetime" json:"pay_time"`
	ExpireTime      time.Time  `gorm:"type:datetime;not null" json:"expire_time"`
	CallbackRaw     string     `gorm:"type:text" json:"callback_raw"`
	CreatedAt       time.Time  `gorm:"type:datetime;not null;default:current_timestamp" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"type:datetime;not null;default:current_timestamp on update current_timestamp" json:"updated_at"`
}

func (PaymentTransaction) TableName() string { return "payment_transactions" }

type Account struct {
	ID            uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID        uint64    `gorm:"uniqueIndex:uk_user_type;not null" json:"user_id"`
	UserType      uint8     `gorm:"uniqueIndex:uk_user_type;not null" json:"user_type"`
	Balance       float64   `gorm:"type:decimal(14,2);not null;default:0" json:"balance"`
	FreezeBalance float64   `gorm:"type:decimal(14,2);not null;default:0" json:"freeze_balance"`
	PasswordPay   string    `gorm:"type:varchar(64);default:''" json:"-"`
	Status        uint8     `gorm:"type:tinyint;not null;default:1" json:"status"` // 1-正常 0-冻结
	CreatedAt     time.Time `gorm:"type:datetime;not null;default:current_timestamp" json:"created_at"`
	UpdatedAt     time.Time `gorm:"type:datetime;not null;default:current_timestamp on update current_timestamp" json:"updated_at"`
}

func (Account) TableName() string { return "accounts" }

type AccountLog struct {
	ID             uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	AccountID      uint64    `gorm:"index;not null" json:"account_id"`
	TransNo        string    `gorm:"type:varchar(64);not null" json:"trans_no"`
	OrderNo        string    `gorm:"type:varchar(32);default:''" json:"order_no"`
	Type           string    `gorm:"type:varchar(16);not null" json:"type"`
	Amount         float64   `gorm:"type:decimal(12,2);not null" json:"amount"`
	BalanceBefore  float64   `gorm:"type:decimal(14,2);not null" json:"balance_before"`
	BalanceAfter   float64   `gorm:"type:decimal(14,2);not null" json:"balance_after"`
	Note           string    `gorm:"type:varchar(256);default:''" json:"note"`
	CreatedAt      time.Time `gorm:"type:datetime;not null;default:current_timestamp" json:"created_at"`
}

func (AccountLog) TableName() string { return "account_logs" }

type RefundRecord struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	RefundNo     string    `gorm:"type:varchar(32);uniqueIndex;not null" json:"refund_no"`
	TransNo      string    `gorm:"type:varchar(64);index;not null" json:"trans_no"`
	OrderNo      string    `gorm:"type:varchar(32);not null" json:"order_no"`
	BuyerID      uint64    `gorm:"index;not null" json:"buyer_id"`
	Amount       float64   `gorm:"type:decimal(12,2);not null" json:"amount"`
	Status       uint8     `gorm:"type:tinyint;not null;default:0" json:"status"` // 0-处理中 1-成功 2-失败
	Reason       string    `gorm:"type:varchar(256);default:''" json:"reason"`
	ProcessTime  *time.Time `gorm:"type:datetime" json:"process_time"`
	CreatedAt    time.Time  `gorm:"type:datetime;not null;default:current_timestamp" json:"created_at"`
}

func (RefundRecord) TableName() string { return "refund_records" }
