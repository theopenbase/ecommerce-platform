package model

import (
	"time"
)

type User struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Mobile     string    `gorm:"type:varchar(20);uniqueIndex;not null" json:"mobile"`
	Nickname   string    `gorm:"type:varchar(32);not null" json:"nickname"`
	AvatarURL  string    `gorm:"type:varchar(512);default:''" json:"avatar_url"`
	Gender     uint8     `gorm:"type:tinyint;default:0" json:"gender"` // 0-unknown 1-male 2-female
	Status     uint8     `gorm:"type:tinyint;default:1" json:"status"` // 1-normal 0-banned
	LastLogin  time.Time `gorm:"type:datetime;default:null" json:"last_login"`
	LastIP     string    `gorm:"type:varchar(45);default:''" json:"last_ip"`
	Password    string    `gorm:"type:varchar(64);default:''" json:"-"` // hashed, never expose
	PayPassword string    `gorm:"type:varchar(64);default:''" json:"-"` // hashed, never expose
	CreatedAt  time.Time `gorm:"type:datetime;not null;default:current_timestamp" json:"created_at"`
	UpdatedAt  time.Time `gorm:"type:datetime;not null;default:current_timestamp on update current_timestamp" json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}

// UserProfile 用户详细信息（对外暴露）
type UserProfile struct {
	ID        uint64    `json:"id"`
	Mobile    string    `json:"mobile"`
	Nickname  string    `json:"nickname"`
	AvatarURL string    `json:"avatar_url"`
	Gender    uint8     `json:"gender"`
	Status    uint8     `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// ToProfile 转换为对外暴露的Profile
func (u *User) ToProfile() *UserProfile {
	return &UserProfile{
		ID:        u.ID,
		Mobile:    u.Mobile,
		Nickname:  u.Nickname,
		AvatarURL: u.AvatarURL,
		Gender:    u.Gender,
		Status:    u.Status,
		CreatedAt: u.CreatedAt,
	}
}

type LoginLog struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID       uint64    `gorm:"index;not null" json:"user_id"`
	IP           string    `gorm:"type:varchar(45);not null" json:"ip"`
	Device       string    `gorm:"type:varchar(128);default:''" json:"device"`
	Status       uint8     `gorm:"type:tinyint;not null" json:"status"` // 1-success 0-fail
	FailReason   string    `gorm:"type:varchar(128);default:''" json:"fail_reason"`
	LoginTime    time.Time `gorm:"type:datetime;not null" json:"login_time"`
}

func (LoginLog) TableName() string {
	return "login_logs"
}
