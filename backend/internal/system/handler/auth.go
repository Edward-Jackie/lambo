package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/Edward-Jackie/lambo/backend/internal/system/model"
	"github.com/Edward-Jackie/lambo/backend/internal/system/service"
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
	group.GET("/system/options", h.Options)
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
		ctx.Request = ctx.Request.WithContext(model.WithActor(ctx.Request.Context(), account.ID))
		ctx.Next()
	}
}

// Authorize 仅匹配显式授予的权限码，不提供管理员通配绕过。
func (h *AuthHandler) Authorize(ctx *gin.Context, code string) bool {
	account := currentAccount(ctx)
	if account == nil {
		return false
	}
	for _, granted := range account.Codes {
		if granted == code {
			ctx.Request = ctx.Request.WithContext(model.WithRequiredCode(ctx.Request.Context(), code))
			return true
		}
	}
	return false
}

func (h *AuthHandler) Info(ctx *gin.Context) {
	account := currentAccount(ctx)
	roles := make([]string, 0)
	for _, role := range account.Roles {
		roles = append(roles, role.Code)
	}
	home := "/no-access"
	for _, menu := range account.Menus {
		if menu.Kind == model.MenuPage && menu.Path != nil {
			home = *menu.Path
			break
		}
	}
	resp.HandleSuccess(ctx, gin.H{
		"userId": strconv.FormatUint(account.ID, 10), "username": account.Username,
		"realName": account.RealName, "roles": roles, "avatar": "",
		"homePath": home, "desc": "Lambo 管理后台", "token": "",
	})
}

func (h *AuthHandler) Codes(ctx *gin.Context) {
	resp.HandleSuccess(ctx, currentAccount(ctx).Codes)
}

func (h *AuthHandler) Menus(ctx *gin.Context) {
	var build func(float64) []gin.H
	build = func(parent float64) []gin.H {
		result := make([]gin.H, 0)
		for _, m := range currentAccount(ctx).Menus {
			if m.ParentId != parent || m.Kind == model.MenuButton {
				continue
			}
			component := "BasicLayout"
			if m.Component != nil {
				component = *m.Component
			}
			item := gin.H{"name": m.RouteName, "path": m.Path, "component": component, "meta": gin.H{"title": m.Name, "icon": m.Icon, "keepAlive": true}}
			if m.Kind == model.MenuDirectory {
				children := build(float64(m.ID))
				if len(children) == 0 {
					continue
				}
				item["children"] = children
			}
			result = append(result, item)
		}
		return result
	}
	resp.HandleSuccess(ctx, build(0))
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
