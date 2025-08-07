package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"bkp-drive/internal/models"
	"bkp-drive/internal/services"
)

type AuthHandler struct {
	userService *services.UserService
}

func NewAuthHandler(userService *services.UserService) *AuthHandler {
	return &AuthHandler{
		userService: userService,
	}
}

// Register 用户注册
// @Summary      用户注册
// @Description  创建新用户账号
// @Tags         认证
// @Accept       json
// @Produce      json
// @Param        request   body      models.UserRegisterRequest  true  "注册信息"
// @Success      201       {object}  models.UserRegisterResponse
// @Failure      400       {object}  models.ErrorResponse
// @Failure      409       {object}  models.ErrorResponse
// @Failure      500       {object}  models.ErrorResponse
// @Router       /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.UserRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "请求参数错误: " + err.Error(),
		})
		return
	}

	user, err := h.userService.RegisterUser(&req)
	if err != nil {
		if err.Error() == "用户名已存在" {
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Success: false,
				Error:   err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.UserRegisterResponse{
		Success: true,
		Message: "注册成功",
		User:    user,
	})
}

// Login 用户登录
// @Summary      用户登录
// @Description  用户身份验证并获取访问令牌
// @Tags         认证
// @Accept       json
// @Produce      json
// @Param        request   body      models.UserLoginRequest  true  "登录信息"
// @Success      200       {object}  models.UserLoginResponse
// @Failure      400       {object}  models.ErrorResponse
// @Failure      401       {object}  models.ErrorResponse
// @Failure      500       {object}  models.ErrorResponse
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "请求参数错误: " + err.Error(),
		})
		return
	}

	user, token, err := h.userService.LoginUser(&req)
	if err != nil {
		if err.Error() == "用户名或密码错误" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Success: false,
				Error:   err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.UserLoginResponse{
		Success: true,
		Message: "登录成功",
		Token:   token,
		User:    user,
	})
}

// GetProfile 获取用户信息
// @Summary      获取用户信息
// @Description  获取当前登录用户的基本信息
// @Tags         认证
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200       {object}  models.UserLoginResponse
// @Failure      401       {object}  models.ErrorResponse
// @Router       /auth/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Success: false,
			Error:   "用户未登录",
		})
		return
	}

	username, _ := c.Get("username")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"user": gin.H{
			"user_id":  userID,
			"username": username,
		},
	})
}

// Logout 用户登出
// @Summary      用户登出
// @Description  用户登出（客户端需清除token）
// @Tags         认证
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200       {object}  models.ErrorResponse
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "登出成功",
	})
}