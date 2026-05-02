package handler

import (
	security "github.com/ecommerce/common/middleware/security"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/ecommerce/user-service/internal/middleware"
	"github.com/ecommerce/user-service/internal/model"
	"github.com/ecommerce/user-service/internal/service"
)

type UserHandler struct {
	svc *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

func (h *UserHandler) RegisterRoutes(router *gin.Engine, authMW *middleware.AuthMiddleware) {
	// 公开接口
	router.POST("/api/v1/users/register", h.Register)
	router.POST("/api/v1/users/login", h.Login)
	router.POST("/api/v1/users/send_code", h.SendCode)
	router.POST("/api/v1/users/refresh_token", h.RefreshToken)

	// 需要认证的接口
	auth := authMW.RequireAuth()
	authGroup := router.Group("/api/v1/users", auth)
	{
		authGroup.GET("/me", h.GetProfile)
		authGroup.PUT("/me", h.UpdateProfile)
		authGroup.POST("/password", h.SetPassword)
		authGroup.POST("/pay_password", h.SetPayPassword)
		authGroup.GET("/login_logs", h.GetLoginLogs)

		// 收货地址
		authGroup.GET("/addresses", h.GetAddresses)
		authGroup.POST("/addresses", h.CreateAddress)
		authGroup.PUT("/addresses/:id", h.UpdateAddress)
		authGroup.DELETE("/addresses/:id", h.DeleteAddress)
		authGroup.PUT("/addresses/:id/default", h.SetDefaultAddress)

		// 会员
		authGroup.GET("/member/profile", h.GetMemberProfile)

		// 积分
		authGroup.GET("/points/balance", h.GetPoints)

		// 登出
		authGroup.POST("/logout", h.Logout)
	}
}

// ============ 认证接口 ============

func (h *UserHandler) SendCode(c *gin.Context) {
	var req model.SendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, security.SafeErr(err))
		return
	}

	if err := h.svc.SendCode(c.Request.Context(), req.Mobile); err != nil {
		c.JSON(http.StatusInternalServerError, security.SafeErr(err))
		return
	}
	c.JSON(http.StatusOK, success(nil))
}

func (h *UserHandler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, security.SafeErr(err))
		return
	}

	resp, err := h.svc.Register(c.Request.Context(), &req)
	if err != nil {
		switch err {
		case service.ErrUserExists:
			c.JSON(http.StatusConflict, security.SafeErr(err))
		case service.ErrInvalidCode:
			c.JSON(http.StatusUnauthorized, security.SafeErr(err))
		default:
			c.JSON(http.StatusInternalServerError, security.SafeErr(err))
		}
		return
	}
	c.JSON(http.StatusCreated, successWithData(resp))
}

func (h *UserHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, security.SafeErr(err))
		return
	}

	ip := c.ClientIP()
	resp, err := h.svc.Login(c.Request.Context(), &req, ip)
	if err != nil {
		switch err {
		case service.ErrUserNotFound:
			c.JSON(http.StatusNotFound, security.SafeErr(err))
		case service.ErrInvalidCode:
			c.JSON(http.StatusUnauthorized, security.SafeErr(err))
		default:
			c.JSON(http.StatusUnauthorized, security.SafeErr(err))
		}
		return
	}
	c.JSON(http.StatusOK, successWithData(resp))
}

func (h *UserHandler) RefreshToken(c *gin.Context) {
	var req model.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, security.SafeErr(err))
		return
	}

	resp, err := h.svc.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, security.SafeErr(err))
		return
	}
	c.JSON(http.StatusOK, successWithData(resp))
}

func (h *UserHandler) Logout(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	var req model.RefreshTokenRequest
	c.ShouldBindJSON(&req)
	h.svc.Logout(c.Request.Context(), userID, req.RefreshToken)
	c.JSON(http.StatusOK, success(nil))
}

// ============ 用户信息接口 ============

func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	profile, err := h.svc.GetProfile(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, security.SafeErr(err))
		return
	}
	c.JSON(http.StatusOK, successWithData(profile))
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	var req model.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, security.SafeErr(err))
		return
	}

	profile, err := h.svc.UpdateProfile(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, security.SafeErr(err))
		return
	}
	c.JSON(http.StatusOK, successWithData(profile))
}

