package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/ecommerce/payment-service/internal/model"
	"github.com/ecommerce/payment-service/internal/repository"
)

var (
	ErrTransNotFound    = errors.New("transaction not found")
	ErrOrderPaid       = errors.New("order already paid")
	ErrInsufficientBal = errors.New("insufficient balance")
	ErrRefundFailed    = errors.New("refund failed")
	ErrLockFailed      = errors.New("acquire lock failed")
)

type PaymentService struct {
	repo *repository.PaymentRepository
}

func NewPaymentService(repo *repository.PaymentRepository) *PaymentService {
	return &PaymentService{repo: repo}
}

// genTransNo generates payment transaction number
func genTransNo() string {
	return fmt.Sprintf("%s%010d", time.Now().Format("20060102150405"), rand.Int63n(9999999999))
}

func genRefundNo() string {
	return fmt.Sprintf("REF%s%d", time.Now().Format("20060102"), rand.Intn(100000))
}

// ============ 担保交易：创建支付 ============

// CreatePayment 创建支付流水（担保交易：款项先进入平台中间账户）
func (s *PaymentService) CreatePayment(ctx context.Context, buyerID uint64, req *model.PayRequest) (*model.PayResponse, error) {
	// 检查订单是否已支付
	existing, _ := s.repo.FindTransByOrderNo(ctx, req.OrderNo)
	if existing != nil && existing.Status == model.PayStatusSuccess {
		return nil, ErrOrderPaid
	}

	transNo := genTransNo()
	expireTime := time.Now().Add(15 * time.Minute) // 15分钟支付超时

	trans := &model.PaymentTransaction{
		TransNo:    transNo,
		OrderNo:   req.OrderNo,
		BuyerID:   buyerID,
		Channel:   req.Channel,
		Amount:    req.Amount,
		Status:    model.PayStatusPending,
		ExpireTime: expireTime,
	}

	if err := s.repo.CreateTrans(ctx, trans); err != nil {
		return nil, err
	}

	// 构建支付参数（各渠道实际调用）
	var payURL, qrCode string
	switch req.Channel {
	case model.ChannelAlipay:
		payURL = fmt.Sprintf("https://openapi.alipay.com/gateway.do?out_trade_no=%s", transNo)
		qrCode = fmt.Sprintf("https://qr.alipay.com/%s", transNo)
	case model.ChannelWechat:
		payURL = fmt.Sprintf("https://api.mch.weixin.qq.com/pay/unifiedorder?out_trade_no=%s", transNo)
		qrCode = fmt.Sprintf("weixin://wxpay/bizpayurl?pr=%s", transNo)
	case model.ChannelBalance:
		// 余额支付直接扣减
		payURL = ""
	}

	return &model.PayResponse{
		TransNo:   transNo,
		Channel:   req.Channel,
		PayURL:    payURL,
		QRCode:    qrCode,
		ExpiresIn: 900,
	}, nil
}

// ============ 支付回调 ============

// CallbackAlipay 支付宝回调
func (s *PaymentService) CallbackAlipay(ctx context.Context, transNo, channelTransNo string, status string) error {
	trans, err := s.repo.FindTransByTransNo(ctx, transNo)
	if err != nil {
		return ErrTransNotFound
	}
	if status != "TRADE_SUCCESS" {
		trans.Status = model.PayStatusFailed
		return s.repo.UpdateTrans(ctx, trans)
	}
	trans.Status = model.PayStatusSuccess
	trans.ChannelTransNo = channelTransNo
	now := time.Now()
	trans.PayTime = &now
	if err := s.repo.UpdateTrans(ctx, trans); err != nil {
		return err
	}
	// TODO: 发送支付成功消息给 order-service（MQ）
	return nil
}

// CallbackWechat 微信支付回调
func (s *PaymentService) CallbackWechat(ctx context.Context, transNo, channelTransNo string, resultCode string) error {
	trans, err := s.repo.FindTransByTransNo(ctx, transNo)
	if err != nil {
		return ErrTransNotFound
	}
	if resultCode != "SUCCESS" {
		trans.Status = model.PayStatusFailed
		return s.repo.UpdateTrans(ctx, trans)
	}
	trans.Status = model.PayStatusSuccess
	trans.ChannelTransNo = channelTransNo
	now := time.Now()
	trans.PayTime = &now
	return s.repo.UpdateTrans(ctx, trans)
}

// ============ 余额支付 ============

