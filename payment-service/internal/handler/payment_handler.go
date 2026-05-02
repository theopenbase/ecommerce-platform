package handler

import (
	security "github.com/ecommerce/common/middleware/security"
	"net/http"

	"fmt"
	"github.com/gin-gonic/gin"

	"github.com/ecommerce/payment-service/internal/model"
	"github.com/ecommerce/payment-service/internal/service"
)

type PaymentHandler struct {
	svc *service.PaymentService
}

func NewPaymentHandler(svc *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{svc: svc}
}

func (h *PaymentHandler) RegisterRoutes(router *gin.Engine) {
	router.POST("/api/v1/payments/pay", h.Pay)
	router.POST("/api/v1/payments/balance/pay", h.PayByBalance)
	router.GET("/api/v1/payments/:trans_no", h.GetStatus)
	router.POST("/api/v1/payments/:trans_no/refund", h.Refund)

	// 支付回调（各渠道回调地址，公开）
	router.POST("/api/v1/payments/callback/alipay", h.CallbackAlipay)
	router.POST("/api/v1/payments/callback/wechat", h.CallbackWechat)

	// 账户
	router.GET("/api/v1/accounts/:user_id", h.GetAccount)
	router.GET("/api/v1/accounts/:id/logs", h.GetAccountLogs)
}

func (h *PaymentHandler) Pay(c *gin.Context) {
	var req model.PayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, security.SafeErr(err))
		return
	}
	buyerID := getUserID(c)
	resp, err := h.svc.CreatePayment(c.Request.Context(), buyerID, &req)
	if err != nil {
		if err == service.ErrOrderPaid {
			c.JSON(http.StatusConflict, security.SafeErr(err))
			return
		}
		c.JSON(http.StatusInternalServerError, security.SafeErr(err))
		return
	}
	c.JSON(http.StatusOK, success(resp))
}

func (h *PaymentHandler) PayByBalance(c *gin.Context) {
	var req struct {
		TransNo string `json:"trans_no" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, security.SafeErr(err))
		return
	}
	buyerID := getUserID(c)
	if err := h.svc.PayByBalance(c.Request.Context(), buyerID, req.TransNo); err != nil {
		switch err {
		case service.ErrTransNotFound:
			c.JSON(http.StatusNotFound, security.SafeErr(err))
		case service.ErrInsufficientBal:
			c.JSON(http.StatusBadRequest, security.SafeErr(err))
		default:
			c.JSON(http.StatusInternalServerError, security.SafeErr(err))
		}
		return
	}
	c.JSON(http.StatusOK, success(nil))
}

func (h *PaymentHandler) GetStatus(c *gin.Context) {
	transNo := c.Param("trans_no")
	trans, err := h.svc.GetPaymentStatus(c.Request.Context(), transNo)
	if err != nil {
		c.JSON(http.StatusNotFound, security.SafeErr(err))
		return
	}
	c.JSON(http.StatusOK, success(trans))
}

func (h *PaymentHandler) Refund(c *gin.Context) {
	var req model.RefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, security.SafeErr(err))
		return
	}
	buyerID := getUserID(c)
	resp, err := h.svc.Refund(c.Request.Context(), buyerID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, security.SafeErr(err))
		return
	}
	c.JSON(http.StatusOK, success(resp))
}

func (h *PaymentHandler) CallbackAlipay(c *gin.Context) {
	// 支付宝回调验签（生产需用 AlipayPublicKey 验签）
	transNo := c.PostForm("out_trade_no")
	status := c.PostForm("trade_status")
	channelTransNo := c.PostForm("trade_no")
	if err := h.svc.CallbackAlipay(c.Request.Context(), transNo, channelTransNo, status); err != nil {
		c.JSON(http.StatusInternalServerError, security.SafeErr(err))
		return
	}
	c.JSON(http.StatusOK, success("success"))
}

func (h *PaymentHandler) CallbackWechat(c *gin.Context) {
	// 微信回调验签（生产需用 WechatAPIKey 验签）
	transNo := c.PostForm("out_trade_no")
	resultCode := c.PostForm("result_code")
	channelTransNo := c.PostForm("transaction_id")
	if err := h.svc.CallbackWechat(c.Request.Context(), transNo, channelTransNo, resultCode); err != nil {
		c.JSON(http.StatusInternalServerError, security.SafeErr(err))
		return
	}
	c.JSON(http.StatusOK, success("success"))
}

func (h *PaymentHandler) GetAccount(c *gin.Context) {
	// 简化：user_id 从 header 获取
	userID := getUserID(c)
	account, err := h.svc.GetAccount(c.Request.Context(), userID, model.AccountTypeBuyer)
	if err != nil {
		c.JSON(http.StatusNotFound, security.SafeErr(err))
		return
	}
	c.JSON(http.StatusOK, success(account))
}

func (h *PaymentHandler) GetAccountLogs(c *gin.Context) {
	accountID := uint64(0)
	fmt.Sscanf(c.Param("id"), "%d", &accountID)
	logs, err := h.svc.GetAccountLogs(c.Request.Context(), accountID, 20)
	if err != nil {
		c.JSON(http.StatusInternalServerError, security.SafeErr(err))
		return
	}
	c.JSON(http.StatusOK, success(logs))
}

func getUserID(c *gin.Context) uint64 {
	var uid uint64
	fmt.Sscanf(c.GetHeader("X-User-ID"), "%d", &uid)
	return uid
}

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func success(data interface{}) Response { return Response{Code: 0, Message: "success", Data: data} }
func failed(code int, msg, _ string) Response { return Response{Code: code, Message: msg} }
