package handler

import (
	security "github.com/ecommerce/common/middleware/security"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"

	"github.com/ecommerce/order-service/internal/middleware"
	"github.com/ecommerce/order-service/internal/model"
	"github.com/ecommerce/order-service/internal/service"
)

type OrderHandler struct {
	svc   *service.OrderService
	cache *redis.Client
}

func NewOrderHandler(svc *service.OrderService, cache *redis.Client) *OrderHandler {
	return &OrderHandler{svc: svc, cache: cache}
}

func (h *OrderHandler) RegisterRoutes(router *gin.Engine) {
	// 购物车
	router.GET("/api/v1/cart", h.GetCart)
	router.POST("/api/v1/cart/items", h.AddToCart)
	router.PUT("/api/v1/cart/items/:id", h.UpdateCart)
	router.DELETE("/api/v1/cart/items/:id", h.RemoveCart)
	router.PUT("/api/v1/cart/select", h.SelectCart)

	// 订单
	router.POST("/api/v1/orders", h.CreateOrder)
	router.GET("/api/v1/orders", h.ListOrders)
	router.GET("/api/v1/orders/:order_no", h.GetOrderDetail)
	router.PUT("/api/v1/orders/:order_no/cancel", h.CancelOrder)
	router.PUT("/api/v1/orders/:order_no/receive", h.ConfirmReceive)
	router.POST("/api/v1/orders/:order_no/refund", middleware.PayPasswordMiddleware(h.cache), h.ApplyRefund)
}

// ============ 购物车 ============

func (h *OrderHandler) GetCart(c *gin.Context) {
	userID := getUserID(c)
	resp, err := h.svc.GetCart(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, security.SafeErr(err))
  log.Printf("[ORDER HANDLER ERROR] 获取失败: %v", err)
		return
	}
	c.JSON(http.StatusOK, success(resp))
}

func (h *OrderHandler) AddToCart(c *gin.Context) {
	userID := getUserID(c)
	var req model.CartAddRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, security.SafeErr(err))
  log.Printf("[ORDER HANDLER ERROR] 参数错误: %v", err)
		return
	}
	cart, err := h.svc.AddToCart(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, security.SafeErr(err))
  log.Printf("[ORDER HANDLER ERROR] 添加失败: %v", err)
		return
	}
	c.JSON(http.StatusCreated, success(cart))
}

func (h *OrderHandler) UpdateCart(c *gin.Context) {
	userID := getUserID(c)
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req model.CartUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, security.SafeErr(err))
  log.Printf("[ORDER HANDLER ERROR] 参数错误: %v", err)
		return
	}
	cart, err := h.svc.UpdateCart(c.Request.Context(), userID, id, &req)
	if err != nil {
		if err == service.ErrCartNotFound {
			c.JSON(http.StatusNotFound, security.SafeErr(err))
   log.Printf("[ORDER HANDLER ERROR] 购物车项不存在: %v", err)
			return
		}
		c.JSON(http.StatusInternalServerError, security.SafeErr(err))
  log.Printf("[ORDER HANDLER ERROR] 更新失败: %v", err)
		return
	}
	c.JSON(http.StatusOK, success(cart))
}

func (h *OrderHandler) RemoveCart(c *gin.Context) {
	userID := getUserID(c)
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.RemoveCart(c.Request.Context(), userID, id); err != nil {
		if err == service.ErrCartNotFound {
			c.JSON(http.StatusNotFound, security.SafeErr(err))
   log.Printf("[ORDER HANDLER ERROR] 购物车项不存在: %v", err)
			return
		}
		c.JSON(http.StatusInternalServerError, security.SafeErr(err))
  log.Printf("[ORDER HANDLER ERROR] 删除失败: %v", err)
		return
	}
	c.JSON(http.StatusOK, success(nil))
}

