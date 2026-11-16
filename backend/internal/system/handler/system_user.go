// 由 Lambo 生成，可按业务修改；再次生成遇到差异会停止。
package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Edward-Jackie/lambo/backend/internal/system/model"
	"github.com/Edward-Jackie/lambo/backend/internal/system/service"
	"github.com/Edward-Jackie/lambo/backend/pkg/helper/resp"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strconv"
)

const (
	SystemUserAPIPath          = "/api/system/users"
	SystemUserPermissionQuery  = "system:user:query"
	SystemUserPermissionCreate = "system:user:create"
	SystemUserPermissionUpdate = "system:user:update"
	SystemUserPermissionDelete = "system:user:delete"
	systemUserMaxBodyBytes     = 1 << 20
)

type SystemUserHandler struct {
	*Handler
	systemUserService service.SystemUserService
}

func NewSystemUserHandler(handler *Handler, svc service.SystemUserService) *SystemUserHandler {
	return &SystemUserHandler{Handler: handler, systemUserService: svc}
}

// RegisterRoutes 显式接入应用权限判断；未配置时拒绝注册，避免产生公开写接口。
func (h *SystemUserHandler) RegisterRoutes(router gin.IRouter, authorize func(*gin.Context, string) bool) error {
	if router == nil || authorize == nil {
		return errors.New("注册业务路由需要路由器和权限判断函数")
	}
	require := func(permission string) gin.HandlerFunc {
		return func(ctx *gin.Context) {
			if !authorize(ctx, permission) {
				ctx.Abort()
				resp.HandleError(ctx, http.StatusForbidden, 1, "无操作权限", nil)
				return
			}
			ctx.Next()
		}
	}
	router.GET(SystemUserAPIPath, require(SystemUserPermissionQuery), h.ListSystemUser)
	router.GET(SystemUserAPIPath+"/:id", require(SystemUserPermissionQuery), h.GetSystemUser)
	router.POST(SystemUserAPIPath, require(SystemUserPermissionCreate), h.CreateSystemUser)
	router.PUT(SystemUserAPIPath+"/:id", require(SystemUserPermissionUpdate), h.UpdateSystemUser)
	router.DELETE(SystemUserAPIPath+"/:id", require(SystemUserPermissionDelete), h.DeleteSystemUser)
	return nil
}

func (h *SystemUserHandler) ListSystemUser(ctx *gin.Context) {
	query := model.SystemUserQuery{Page: 1, PageSize: model.SystemUserDefaultPageSize}
	if err := ctx.ShouldBindQuery(&query); err != nil {
		h.handleSystemUserError(ctx, fmt.Errorf("%w: 查询参数格式错误", model.ErrSystemUserInvalid))
		return
	}
	result, err := h.systemUserService.ListSystemUser(ctx.Request.Context(), query)
	if err != nil {
		h.handleSystemUserError(ctx, err)
		return
	}
	resp.HandleSuccess(ctx, result)
}

func (h *SystemUserHandler) GetSystemUser(ctx *gin.Context) {
	id, err := h.systemUserID(ctx)
	if err != nil {
		h.handleSystemUserError(ctx, err)
		return
	}
	item, err := h.systemUserService.GetSystemUser(ctx.Request.Context(), id)
	if err != nil {
		h.handleSystemUserError(ctx, err)
		return
	}
	resp.HandleSuccess(ctx, item)
}

func (h *SystemUserHandler) CreateSystemUser(ctx *gin.Context) {
	payload, err := h.decodeSystemUserPayload(ctx)
	if err != nil {
		h.handleSystemUserError(ctx, err)
		return
	}
	item, err := h.systemUserService.CreateSystemUser(ctx.Request.Context(), payload)
	if err != nil {
		h.handleSystemUserError(ctx, err)
		return
	}
	h.logger.WithContext(ctx).Info("新增业务记录", zap.String("resource", "system-user"), zap.Uint64("actorID", model.ActorID(ctx.Request.Context())), zap.Any("id", item.ID))
	resp.HandleSuccess(ctx, item)
}

func (h *SystemUserHandler) UpdateSystemUser(ctx *gin.Context) {
	id, err := h.systemUserID(ctx)
	if err != nil {
		h.handleSystemUserError(ctx, err)
		return
	}
	payload, err := h.decodeSystemUserPayload(ctx)
	if err != nil {
		h.handleSystemUserError(ctx, err)
		return
	}
	item, err := h.systemUserService.UpdateSystemUser(ctx.Request.Context(), id, payload)
	if err != nil {
		h.handleSystemUserError(ctx, err)
		return
	}
	h.logger.WithContext(ctx).Info("编辑业务记录", zap.String("resource", "system-user"), zap.Uint64("actorID", model.ActorID(ctx.Request.Context())), zap.Any("id", id))
	resp.HandleSuccess(ctx, item)
}

