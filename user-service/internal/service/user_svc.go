package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/ecommerce/user-service/internal/model"
	"github.com/ecommerce/user-service/internal/pkg/jwt"
	"github.com/ecommerce/user-service/internal/pkg/smc"
	"github.com/ecommerce/user-service/internal/repository"
)

type UserService struct {
	repo      *repository.UserRepository
	jwtMgr    *jwt.Manager
	smsProv   smc.SMSProvider
	accessTTL int64
	refreshTTL int64
}

func NewUserService(repo *repository.UserRepository, jwtMgr *jwt.Manager, smsProv smc.SMSProvider, accessTTL, refreshTTL int64) *UserService {
	return &UserService{
		repo:      repo,
		jwtMgr:    jwtMgr,
		smsProv:   smsProv,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

// ============ 认证 ============

// SendCode 发送验证码
func (s *UserService) SendCode(ctx context.Context, mobile string) error {
	code := fmt.Sprintf("%06d", rand.Intn(1000000))
	if err := s.repo.CacheSMSCode(ctx, mobile, code); err != nil {
		return err
	}
	return s.smsProv.SendCode(ctx, mobile, code)
}

// Register 用户注册
func (s *UserService) Register(ctx context.Context, req *model.RegisterRequest) (*model.UserResponse, error) {
	// 1. 验证码校验
	storedCode, err := s.repo.GetSMSCode(ctx, req.Mobile)
	if err != nil || storedCode != req.Code {
		return nil, ErrInvalidCode
	}
	defer s.repo.DelSMSCode(ctx, req.Mobile)

	// 2. 检查手机号是否已注册
	existing, _ := s.repo.FindByMobile(ctx, req.Mobile)
	if existing != nil && existing.ID > 0 {
		return nil, ErrUserExists
	}

	// 3. 创建用户（无密码方式注册，密码可后续设置）
	user := &model.User{
		Mobile:   req.Mobile,
		Nickname: req.Nickname,
		Status:   1,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	// 4. 初始化会员记录
	member := &model.Member{
		UserID: user.ID,
		Level:  0,
	}
	s.repo.CreateMember(ctx, member)

	// 5. 初始化积分账户
	pointsAccount := &model.PointsAccount{
		UserID: user.ID,
		Balance: 0,
	}
	s.repo.CreatePointsAccount(ctx, pointsAccount)

	// 6. 生成 Token
	return s.issueTokens(ctx, user)
}

// Login 登录
func (s *UserService) Login(ctx context.Context, req *model.LoginRequest, ip string) (*model.UserResponse, error) {
	// 1. 验证码校验
	storedCode, err := s.repo.GetSMSCode(ctx, req.Mobile)
	if err != nil || storedCode != req.Code {
		// 验证码错误，记录失败日志（此时无 userID）
		s.repo.CreateLoginLog(ctx, &model.LoginLog{
			IP:        ip,
			Status:    0,
			FailReason: "invalid_code",
			LoginTime: time.Now(),
		})
		s.repo.DelSMSCode(ctx, req.Mobile)
		return nil, ErrInvalidCode
	}
	s.repo.DelSMSCode(ctx, req.Mobile)

	// 2. 查找用户
	user, err := s.repo.FindByMobile(ctx, req.Mobile)
	if err != nil {
		s.repo.CreateLoginLog(ctx, &model.LoginLog{
			IP:        ip,
			Status:    0,
			FailReason: "user_not_found",
			LoginTime: time.Now(),
		})
		return nil, ErrUserNotFound
	}

	if user.Status != 1 {
		s.repo.CreateLoginLog(ctx, &model.LoginLog{
			UserID:    user.ID,
			IP:        ip,
			Status:    0,
			FailReason: "account_banned",
			LoginTime: time.Now(),
		})
		return nil, errors.New("account banned")
	}

	// 3. 更新登录信息
	s.repo.UpdateLoginInfo(ctx, user.ID, ip)

	// 4. 记录成功登录日志
	s.repo.CreateLoginLog(ctx, &model.LoginLog{
		UserID:    user.ID,
		IP:        ip,
		Status:    1,
		LoginTime: time.Now(),
	})

	// 5. 生成 Token
	return s.issueTokens(ctx, user)
}

// issueTokens 生成并返回 Token 对
func (s *UserService) issueTokens(ctx context.Context, user *model.User) (*model.UserResponse, error) {
	accessToken, err := s.jwtMgr.GenerateAccessToken(user.ID, s.accessTTL)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtMgr.GenerateRefreshToken(user.ID, s.refreshTTL)
	if err != nil {
		return nil, err
	}

	// 缓存 refresh token
	s.repo.CacheRefreshToken(ctx, user.ID, refreshToken, time.Duration(s.refreshTTL)*time.Second)

	return &model.UserResponse{
		User: user.ToProfile(),
		Tokens: &model.TokenResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresIn:    s.accessTTL,
			TokenType:    "Bearer",
		},
	}, nil
}

// RefreshToken 刷新 Token
func (s *UserService) RefreshToken(ctx context.Context, refreshToken string) (*model.TokenResponse, error) {
	claims, err := s.jwtMgr.ValidateToken(refreshToken)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != "refresh" {
		return nil, errors.New("invalid token type")
	}

	// 检查 token 是否在黑名单（未作废）
	valid, err := s.repo.ValidateRefreshToken(ctx, claims.UserID, refreshToken)
	if err != nil || !valid {
		return nil, errors.New("token revoked")
	}

	// 生成新 access token
	accessToken, err := s.jwtMgr.GenerateAccessToken(claims.UserID, s.accessTTL)
	if err != nil {
		return nil, err
	}

	return &model.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    s.accessTTL,
		TokenType:    "Bearer",
	}, nil
}

// Logout 登出（作废 refresh token）
func (s *UserService) Logout(ctx context.Context, userID uint64, refreshToken string) error {
	return s.repo.InvalidateRefreshToken(ctx, userID, refreshToken)
}

// SetPassword 设置密码（通过旧密码）
func (s *UserService) SetPassword(ctx context.Context, userID uint64, req *model.UpdatePasswordRequest) error {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return ErrUserNotFound
	}

	if user.Password != "" {
		// 验证旧密码
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
			return ErrIncorrectPassword
		}
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), 12)
	if err != nil {
		return err
	}

	return s.repo.UpdatePassword(ctx, userID, string(hashed))
}

