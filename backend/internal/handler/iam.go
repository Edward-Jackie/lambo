package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/Edward-Jackie/lambo/backend/internal/model"
	"github.com/Edward-Jackie/lambo/backend/pkg/helper/resp"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (h *Handler) handleIAMError(ctx *gin.Context, err error) bool {
	if errors.Is(err, model.ErrIAMInvalid) || errors.Is(err, model.ErrIAMConflict) || errors.Is(err, model.ErrIAMForbidden) {
		h.logger.WithContext(ctx).Warn("权限管理操作被拒绝", zap.Uint64("actorID", model.ActorID(ctx.Request.Context())), zap.String("method", ctx.Request.Method), zap.String("path", ctx.FullPath()), zap.Error(err))
	}
	switch {
	case errors.Is(err, model.ErrIAMInvalid):
		resp.HandleError(ctx, 400, 1, err.Error(), nil)
	case errors.Is(err, model.ErrIAMConflict):
		resp.HandleError(ctx, 409, 1, err.Error(), nil)
	case errors.Is(err, model.ErrIAMForbidden):
		resp.HandleError(ctx, 403, 1, err.Error(), nil)
	default:
		return false
	}
	return true
}

func (h *AuthHandler) RequireCode(code string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !h.Authorize(ctx, code) {
			ctx.Abort()
			resp.HandleError(ctx, 403, 1, "无操作权限", nil)
			return
		}
		ctx.Next()
	}
}

func (h *AuthHandler) Options(ctx *gin.Context) {
	allowed := false
	for _, code := range []string{model.PermissionUserQuery, model.PermissionUserCreate, model.PermissionUserUpdate, model.PermissionRoleQuery, model.PermissionRoleCreate, model.PermissionRoleUpdate, model.PermissionMenuQuery, model.PermissionMenuCreate, model.PermissionMenuUpdate} {
		if h.Authorize(ctx, code) {
			allowed = true
		}
	}
	if !allowed {
		resp.HandleError(ctx, 403, 1, "无操作权限", nil)
		return
	}
	result, err := h.auth.ManagementOptions(ctx.Request.Context())
	if err != nil {
		h.handleAuthError(ctx, err)
		return
	}
	result["grantedCodes"] = currentAccount(ctx).Codes
	resp.HandleSuccess(ctx, result)
}

func (h *SystemUserHandler) ResetPassword(ctx *gin.Context) {
	id, err := h.systemUserID(ctx)
	if err != nil {
		h.handleSystemUserError(ctx, err)
		return
	}
	var payload struct {
		Password string `json:"password"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(ctx.Writer, ctx.Request.Body, 4096))
	decoder.DisallowUnknownFields()
	if ctx.ContentType() != "application/json" {
		h.handleSystemUserError(ctx, fmt.Errorf("%w: 请提供 JSON 密码", model.ErrIAMInvalid))
		return
	}
	if err = decoder.Decode(&payload); err == nil {
		var extra interface{}
		if decoder.Decode(&extra) != io.EOF {
			err = fmt.Errorf("请求体包含多余内容")
		}
	}
	if err != nil {
		h.handleSystemUserError(ctx, fmt.Errorf("%w: 密码请求无效", model.ErrIAMInvalid))
		return
	}
	if err := h.systemUserService.ResetSystemUserPassword(ctx.Request.Context(), id, payload.Password); err != nil {
		h.handleSystemUserError(ctx, err)
		return
	}
	h.logger.WithContext(ctx).Info("重置账号密码", zap.Uint64("accountID", id), zap.Uint64("actorID", model.ActorID(ctx.Request.Context())))
	resp.HandleSuccess(ctx, nil)
}
