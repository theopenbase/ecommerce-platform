package handler

import (
	security "github.com/ecommerce/common/middleware/security"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/ecommerce/goods-service/internal/middleware"
	"github.com/ecommerce/goods-service/internal/model"
	"github.com/ecommerce/goods-service/internal/service"
)

type GoodsHandler struct {
	svc *service.GoodsService
}

func NewGoodsHandler(svc *service.GoodsService) *GoodsHandler {
	return &GoodsHandler{svc: svc}
}

func (h *GoodsHandler) RegisterRoutes(router *gin.Engine) {
	// C端 - 公开接口
	router.GET("/api/v1/categories", h.GetCategories)
	router.GET("/api/v1/categories/:id", h.GetCategory)
	router.GET("/api/v1/categories/:id/attributes", h.GetCategoryAttrs)
	router.GET("/api/v1/products", h.ListProducts)
	router.GET("/api/v1/products/spu/:id", h.GetSpuDetail)
	router.GET("/api/v1/goods/skus/:id/snapshot", h.GetSkuSnapshot)

	// 品牌
	router.GET("/api/v1/brands", h.ListBrands)

	// B端 - 商家接口（需认证）
	seller := router.Group("/api/v1/seller/products", middleware.ShopAuth())
	{
		seller.POST("", h.CreateSpu)
		seller.PUT("/:id", h.UpdateSpu)
		seller.POST("/sku", h.CreateSku)
		seller.PUT("/sku/:id", h.UpdateSku)
		seller.PUT("/sku/:id/status", h.UpdateSkuStatus)
	}
}

// ============ C端接口 ============

func (h *GoodsHandler) GetCategories(c *gin.Context) {
	tree, err := h.svc.BuildCategoryTree(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, security.SafeErr(err))
		return
	}
	c.JSON(http.StatusOK, success(tree))
}

func (h *GoodsHandler) GetCategory(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	cat, err := h.svc.GetCategoryByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, security.SafeErr(err))
		return
	}
	c.JSON(http.StatusOK, success(cat))
}

func (h *GoodsHandler) GetCategoryAttrs(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	templates, err := h.svc.GetCategoryAttrTemplates(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, security.SafeErr(err))
		return
	}
	c.JSON(http.StatusOK, success(templates))
}

func (h *GoodsHandler) ListProducts(c *gin.Context) {
	var q model.GoodsListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, security.SafeErr(err))
		return
	}

	resp, err := h.svc.ListGoods(c.Request.Context(), &q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, security.SafeErr(err))
		return
	}
	c.JSON(http.StatusOK, success(resp))
}

func (h *GoodsHandler) GetSpuDetail(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	detail, err := h.svc.GetSpuDetail(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, security.SafeErr(err))
		return
	}
	c.JSON(http.StatusOK, success(detail))
}

func (h *GoodsHandler) ListBrands(c *gin.Context) {
	// 暂省略 brand 列表接口
	c.JSON(http.StatusOK, success([]interface{}{}))
}

// ============ B端商家接口 ============

func (h *GoodsHandler) CreateSpu(c *gin.Context) {
	shopID, _ := middleware.GetShopID(c)
	var req model.CreateSpuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, security.SafeErr(err))
		return
	}

	spu, err := h.svc.CreateSpu(c.Request.Context(), shopID, &req)
	if err != nil {
		switch err {
		case service.ErrSpuCodeExists:
			c.JSON(http.StatusConflict, security.SafeErr(err))
		case service.ErrCategoryNotFound:
			c.JSON(http.StatusBadRequest, security.SafeErr(err))
		default:
			c.JSON(http.StatusInternalServerError, security.SafeErr(err))
		}
		return
	}
	c.JSON(http.StatusCreated, success(spu))
}

func (h *GoodsHandler) UpdateSpu(c *gin.Context) {
	shopID, _ := middleware.GetShopID(c)
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req model.CreateSpuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, security.SafeErr(err))
		return
	}

	spu, err := h.svc.UpdateSpu(c.Request.Context(), shopID, id, &req)
	if err != nil {
		if err == service.ErrSpuNotFound {
			c.JSON(http.StatusNotFound, security.SafeErr(err))
			return
		}
		c.JSON(http.StatusInternalServerError, security.SafeErr(err))
		return
	}
	c.JSON(http.StatusOK, success(spu))
}

func (h *GoodsHandler) CreateSku(c *gin.Context) {
	shopID, _ := middleware.GetShopID(c)
	var req struct {
		SpuID uint64                `json:"spu_id" binding:"required"`
		model.CreateSkuRequest
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, security.SafeErr(err))
		return
	}

	sku, err := h.svc.CreateSku(c.Request.Context(), shopID, req.SpuID, &req.CreateSkuRequest)
	if err != nil {
		switch err {
		case service.ErrSkuCodeExists:
			c.JSON(http.StatusConflict, security.SafeErr(err))
		case service.ErrSpuNotFound:
			c.JSON(http.StatusBadRequest, security.SafeErr(err))
		default:
			c.JSON(http.StatusInternalServerError, security.SafeErr(err))
		}
		return
	}
	c.JSON(http.StatusCreated, success(sku))
}

func (h *GoodsHandler) UpdateSku(c *gin.Context) {
	shopID, _ := middleware.GetShopID(c)
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req model.CreateSkuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, security.SafeErr(err))
		return
	}

	sku, err := h.svc.UpdateSku(c.Request.Context(), shopID, id, &req)
	if err != nil {
		if err == service.ErrSkuNotFound {
			c.JSON(http.StatusNotFound, security.SafeErr(err))
			return
		}
		c.JSON(http.StatusInternalServerError, security.SafeErr(err))
		return
	}
	c.JSON(http.StatusOK, success(sku))
}

func (h *GoodsHandler) UpdateSkuStatus(c *gin.Context) {
	shopID, _ := middleware.GetShopID(c)
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req model.UpdateSkuStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, security.SafeErr(err))
		return
	}

	if err := h.svc.UpdateSkuStatus(c.Request.Context(), shopID, id, req.Status); err != nil {
		if err == service.ErrSkuNotFound {
			c.JSON(http.StatusNotFound, security.SafeErr(err))
			return
		}
		c.JSON(http.StatusInternalServerError, security.SafeErr(err))
		return
	}
	c.JSON(http.StatusOK, success(nil))
}

// GetSkuSnapshot 获取 SKU 快照（供 order-service 下单前填充购物车）
func (h *GoodsHandler) GetSkuSnapshot(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	snapshot, err := h.svc.GetSkuSnapshot(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, security.SafeErr(err))
		return
	}
	c.JSON(http.StatusOK, success(snapshot))
}

// ============ 辅助函数 ============

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func success(data interface{}) Response {
	return Response{Code: 0, Message: "success", Data: data}
}

func failed(code int, message, errMsg string) Response {
	return Response{Code: code, Message: message}
}