// SetPayPassword 设置支付密码（通过短信验证码）
func (s *UserService) SetPayPassword(ctx context.Context, userID uint64, req *model.SetPayPasswordRequest) error {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return ErrUserNotFound
	}

	storedCode, err := s.repo.GetSMSCode(ctx, user.Mobile)
	if err != nil || storedCode != req.Code {
		return ErrInvalidCode
	}
	defer s.repo.DelSMSCode(ctx, user.Mobile)

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.PayPassword), 12)
	if err != nil {
		return err
	}

	// 定向更新 pay_password 字段，不走全量 Update
	return s.repo.UpdatePayPassword(ctx, user.ID, string(hashed))
}

// ============ 用户信息 ============

// GetProfile 获取个人资料
func (s *UserService) GetProfile(ctx context.Context, userID uint64) (*model.UserProfile, error) {
	user, err := s.repo.GetCachedUser(ctx, userID)
	if err != nil {
		user, err = s.repo.FindByID(ctx, userID)
		if err != nil {
			return nil, ErrUserNotFound
		}
		s.repo.CacheUser(ctx, user)
	}
	profile := user.ToProfile()
	return profile, nil
}

// UpdateProfile 更新个人资料
func (s *UserService) UpdateProfile(ctx context.Context, userID uint64, req *model.UpdateProfileRequest) (*model.UserProfile, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}
	if req.AvatarURL != "" {
		user.AvatarURL = req.AvatarURL
	}
	if req.Gender != 0 {
		user.Gender = req.Gender
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	s.repo.InvalidateUserCache(ctx, userID)
	return user.ToProfile(), nil
}

// GetLoginLogs 获取登录日志
func (s *UserService) GetLoginLogs(ctx context.Context, userID uint64, limit int) ([]model.LoginLog, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.repo.FindLoginLogsByUserID(ctx, userID, limit)
}

// ============ 收货地址 ============

// GetAddresses 获取收货地址列表
func (s *UserService) GetAddresses(ctx context.Context, userID uint64) ([]model.ReceiverAddress, error) {
	return s.repo.FindAddressesByUserID(ctx, userID)
}

// CreateAddress 创建收货地址
func (s *UserService) CreateAddress(ctx context.Context, userID uint64, req *model.AddressRequest) (*model.ReceiverAddress, error) {
	count, _ := s.repo.CountAddress(ctx, userID)
	if count >= 20 {
		return nil, ErrAddressLimitReached
	}

	addr := &model.ReceiverAddress{
		UserID:       userID,
		Receiver:     req.Receiver,
		Mobile:       req.Mobile,
		ProvinceCode: req.ProvinceCode,
		CityCode:     req.CityCode,
		DistrictCode: req.DistrictCode,
		ProvinceName: req.ProvinceName,
		CityName:     req.CityName,
		DistrictName: req.DistrictName,
		Detail:       req.Detail,
		Tag:          req.Tag,
		IsDefault:    req.IsDefault,
	}

	if req.IsDefault == 1 {
		s.repo.ClearDefaultAddress(ctx, userID)
	}

	if err := s.repo.CreateAddress(ctx, addr); err != nil {
		return nil, err
	}
	return addr, nil
}

