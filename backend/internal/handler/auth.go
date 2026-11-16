package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/Edward-Jackie/lambo/backend/internal/model"
	"github.com/Edward-Jackie/lambo/backend/internal/service"
	"github.com/Edward-Jackie/lambo/backend/pkg/helper/resp"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const accountContextKey = "lambo.auth.account"

type AuthHandler struct {
	*Handler
	auth *service.AuthService
}

func NewAuthHandler(handler *Handler, auth *service.AuthService) *AuthHandler {
	return &AuthHandler{Handler: handler, auth: auth}
}

func (h *AuthHandler) RegisterRoutes(router *gin.Engine) {
	router.POST("/api/auth/login", h.Login)
	// 登出幂等，失效会话也可清理客户端状态，避免认证拦截器重复登出。
	router.POST("/api/auth/logout", h.Logout)
	group := router.Group("/api", h.RequireSession())
	group.GET("/user/info", h.Info)
	group.GET("/auth/codes", h.Codes)
	group.GET("/menu/all", h.Menus)
}

func (h *AuthHandler) Login(ctx *gin.Context) {
	var payload struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, 4096)
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		resp.HandleError(ctx, http.StatusBadRequest, 1, "请提供正确的用户名和密码", nil)
		return
	}
	token, err := h.auth.Login(ctx.Request.Context(), strings.TrimSpace(payload.Username), payload.Password)
	if err != nil {
		h.handleAuthError(ctx, err)
		return
	}
	h.logger.Info("账号登录成功", zap.String("username", payload.Username))
	resp.HandleSuccess(ctx, gin.H{"accessToken": token})
}

func (h *AuthHandler) Logout(ctx *gin.Context) {
	if err := h.auth.Logout(ctx.Request.Context(), bearerToken(ctx)); err != nil {
		h.handleAuthError(ctx, err)
		return
	}
	resp.HandleSuccess(ctx, nil)
}

func (h *AuthHandler) RequireSession() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		account, err := h.auth.Authenticate(ctx.Request.Context(), bearerToken(ctx))
		if err != nil {
			ctx.Abort()
			h.handleAuthError(ctx, err)
			return
		}
		ctx.Set(accountContextKey, account)
		ctx.Next()
	}
}

// Authorize 仅匹配显式授予的权限码，不提供管理员通配绕过。
func (h *AuthHandler) Authorize(ctx *gin.Context, code string) bool {
	account := currentAccount(ctx)
	if account == nil {
		return false
	}
	for _, granted := range accountCodes(account.Role) {
		if granted == code {
			return true
		}
	}
	return false
}

func (h *AuthHandler) Info(ctx *gin.Context) {
	account := currentAccount(ctx)
	resp.HandleSuccess(ctx, gin.H{
		"userId": strconv.FormatUint(account.ID, 10), "username": account.Username,
		"realName": account.RealName, "roles": []string{account.Role}, "avatar": "",
		"homePath": "/business/category", "desc": "Lambo 管理后台", "token": "",
	})
}

func (h *AuthHandler) Codes(ctx *gin.Context) {
	resp.HandleSuccess(ctx, accountCodes(currentAccount(ctx).Role))
}

func (h *AuthHandler) Menus(ctx *gin.Context) {
	menus := []gin.H{}
	if h.Authorize(ctx, CategoryPermissionQuery) {
		menus = append(menus, gin.H{
			"name": "BusinessCategory", "path": "/business/category", "component": "/business/category/index",
			"meta": gin.H{"title": "分类管理", "icon": "lucide:folder", "keepAlive": true},
		})
	}
	resp.HandleSuccess(ctx, menus)
}

func (h *AuthHandler) handleAuthError(ctx *gin.Context, err error) {
	if errors.Is(err, service.ErrUnauthenticated) {
		resp.HandleError(ctx, http.StatusUnauthorized, 1, "账号密码错误或登录已失效", nil)
		return
	}
	h.logger.WithContext(ctx).Error("认证接口失败", zap.Error(err))
	resp.HandleError(ctx, http.StatusInternalServerError, 1, "认证服务暂时不可用", nil)
}

func currentAccount(ctx *gin.Context) *model.Account {
	value, _ := ctx.Get(accountContextKey)
	account, _ := value.(*model.Account)
	return account
}

func bearerToken(ctx *gin.Context) string {
	parts := strings.Fields(ctx.GetHeader("Authorization"))
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return parts[1]
}

func accountCodes(role string) []string {
	switch role {
	case model.RoleAdmin:
		return []string{CategoryPermissionQuery, CategoryPermissionCreate, CategoryPermissionUpdate, CategoryPermissionDelete}
	case model.RoleReader:
		return []string{CategoryPermissionQuery}
	default:
		return []string{}
	}
}