func (h *SystemUserHandler) DeleteSystemUser(ctx *gin.Context) {
	id, err := h.systemUserID(ctx)
	if err != nil {
		h.handleSystemUserError(ctx, err)
		return
	}
	if err := h.systemUserService.DeleteSystemUser(ctx.Request.Context(), id); err != nil {
		h.handleSystemUserError(ctx, err)
		return
	}
	h.logger.WithContext(ctx).Info("删除业务记录", zap.String("resource", "system-user"), zap.Uint64("actorID", model.ActorID(ctx.Request.Context())), zap.Any("id", id))
	resp.HandleSuccess(ctx, nil)
}

func (h *SystemUserHandler) systemUserID(ctx *gin.Context) (uint64, error) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: ID 无效", model.ErrSystemUserInvalid)
	}
	return id, model.ValidateSystemUserID(id)
}

func (h *SystemUserHandler) decodeSystemUserPayload(ctx *gin.Context) (model.SystemUserPayload, error) {
	var fields map[string]json.RawMessage
	invalid := fmt.Errorf("%w: 请求体必须是合法的 JSON 对象，且只能包含可编辑字段", model.ErrSystemUserInvalid)
	if ctx.ContentType() != "application/json" {
		return model.SystemUserPayload{}, invalid
	}
	decoder := json.NewDecoder(http.MaxBytesReader(ctx.Writer, ctx.Request.Body, systemUserMaxBodyBytes))
	if err := decoder.Decode(&fields); err != nil {
		var limit *http.MaxBytesError
		if errors.As(err, &limit) {
			return model.SystemUserPayload{}, err
		}
		return model.SystemUserPayload{}, invalid
	}
	if fields == nil {
		return model.SystemUserPayload{}, invalid
	}
	var extra interface{}
	if err := decoder.Decode(&extra); err != io.EOF {
		var limit *http.MaxBytesError
		if errors.As(err, &limit) {
			return model.SystemUserPayload{}, err
		}
		return model.SystemUserPayload{}, invalid
	}
	var payload model.SystemUserPayload
	for name, value := range fields {
		switch name {
		case "username":
			if err := json.Unmarshal(value, &payload.Username); err != nil {
				return model.SystemUserPayload{}, invalid
			}
			if payload.Username == nil {
				return model.SystemUserPayload{}, invalid
			}
		case "realName":
			if err := json.Unmarshal(value, &payload.RealName); err != nil {
				return model.SystemUserPayload{}, invalid
			}
			if payload.RealName == nil {
				return model.SystemUserPayload{}, invalid
			}
		case "enabled":
			if err := json.Unmarshal(value, &payload.Enabled); err != nil {
				return model.SystemUserPayload{}, invalid
			}
			if payload.Enabled == nil {
				return model.SystemUserPayload{}, invalid
			}
		case "remark":
			if err := json.Unmarshal(value, &payload.Remark); err != nil {
				return model.SystemUserPayload{}, invalid
			}

		case "roleIds":
			if err := json.Unmarshal(value, &payload.RoleIDs); err != nil || payload.RoleIDs == nil {
				return model.SystemUserPayload{}, invalid
			}
		case "password":
			if err := json.Unmarshal(value, &payload.Password); err != nil || payload.Password == nil {
				return model.SystemUserPayload{}, invalid
			}
		default:
			return model.SystemUserPayload{}, invalid
		}
	}
	return payload, nil
}

func (h *SystemUserHandler) handleSystemUserError(ctx *gin.Context, err error) {
	var limit *http.MaxBytesError
	if h.handleIAMError(ctx, err) {
		return
	}
	switch {
	case errors.As(err, &limit):
		resp.HandleError(ctx, http.StatusRequestEntityTooLarge, 1, "请求体过大", nil)
	case errors.Is(err, model.ErrSystemUserInvalid):
		resp.HandleError(ctx, http.StatusBadRequest, 1, err.Error(), nil)
	case errors.Is(err, model.ErrSystemUserNotFound):
		resp.HandleError(ctx, http.StatusNotFound, 1, "记录不存在", nil)
	default:
		h.logger.WithContext(ctx).Error("业务接口失败", zap.String("resource", "system-user"), zap.String("method", ctx.Request.Method), zap.Error(err))
		resp.HandleError(ctx, http.StatusInternalServerError, 1, "操作失败，请稍后重试", nil)
	}
}