// PayByBalance 余额直接扣减（担保交易：冻结中间账户）
func (s *PaymentService) PayByBalance(ctx context.Context, buyerID uint64, transNo string) error {
	trans, err := s.repo.FindTransByTransNo(ctx, transNo)
	if err != nil {
		return ErrTransNotFound
	}
	if trans.Status == model.PayStatusSuccess {
		return ErrOrderPaid
	}

	// 获取买家账户
	buyerAccount, err := s.repo.FindAccountByUserID(ctx, buyerID, model.AccountTypeBuyer)
	if err != nil {
		return ErrInsufficientBal
	}
	if buyerAccount.Balance < trans.Amount {
		return ErrInsufficientBal
	}

	// 获取平台中间账户
	platformAccount, err := s.repo.FindAccountByUserID(ctx, 0, model.AccountTypePlatform)
	if err != nil {
		// 创建平台账户
		platformAccount = &model.Account{UserID: 0, UserType: model.AccountTypePlatform, Balance: 0}
		s.repo.CreateAccount(ctx, platformAccount)
	}

	// 余额扣减 + 中间账户入账
	before := buyerAccount.Balance
	buyerAccount.Balance -= trans.Amount
	if err := s.repo.UpdateAccount(ctx, buyerAccount); err != nil {
		return err
	}
	platformBefore := platformAccount.Balance
	platformAccount.Balance += trans.Amount
	if err := s.repo.UpdateAccount(ctx, platformAccount); err != nil {
		return err
	}

	// 账务流水
	s.repo.CreateAccountLog(ctx, &model.AccountLog{
		AccountID: buyerAccount.ID, TransNo: transNo, OrderNo: trans.OrderNo,
		Type: model.AccountLogPay, Amount: -trans.Amount,
		BalanceBefore: before, BalanceAfter: buyerAccount.Balance,
	})
	s.repo.CreateAccountLog(ctx, &model.AccountLog{
		AccountID: platformAccount.ID, TransNo: transNo, OrderNo: trans.OrderNo,
		Type: model.AccountLogRecharge, Amount: trans.Amount,
		BalanceBefore: platformBefore, BalanceAfter: platformAccount.Balance,
	})

	now := time.Now()
	trans.Status = model.PayStatusSuccess
	trans.PayTime = &now
	return s.repo.UpdateTrans(ctx, trans)
}

// ============ 退款 ============

// Refund 退款（原路退回）
func (s *PaymentService) Refund(ctx context.Context, buyerID uint64, req *model.RefundRequest) (*model.RefundResponse, error) {
	locked, err := s.repo.AcquireLock(ctx, fmt.Sprintf("refund:%s", req.TransNo), 10*time.Second)
	if err != nil || !locked {
		return nil, ErrLockFailed
	}
	defer s.repo.ReleaseLock(ctx, fmt.Sprintf("refund:%s", req.TransNo))

	trans, err := s.repo.FindTransByTransNo(ctx, req.TransNo)
	if err != nil {
		return nil, ErrTransNotFound
	}
	if trans.Status != model.PayStatusSuccess {
		return nil, errors.New("transaction not paid")
	}

	refundNo := genRefundNo()
	refund := &model.RefundRecord{
		RefundNo: refundNo,
		TransNo: req.TransNo,
		OrderNo: trans.OrderNo,
		BuyerID: buyerID,
		Amount: req.Amount,
		Status: 0,
		Reason: req.Reason,
	}
	if err := s.repo.CreateRefund(ctx, refund); err != nil {
		return nil, err
	}

	// 退款到买家余额（余额支付原路退回）
	if trans.Channel == model.ChannelBalance || trans.Channel == model.ChannelAlipay || trans.Channel == model.ChannelWechat {
		buyerAccount, err := s.repo.FindAccountByUserID(ctx, buyerID, model.AccountTypeBuyer)
		if err != nil {
			buyerAccount = &model.Account{UserID: buyerID, UserType: model.AccountTypeBuyer, Balance: 0}
			s.repo.CreateAccount(ctx, buyerAccount)
		}
		before := buyerAccount.Balance
		buyerAccount.Balance += req.Amount
		if err := s.repo.UpdateAccount(ctx, buyerAccount); err != nil {
			return nil, ErrRefundFailed
		}
		s.repo.CreateAccountLog(ctx, &model.AccountLog{
			AccountID: buyerAccount.ID, TransNo: req.TransNo, OrderNo: trans.OrderNo,
			Type: model.AccountLogRefund, Amount: req.Amount,
			BalanceBefore: before, BalanceAfter: buyerAccount.Balance,
			Note: "退款",
		})
	}

	// 更新退款状态
	now := time.Now()
	refund.Status = 1
	refund.ProcessTime = &now
	s.repo.UpdateRefund(ctx, refund)

	// 更新支付流水状态
	trans.Status = model.PayStatusRefunded
	s.repo.UpdateTrans(ctx, trans)

	return &model.RefundResponse{
		RefundNo: refundNo,
		Amount:   req.Amount,
		Status:   1,
	}, nil
}

// GetPaymentStatus 查询支付状态
func (s *PaymentService) GetPaymentStatus(ctx context.Context, transNo string) (*model.PaymentTransaction, error) {
	return s.repo.FindTransByTransNo(ctx, transNo)
}

// GetAccount 获取账户信息
func (s *PaymentService) GetAccount(ctx context.Context, userID uint64, userType uint8) (*model.Account, error) {
	return s.repo.FindAccountByUserID(ctx, userID, userType)
}

// GetAccountLogs 获取账务流水
func (s *PaymentService) GetAccountLogs(ctx context.Context, accountID uint64, limit int) ([]model.AccountLog, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.repo.FindAccountLogs(ctx, accountID, limit)
}
