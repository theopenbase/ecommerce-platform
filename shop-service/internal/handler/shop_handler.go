package handler

import (
	security "github.com/ecommerce/common/middleware/security"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/ecommerce/shop-service/internal/model"
	"github.com/ecommerce/shop-service/internal/service"
)

type ShopHandler struct {
	svc *service.ShopService
}

func NewShopHandler(svc *service.ShopService) *ShopHandler {
	return &ShopHandler{svc: svc}
}

func (h *ShopHandler) RegisterRoutes(router *gin.Engine) {
	// C端
	router.GET("/api/v1/shops/:id", h.GetShop)
	router.GET("/api/v1/shops/:id/performance", h.GetPerformance)
	router.GET("/api/v1/shops/:id/decoration", h.GetDecoration)
	router.GET("/api/v1/shops/:id/freight", h.GetFreightTemplates)

	// B端商家
	seller := router.Group("/api/v1/seller/shop")
	seller.Use(authMiddleware())
	{
		seller.POST("", h.RegisterShop)
		seller.PUT("", h.UpdateShop)
		seller.POST("/qualifications", h.SubmitQualification)
		seller.POST("/freight", h.CreateFreightTemplate)
		seller.DELETE("/freight/:id", h.DeleteFreightTemplate)
		seller.PUT("/decoration", h.SaveDecoration)
	}
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		shopIDStr := c.GetHeader("X-Shop-ID")
		if shopIDStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 7001, "message": "unauthorized"})
			return
		}
		shopID, _ := strconv.ParseUint(shopIDStr, 10, 64)
		c.Set("shop_id", shopID)
		c.Next()
	}
}

func getShopID(c *gin.Context) uint64 {
	uid, _ := strconv.ParseUint(c.GetHeader("X-Shop-ID"), 10, 64)
	return uid
}

func getUserID(c *gin.Context) uint64 {
	uid, _ := strconv.ParseUint(c.GetHeader("X-User-ID"), 10, 64)
	return uid
}

func (h *ShopHandler) RegisterShop(c *gin.Context) {
	userID := getUserID(c)
	var req model.RegisterShopRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, security.SafeErr(err))
		return
	}
	shop, err := h.svc.RegisterShop(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, security.SafeErr(err))
		return
	}
	c.JSON(http.StatusCreated, success(shop))
}

func (h *ShopHandler) UpdateShop(c *gin.Context) {
	shopID := getShopID(c)
	userID := getUserID(c)
	var req model.UpdateShopRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, security.SafeErr(err))
		return
	}
	shop, err := h.svc.UpdateShop(c.Request.Context(), userID, shopID, &req)
	if err != nil {
		if err == service.ErrShopNotFound {
			c.JSON(http.StatusNotFound, security.SafeErr(err))
			return
		}
		c.JSON(http.StatusInternalServerError, security.SafeErr(err))
		return
	}
	c.JSON(http.StatusOK, success(shop))
}

func (h *ShopHandler) GetShop(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	shop, err := h.svc.GetShop(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, security.SafeErr(err))
		return
	}
	c.JSON(http.StatusOK, success(shop))
}

func (h *ShopHandler) GetPerformance(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	perf, err := h.svc.GetShopPerformance(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, security.SafeErr(err))
		return
	}
	c.JSON(http.StatusOK, success(perf))
}

func (h *ShopHandler) SubmitQualification(c *gin.Context) {
	shopID := getShopID(c)
	userID := getUserID(c)
	var req struct {
		QualType string `json:"qual_type" binding:"required"`
		CertNo   string `json:"cert_no"`
		FrontURL string `json:"front_url" binding:"required"`
		BackURL  string `json:"back_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, security.SafeErr(err))
		return
	}
	q, err := h.svc.SubmitQualification(c.Request.Context(), userID, shopID, req.QualType, req.CertNo, req.FrontURL, req.BackURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, security.SafeErr(err))
		return
	}
	c.JSON(http.StatusCreated, success(q))
}

func (h *ShopHandler) CreateFreightTemplate(c *gin.Context) {
	shopID := getShopID(c)
	var req model.CreateFreightTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, security.SafeErr(err))
		return
	}
	t, err := h.svc.CreateFreightTemplate(c.Request.Context(), shopID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, security.SafeErr(err))
		return
	}
	c.JSON(http.StatusCreated, success(t))
}

func (h *ShopHandler) GetFreightTemplates(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	templates, err := h.svc.GetFreightTemplates(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, security.SafeErr(err))
		return
	}
	c.JSON(http.StatusOK, success(templates))
}

func (h *ShopHandler) DeleteFreightTemplate(c *gin.Context) {
	shopID := getShopID(c)
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.DeleteFreightTemplate(c.Request.Context(), shopID, id); err != nil {
		c.JSON(http.StatusInternalServerError, security.SafeErr(err))
		return
	}
	c.JSON(http.StatusOK, success(nil))
}

func (h *ShopHandler) SaveDecoration(c *gin.Context) {
	shopID := getShopID(c)
	var req model.UpdateDecorationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, security.SafeErr(err))
		return
	}
	if err := h.svc.SaveDecoration(c.Request.Context(), shopID, req.Layout); err != nil {
		c.JSON(http.StatusInternalServerError, security.SafeErr(err))
		return
	}
	c.JSON(http.StatusOK, success(nil))
}

func (h *ShopHandler) GetDecoration(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	d, err := h.svc.GetDecoration(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, security.SafeErr(err))
		return
	}
	c.JSON(http.StatusOK, success(d))
}

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func success(data interface{}) Response { return Response{Code: 0, Message: "success", Data: data} }
func failed(code int, msg, _ string) Response { return Response{Code: code, Message: msg} }