func (h *UserHandler) SetPassword(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	var req model.UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, security.SafeErr(err))
		return
	}

	if err := h.svc.SetPassword(c.Request.Context(), userID, &req); err != nil {
		switch err {
		case service.ErrIncorrectPassword:
			c.JSON(http.StatusUnauthorized, security.SafeErr(err))
		default:
			c.JSON(http.StatusInternalServerError, security.SafeErr(err))
		}
		return
	}
	c.JSON(http.StatusOK, success(nil))
}

func (h *UserHandler) SetPayPassword(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	var req model.SetPayPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, security.SafeErr(err))
		return
	}

	if err := h.svc.SetPayPassword(c.Request.Context(), userID, &req); err != nil {
		c.JSON(http.StatusInternalServerError, security.SafeErr(err))
		return
	}
	c.JSON(http.StatusOK, success(nil))
}

func (h *UserHandler) GetLoginLogs(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	logs, err := h.svc.GetLoginLogs(c.Request.Context(), userID, 20)
	if err != nil {
		c.JSON(http.StatusInternalServerError, security.SafeErr(err))
		return
	}
	c.JSON(http.StatusOK, successWithData(logs))
}

// ============ 收货地址 ============

func (h *UserHandler) GetAddresses(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	addrs, err := h.svc.GetAddresses(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, security.SafeErr(err))
		return
	}
	c.JSON(http.StatusOK, successWithData(addrs))
}

func (h *UserHandler) CreateAddress(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	var req model.AddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, security.SafeErr(err))
		return
	}

	addr, err := h.svc.CreateAddress(c.Request.Context(), userID, &req)
	if err != nil {
		if err == service.ErrAddressLimitReached {
			c.JSON(http.StatusBadRequest, security.SafeErr(err))
			return
		}
		c.JSON(http.StatusInternalServerError, security.SafeErr(err))
		return
	}
	c.JSON(http.StatusCreated, successWithData(addr))
}

func (h *UserHandler) UpdateAddress(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	var req model.AddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, security.SafeErr(err))
		return
	}

	addrID := parseUint64ID(c)
	addr, err := h.svc.UpdateAddress(c.Request.Context(), userID, addrID, &req)
	if err != nil {
		if err == service.ErrAddressNotFound {
			c.JSON(http.StatusNotFound, security.SafeErr(err))
			return
		}
		c.JSON(http.StatusInternalServerError, security.SafeErr(err))
		return
	}
	c.JSON(http.StatusOK, successWithData(addr))
}

func (h *UserHandler) DeleteAddress(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	addrID := parseUint64ID(c)
	if err := h.svc.DeleteAddress(c.Request.Context(), userID, addrID); err != nil {
		if err == service.ErrAddressNotFound {
			c.JSON(http.StatusNotFound, security.SafeErr(err))
			return
		}
		c.JSON(http.StatusInternalServerError, security.SafeErr(err))
		return
	}
	c.JSON(http.StatusOK, success(nil))
}

func (h *UserHandler) SetDefaultAddress(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	addrID := parseUint64ID(c)
	if err := h.svc.SetDefaultAddress(c.Request.Context(), userID, addrID); err != nil {
		c.JSON(http.StatusInternalServerError, security.SafeErr(err))
		return
	}
	c.JSON(http.StatusOK, success(nil))
}

// ============ 会员 & 积分 ============

func (h *UserHandler) GetMemberProfile(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	resp, err := h.svc.GetMemberProfile(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, security.SafeErr(err))
		return
	}
	c.JSON(http.StatusOK, successWithData(resp))
}

func (h *UserHandler) GetPoints(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	account, err := h.svc.GetPoints(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, security.SafeErr(err))
		return
	}
	c.JSON(http.StatusOK, successWithData(account))
}

// ============ 辅助函数 ============

func parseUint64ID(c *gin.Context) uint64 {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 64)
	return id
}

// ============ 统一响应结构 ============

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func success(data interface{}) Response {
	return Response{Code: 0, Message: "success", Data: data}
}

func successWithData(data interface{}) Response {
	return Response{Code: 0, Message: "success", Data: data}
}

func failed(code int, message string, errMsg string) Response {
	return Response{Code: code, Message: message}
}