// UpdateAddress 更新收货地址
func (s *UserService) UpdateAddress(ctx context.Context, userID, addrID uint64, req *model.AddressRequest) (*model.ReceiverAddress, error) {
	addr, err := s.repo.FindAddressByID(ctx, addrID, userID)
	if err != nil {
		return nil, ErrAddressNotFound
	}

	addr.Receiver = req.Receiver
	addr.Mobile = req.Mobile
	addr.ProvinceCode = req.ProvinceCode
	addr.CityCode = req.CityCode
	addr.DistrictCode = req.DistrictCode
	addr.ProvinceName = req.ProvinceName
	addr.CityName = req.CityName
	addr.DistrictName = req.DistrictName
	addr.Detail = req.Detail
	addr.Tag = req.Tag

	if req.IsDefault == 1 {
		s.repo.ClearDefaultAddress(ctx, userID)
		addr.IsDefault = 1
	}

	if err := s.repo.UpdateAddress(ctx, addr); err != nil {
		return nil, err
	}
	return addr, nil
}

// DeleteAddress 删除收货地址
func (s *UserService) DeleteAddress(ctx context.Context, userID, addrID uint64) error {
	_, err := s.repo.FindAddressByID(ctx, addrID, userID)
	if err != nil {
		return ErrAddressNotFound
	}
	return s.repo.DeleteAddress(ctx, addrID, userID)
}

// SetDefaultAddress 设置默认收货地址
func (s *UserService) SetDefaultAddress(ctx context.Context, userID, addrID uint64) error {
	addr, err := s.repo.FindAddressByID(ctx, addrID, userID)
	if err != nil {
		return ErrAddressNotFound
	}
	s.repo.ClearDefaultAddress(ctx, userID)
	addr.IsDefault = 1
	return s.repo.UpdateAddress(ctx, addr)
}

// ============ 会员 & 积分 ============

// GetMemberProfile 获取会员信息
func (s *UserService) GetMemberProfile(ctx context.Context, userID uint64) (*model.MemberResponse, error) {
	member, err := s.repo.FindMemberByUserID(ctx, userID)
	if err != nil {
		// 创建默认会员记录
		member = &model.Member{UserID: userID, Level: 0}
		s.repo.CreateMember(ctx, member)
	}

	levels, _ := s.repo.FindAllMemberLevels(ctx)
	levelMap := make(map[uint8]string)
	nextThreshold := float64(0)
	for _, lv := range levels {
		levelMap[lv.Level] = lv.Name
		if lv.Level == member.Level+1 {
			nextThreshold = lv.Threshold
		}
	}

	levelName := levelMap[member.Level]
	if levelName == "" {
		levelName = "注册会员"
	}

	resp := &model.MemberResponse{
		Member: &model.MemberProfile{
			UserID:           userID,
			Level:            member.Level,
			LevelName:        levelName,
			TotalSpend:       member.TotalSpend,
			GrowthValue:      member.GrowthValue,
			NextThreshold:    nextThreshold,
			NextGrowthNeeded: int64(nextThreshold - member.TotalSpend),
		},
	}

	return resp, nil
}

// GetPoints 获取积分信息
func (s *UserService) GetPoints(ctx context.Context, userID uint64) (*model.PointsAccount, error) {
	account, err := s.repo.FindPointsAccountByUserID(ctx, userID)
	if err != nil {
		// 创建默认积分账户
		account = &model.PointsAccount{UserID: userID}
		s.repo.CreatePointsAccount(ctx, account)
	}
	return account, nil
}

// AddPoints 增加积分（供外部服务调用，通过 RPC 或消息队列触发）
func (s *UserService) AddPoints(ctx context.Context, userID uint64, points int64, bizType, orderNo string) error {
	account, err := s.repo.FindPointsAccountByUserID(ctx, userID)
	if err != nil {
		account = &model.PointsAccount{UserID: userID}
		s.repo.CreatePointsAccount(ctx, account)
	}

	before := account.Balance
	account.Balance += points
	account.TotalEarned += points

	if err := s.repo.UpdatePointsAccount(ctx, account); err != nil {
		return err
	}

	// 记录积分流水
	expireDate := time.Now().AddDate(2, 0, 0) // 2年后过期
	log := &model.PointsLog{
		UserID:        userID,
		OrderNo:       orderNo,
		Type:          bizType,
		Points:        points,
		BalanceBefore: before,
		BalanceAfter:  account.Balance,
		ExpireDate:    &model.Date{Time: expireDate},
	}
	return s.repo.CreatePointsLog(ctx, log)
}
