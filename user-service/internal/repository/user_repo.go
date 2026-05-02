package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/ecommerce/user-service/internal/model"
)

type UserRepository struct {
	db    *gorm.DB
	cache *redis.Client
}

func NewUserRepository(db *gorm.DB, cache *redis.Client) *UserRepository {
	return &UserRepository{db: db, cache: cache}
}

// Create 创建用户
func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// FindByMobile 按手机号查找用户
func (r *UserRepository) FindByMobile(ctx context.Context, mobile string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("mobile = ?", mobile).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByID 按ID查找用户
func (r *UserRepository) FindByID(ctx context.Context, id uint64) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update 更新用户
func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// UpdatePassword 更新密码
func (r *UserRepository) UpdatePassword(ctx context.Context, userID uint64, hashedPwd string) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userID).Update("password", hashedPwd).Error
}

// UpdateLoginInfo 更新登录信息
func (r *UserRepository) UpdateLoginInfo(ctx context.Context, userID uint64, ip string) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"last_login": time.Now(),
		"last_ip":    ip,
	}).Error
}

// ============ 收货地址 ============

// CreateAddress 创建收货地址
func (r *UserRepository) CreateAddress(ctx context.Context, addr *model.ReceiverAddress) error {
	return r.db.WithContext(ctx).Create(addr).Error
}

// FindAddressesByUserID 查找用户所有收货地址
func (r *UserRepository) FindAddressesByUserID(ctx context.Context, userID uint64) ([]model.ReceiverAddress, error) {
	var addrs []model.ReceiverAddress
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("is_default desc, created_at desc").Find(&addrs).Error
	return addrs, err
}

// FindAddressByID 查找收货地址
func (r *UserRepository) FindAddressByID(ctx context.Context, id, userID uint64) (*model.ReceiverAddress, error) {
	var addr model.ReceiverAddress
	err := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).First(&addr).Error
	if err != nil {
		return nil, err
	}
	return &addr, nil
}

// UpdateAddress 更新收货地址
func (r *UserRepository) UpdateAddress(ctx context.Context, addr *model.ReceiverAddress) error {
	return r.db.WithContext(ctx).Save(addr).Error
}

// DeleteAddress 删除收货地址
func (r *UserRepository) DeleteAddress(ctx context.Context, id, userID uint64) error {
	return r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&model.ReceiverAddress{}).Error
}

// ClearDefaultAddress 清除用户默认地址标记
func (r *UserRepository) ClearDefaultAddress(ctx context.Context, userID uint64) error {
	return r.db.WithContext(ctx).Model(&model.ReceiverAddress{}).Where("user_id = ?", userID).Update("is_default", 0).Error
}