func (h *OrderHandler) SelectCart(c *gin.Context) {
	userID := getUserID(c)
	var req struct {
		Checked bool `json:"checked"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, security.SafeErr(err))
  log.Printf("[ORDER HANDLER ERROR] 参数错误: %v", err)
		return
	}
	if err := h.svc.SelectCartItems(c.Request.Context(), userID, req.Checked); err != nil {
		c.JSON(http.StatusInternalServerError, security.SafeErr(err))
  log.Printf("[ORDER HANDLER ERROR] 操作失败: %v", err)
		return
	}
	c.JSON(http.StatusOK, success(nil))
}

// ============ 订单 ============

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	userID := getUserID(c)
	var req model.ConfirmOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, security.SafeErr(err))
  log.Printf("[ORDER HANDLER ERROR] 参数错误: %v", err)
		return
	}
	resp, err := h.svc.CreateOrder(c.Request.Context(), userID, &req)
	if err != nil {
		if err == service.ErrIdempotentKeyUsed {
			c.JSON(http.StatusConflict, security.SafeErr(err))
   log.Printf("[ORDER HANDLER ERROR] 订单已存在，请勿重复提交: %v", err)
			return
		}
		c.JSON(http.StatusInternalServerError, security.SafeErr(err))
  log.Printf("[ORDER HANDLER ERROR] 创建失败: %v", err)
		return
	}
	c.JSON(http.StatusCreated, success(resp))
}

func (h *OrderHandler) ListOrders(c *gin.Context) {
	userID := getUserID(c)
	var q model.OrderListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, security.SafeErr(err))
  log.Printf("[ORDER HANDLER ERROR] 参数错误: %v", err)
		return
	}
	resp, err := h.svc.ListOrders(c.Request.Context(), userID, &q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, security.SafeErr(err))
  log.Printf("[ORDER HANDLER ERROR] 查询失败: %v", err)
		return
	}
	c.JSON(http.StatusOK, success(resp))
}

func (h *OrderHandler) GetOrderDetail(c *gin.Context) {
	userID := getUserID(c)
	orderNo := c.Param("order_no")
	resp, err := h.svc.GetOrderDetail(c.Request.Context(), userID, orderNo)
	if err != nil {
		if err == service.ErrOrderNotFound {
			c.JSON(http.StatusNotFound, security.SafeErr(err))
   log.Printf("[ORDER HANDLER ERROR] 订单不存在: %v", err)
			return
		}
		c.JSON(http.StatusInternalServerError, security.SafeErr(err))
  log.Printf("[ORDER HANDLER ERROR] 获取失败: %v", err)
		return
	}
	c.JSON(http.StatusOK, success(resp))
}

func (h *OrderHandler) CancelOrder(c *gin.Context) {
	userID := getUserID(c)
	orderNo := c.Param("order_no")
	var req model.CancelOrderRequest
	c.ShouldBindJSON(&req)
	if err := h.svc.CancelOrder(c.Request.Context(), userID, orderNo, req.Reason); err != nil {
		if err == service.ErrOrderNotFound {
			c.JSON(http.StatusNotFound, security.SafeErr(err))
   log.Printf("[ORDER HANDLER ERROR] 订单不存在: %v", err)
			return
		}
		if err == service.ErrInvalidStatus {
			c.JSON(http.StatusBadRequest, security.SafeErr(err))
   log.Printf("[ORDER HANDLER ERROR] 状态不允许取消: %v", err)
			return
		}
		c.JSON(http.StatusInternalServerError, security.SafeErr(err))
  log.Printf("[ORDER HANDLER ERROR] 取消失败: %v", err)
		return
	}
	c.JSON(http.StatusOK, success(nil))
}

func (h *OrderHandler) ConfirmReceive(c *gin.Context) {
	userID := getUserID(c)
	orderNo := c.Param("order_no")
	if err := h.svc.ConfirmReceive(c.Request.Context(), userID, orderNo); err != nil {
		if err == service.ErrOrderNotFound {
			c.JSON(http.StatusNotFound, security.SafeErr(err))
   log.Printf("[ORDER HANDLER ERROR] 订单不存在: %v", err)
			return
		}
		c.JSON(http.StatusInternalServerError, security.SafeErr(err))
  log.Printf("[ORDER HANDLER ERROR] 确认收货失败: %v", err)
		return
	}
	c.JSON(http.StatusOK, success(nil))
}

func (h *OrderHandler) ApplyRefund(c *gin.Context) {
	userID := getUserID(c)
	var req model.ApplyRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, security.SafeErr(err))
  log.Printf("[ORDER HANDLER ERROR] 参数错误: %v", err)
		return
	}
	if err := h.svc.ApplyRefund(c.Request.Context(), userID, &req); err != nil {
		if err == service.ErrOrderNotFound {
			c.JSON(http.StatusNotFound, security.SafeErr(err))
   log.Printf("[ORDER HANDLER ERROR] 订单不存在: %v", err)
			return
		}
		c.JSON(http.StatusInternalServerError, security.SafeErr(err))
  log.Printf("[ORDER HANDLER ERROR] 申请失败: %v", err)
		return
	}
	c.JSON(http.StatusOK, success(nil))
}

// ============ 辅助 ============

func getUserID(c *gin.Context) uint64 {
	uid, _ := strconv.ParseUint(c.GetHeader("X-User-ID"), 10, 64)
	return uid
}

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func success(data interface{}) Response { return Response{Code: 0, Message: "success", Data: data} }
func failed(code int, message, errMsg string) Response { return Response{Code: code, Message: message} }
