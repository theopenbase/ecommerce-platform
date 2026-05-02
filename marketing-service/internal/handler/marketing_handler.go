package handler

import (
	security "github.com/ecommerce/common/middleware/security"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/ecommerce/marketing-service/internal/model"
	"github.com/ecommerce/marketing-service/internal/service"
)

type MarketingHandler struct {
	svc *service.MarketingService
}

func NewMarketingHandler(svc *service.MarketingService) *MarketingHandler {
	return &MarketingHandler{svc: svc}
}

func (h *MarketingHandler) RegisterRoutes(router *gin.Engine) {
	// C端
	router.GET("/api/v1/coupons", h.ListCoupons)
	router.GET("/api/v1/coupons/mine", h.ListMyCoupons)
	router.POST("/api/v1/coupons/:id/receive", h.ReceiveCoupon)
	router.GET("/api/v1/promotions", h.ListPromotions)
	router.GET("/api/v1/promotions/:id", h.GetPromotionDetail)

	// B端商家/平台运营
	router.POST("/api/v1/seller/coupons", h.CreateCoupon)
	router.POST("/api/v1/seller/coupons/:id/publish", h.PublishCoupon)
	router.POST("/api/v1/seller/promotions", h.CreatePromotion)
	router.POST("/api/v1/seller/promotions/:id/skus", h.AddPromotionSku)
}

func (h *MarketingHandler) ListCoupons(c *gin.Context) {
	var shopID *uint64
	if sid := c.Query("shop_id"); sid != "" {
		id, _ := strconv.ParseUint(sid, 10, 64)
		shopID = &id
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	coupons, total, err := h.svc.ListCoupons(c.Request.Context(), shopID, page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, security.SafeErr(err))
		return
	}
	c.JSON(http.StatusOK, success(gin.H{"total": total, "items": coupons}))
}

func (h *MarketingHandler) ListMyCoupons(c *gin.Context) {
	userID := getUserID(c)
	status := byteFromQuery(c.Query("status"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	var st *uint8
	if status > 0 {
		st = &status
	}
	items, total, err := h.svc.ListUserCoupons(c.Request.Context(), userID, st, page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, security.SafeErr(err))
		return
	}
	c.JSON(http.StatusOK, success(gin.H{"total": total, "items": items}))
}

func (h *MarketingHandler) ReceiveCoupon(c *gin.Context) {
	userID := getUserID(c)
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	uc, err := h.svc.ReceiveCoupon(c.Request.Context(), userID, id)
	if err != nil {
		switch err {
		case service.ErrCouponNotFound:
			c.JSON(http.StatusNotFound, security.SafeErr(err))
		case service.ErrCouponDepleted:
			c.JSON(http.StatusBadRequest, security.SafeErr(err))
		case service.ErrCouponLimitReached:
			c.JSON(http.StatusBadRequest, security.SafeErr(err))
		default:
			c.JSON(http.StatusInternalServerError, security.SafeErr(err))
		}
		return
	}
	c.JSON(http.StatusCreated, success(uc))
}

func (h *MarketingHandler) ListPromotions(c *gin.Context) {
	c.JSON(http.StatusOK, success([]interface{}{}))
}

func (h *MarketingHandler) GetPromotionDetail(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	detail, err := h.svc.GetPromotionDetail(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, security.SafeErr(err))
		return
	}
	c.JSON(http.StatusOK, success(detail))
}

func (h *MarketingHandler) CreateCoupon(c *gin.Context) {
	var req model.CreateCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, security.SafeErr(err))
		return
	}
	coupon, err := h.svc.CreateCoupon(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, security.SafeErr(err))
		return
	}
	c.JSON(http.StatusCreated, success(coupon))
}

func (h *MarketingHandler) PublishCoupon(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.PublishCoupon(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, security.SafeErr(err))
		return
	}
	c.JSON(http.StatusOK, success(nil))
}

func (h *MarketingHandler) CreatePromotion(c *gin.Context) {
	var req model.CreatePromotionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, security.SafeErr(err))
		return
	}
	promo, err := h.svc.CreatePromotion(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, security.SafeErr(err))
		return
	}
	c.JSON(http.StatusCreated, success(promo))
}

func (h *MarketingHandler) AddPromotionSku(c *gin.Context) {
	promoID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req model.AddPromotionSkuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, security.SafeErr(err))
		return
	}
	ps, err := h.svc.AddPromotionSku(c.Request.Context(), promoID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, security.SafeErr(err))
		return
	}
	c.JSON(http.StatusCreated, success(ps))
}

func getUserID(c *gin.Context) uint64 {
	var uid uint64
	fmt.Sscanf(c.GetHeader("X-User-ID"), "%d", &uid)
	return uid
}

func byteFromQuery(s string) uint8 {
	var b uint8
	if s == "" {
		return 0
	}
	fmt.Sscanf(s, "%d", &b)
	return b
}

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func success(data interface{}) Response { return Response{Code: 0, Message: "success", Data: data} }
func failed(code int, msg, _ string) Response { return Response{Code: code, Message: msg} }