// CountAddress 统计用户地址数量
func (r *UserRepository) CountAddress(ctx context.Context, userID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.ReceiverAddress{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

// ============ 登录日志 ============

// CreateLoginLog 创建登录日志
func (r *UserRepository) CreateLoginLog(ctx context.Context, log *model.LoginLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// FindLoginLogsByUserID 查找用户登录日志
func (r *UserRepository) FindLoginLogsByUserID(ctx context.Context, userID uint64, limit int) ([]model.LoginLog, error) {
	var logs []model.LoginLog
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("login_time desc").Limit(limit).Find(&logs).Error
	return logs, err
}

// ============ 缓存验证码 ============

func (r *UserRepository) CacheSMSCode(ctx context.Context, mobile, code string) error {
	key := fmt.Sprintf("sms:code:%s", mobile)
	return r.cache.Set(ctx, key, code, 5*time.Minute).Err()
}

func (r *UserRepository) GetSMSCode(ctx context.Context, mobile string) (string, error) {
	key := fmt.Sprintf("sms:code:%s", mobile)
	return r.cache.Get(ctx, key).Result()
}

func (r *UserRepository) DelSMSCode(ctx context.Context, mobile string) error {
	key := fmt.Sprintf("sms:code:%s", mobile)
	return r.cache.Del(ctx, key).Err()
}

// CacheRefreshToken 缓存 refresh token
func (r *UserRepository) CacheRefreshToken(ctx context.Context, userID uint64, token string, ttl time.Duration) error {
	key := fmt.Sprintf("refresh:token:%d:%s", userID, token)
	return r.cache.Set(ctx, key, "1", ttl).Err()
}

// ValidateRefreshToken 验证 refresh token
func (r *UserRepository) ValidateRefreshToken(ctx context.Context, userID uint64, token string) (bool, error) {
	key := fmt.Sprintf("refresh:token:%d:%s", userID, token)
	result, err := r.cache.Exists(ctx, key).Result()
	return result > 0, err
}

// InvalidateRefreshToken 作废 refresh token
func (r *UserRepository) InvalidateRefreshToken(ctx context.Context, userID uint64, token string) error {
	key := fmt.Sprintf("refresh:token:%d:%s", userID, token)
	return r.cache.Del(ctx, key).Err()
}

// InvalidateAllUserTokens 作废用户所有 refresh tokens
func (r *UserRepository) InvalidateAllUserTokens(ctx context.Context, userID uint64) error {
	pattern := fmt.Sprintf("refresh:token:%d:*", userID)
	iter := r.cache.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		r.cache.Del(ctx, iter.Val())
	}
	return iter.Err()
}

// ============ 会员 ============

// FindMemberByUserID 查找会员信息
func (r *UserRepository) FindMemberByUserID(ctx context.Context, userID uint64) (*model.Member, error) {
	var member model.Member
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&member).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

// CreateMember 创建会员记录
func (r *UserRepository) CreateMember(ctx context.Context, member *model.Member) error {
	return r.db.WithContext(ctx).Create(member).Error
}

// UpdateMember 更新会员信息
func (r *UserRepository) UpdateMember(ctx context.Context, member *model.Member) error {
	return r.db.WithContext(ctx).Save(member).Error
}

// FindAllMemberLevels 查找所有会员等级
func (r *UserRepository) FindAllMemberLevels(ctx context.Context) ([]model.MemberLevel, error) {
	var levels []model.MemberLevel
	err := r.db.WithContext(ctx).Order("level asc").Find(&levels).Error
	return levels, err
}

// FindMemberLevelByLevel 按等级查找
func (r *UserRepository) FindMemberLevelByLevel(ctx context.Context, level uint8) (*model.MemberLevel, error) {
	var lv model.MemberLevel
	err := r.db.WithContext(ctx).Where("level = ?", level).First(&lv).Error
	if err != nil {
		return nil, err
	}
	return &lv, nil
}

// ============ 积分 ============

// FindPointsAccountByUserID 查找积分账户
func (r *UserRepository) FindPointsAccountByUserID(ctx context.Context, userID uint64) (*model.PointsAccount, error) {
	var account model.PointsAccount
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

// CreatePointsAccount 创建积分账户
func (r *UserRepository) CreatePointsAccount(ctx context.Context, account *model.PointsAccount) error {
	return r.db.WithContext(ctx).Create(account).Error
}

// UpdatePointsAccount 更新积分账户
func (r *UserRepository) UpdatePointsAccount(ctx context.Context, account *model.PointsAccount) error {
	return r.db.WithContext(ctx).Save(account).Error
}

// CreatePointsLog 创建积分流水
func (r *UserRepository) CreatePointsLog(ctx context.Context, log *model.PointsLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// FindPointsLogsByUserID 查找积分流水
func (r *UserRepository) FindPointsLogsByUserID(ctx context.Context, userID uint64, limit int) ([]model.PointsLog, error) {
	var logs []model.PointsLog
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at desc").Limit(limit).Find(&logs).Error
	return logs, err
}

// ============ Session/缓存 ============

func (r *UserRepository) CacheUser(ctx context.Context, user *model.User) error {
	key := fmt.Sprintf("user:info:%d", user.ID)
	data, _ := json.Marshal(user)
	return r.cache.Set(ctx, key, data, 10*time.Minute).Err()
}

func (r *UserRepository) GetCachedUser(ctx context.Context, userID uint64) (*model.User, error) {
	key := fmt.Sprintf("user:info:%d", userID)
	data, err := r.cache.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}
	var user model.User
	if err := json.Unmarshal(data, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) InvalidateUserCache(ctx context.Context, userID uint64) error {
	key := fmt.Sprintf("user:info:%d", userID)
	return r.cache.Del(ctx, key).Err()
}

// GenerateAuditID 生成审计ID
func GenerateAuditID() string {
	return uuid.New().String()
}

// UpdatePayPassword 定向更新支付密码（不走全量 Update，避免字段覆盖）
func (r *UserRepository) UpdatePayPassword(ctx context.Context, userID uint64, hashedPwd string) error {
	return r.db.Model(&model.User{}).Where("id = ?", userID).Update("pay_password", hashedPwd).Error
}
