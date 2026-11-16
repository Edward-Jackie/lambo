// 由 Lambo 生成，可按业务修改；再次生成遇到差异会停止。
package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Edward-Jackie/lambo/backend/internal/model"
	"github.com/Edward-Jackie/lambo/backend/internal/service"
	"github.com/Edward-Jackie/lambo/backend/pkg/helper/resp"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strconv"
)

const (
	CategoryAPIPath          = "/api/categories"
	CategoryPermissionQuery  = "catalog:category:query"
	CategoryPermissionCreate = "catalog:category:create"
	CategoryPermissionUpdate = "catalog:category:update"
	CategoryPermissionDelete = "catalog:category:delete"
	categoryMaxBodyBytes     = 1 << 20
)

type CategoryHandler struct {
	*Handler
	categoryService service.CategoryService
}

func NewCategoryHandler(handler *Handler, svc service.CategoryService) *CategoryHandler {
	return &CategoryHandler{Handler: handler, categoryService: svc}
}

// RegisterRoutes 显式接入应用权限判断；未配置时拒绝注册，避免产生公开写接口。
func (h *CategoryHandler) RegisterRoutes(router gin.IRouter, authorize func(*gin.Context, string) bool) error {
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
	router.GET(CategoryAPIPath, require(CategoryPermissionQuery), h.ListCategory)
	router.GET(CategoryAPIPath+"/:id", require(CategoryPermissionQuery), h.GetCategory)
	router.POST(CategoryAPIPath, require(CategoryPermissionCreate), h.CreateCategory)
	router.PUT(CategoryAPIPath+"/:id", require(CategoryPermissionUpdate), h.UpdateCategory)
	router.DELETE(CategoryAPIPath+"/:id", require(CategoryPermissionDelete), h.DeleteCategory)
	return nil
}

func (h *CategoryHandler) ListCategory(ctx *gin.Context) {
	query := model.CategoryQuery{Page: 1, PageSize: model.CategoryDefaultPageSize}
	if err := ctx.ShouldBindQuery(&query); err != nil {
		h.handleCategoryError(ctx, fmt.Errorf("%w: 查询参数格式错误", model.ErrCategoryInvalid))
		return
	}
	result, err := h.categoryService.ListCategory(ctx.Request.Context(), query)
	if err != nil {
		h.handleCategoryError(ctx, err)
		return
	}
	resp.HandleSuccess(ctx, result)
}

func (h *CategoryHandler) GetCategory(ctx *gin.Context) {
	id, err := h.categoryID(ctx)
	if err != nil {
		h.handleCategoryError(ctx, err)
		return
	}
	item, err := h.categoryService.GetCategory(ctx.Request.Context(), id)
	if err != nil {
		h.handleCategoryError(ctx, err)
		return
	}
	resp.HandleSuccess(ctx, item)
}

func (h *CategoryHandler) CreateCategory(ctx *gin.Context) {
	payload, err := h.decodeCategoryPayload(ctx)
	if err != nil {
		h.handleCategoryError(ctx, err)
		return
	}
	item, err := h.categoryService.CreateCategory(ctx.Request.Context(), payload)
	if err != nil {
		h.handleCategoryError(ctx, err)
		return
	}
	h.logger.WithContext(ctx).Info("新增业务记录", zap.String("resource", "category"), zap.Any("id", item.ID))
	resp.HandleSuccess(ctx, item)
}

func (h *CategoryHandler) UpdateCategory(ctx *gin.Context) {
	id, err := h.categoryID(ctx)
	if err != nil {
		h.handleCategoryError(ctx, err)
		return
	}
	payload, err := h.decodeCategoryPayload(ctx)
	if err != nil {
		h.handleCategoryError(ctx, err)
		return
	}
	item, err := h.categoryService.UpdateCategory(ctx.Request.Context(), id, payload)
	if err != nil {
		h.handleCategoryError(ctx, err)
		return
	}
	h.logger.WithContext(ctx).Info("编辑业务记录", zap.String("resource", "category"), zap.Any("id", id))
	resp.HandleSuccess(ctx, item)
}

func (h *CategoryHandler) DeleteCategory(ctx *gin.Context) {
	id, err := h.categoryID(ctx)
	if err != nil {
		h.handleCategoryError(ctx, err)
		return
	}
	if err := h.categoryService.DeleteCategory(ctx.Request.Context(), id); err != nil {
		h.handleCategoryError(ctx, err)
		return
	}
	h.logger.WithContext(ctx).Info("删除业务记录", zap.String("resource", "category"), zap.Any("id", id))
	resp.HandleSuccess(ctx, nil)
}

func (h *CategoryHandler) categoryID(ctx *gin.Context) (uint64, error) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: ID 无效", model.ErrCategoryInvalid)
	}
	return id, model.ValidateCategoryID(id)
}

func (h *CategoryHandler) decodeCategoryPayload(ctx *gin.Context) (model.CategoryPayload, error) {
	var fields map[string]json.RawMessage
	invalid := fmt.Errorf("%w: 请求体必须是合法的 JSON 对象，且只能包含可编辑字段", model.ErrCategoryInvalid)
	if ctx.ContentType() != "application/json" {
		return model.CategoryPayload{}, invalid
	}
	decoder := json.NewDecoder(http.MaxBytesReader(ctx.Writer, ctx.Request.Body, categoryMaxBodyBytes))
	if err := decoder.Decode(&fields); err != nil {
		var limit *http.MaxBytesError
		if errors.As(err, &limit) {
			return model.CategoryPayload{}, err
		}
		return model.CategoryPayload{}, invalid
	}
	if fields == nil {
		return model.CategoryPayload{}, invalid
	}
	var extra interface{}
	if err := decoder.Decode(&extra); err != io.EOF {
		var limit *http.MaxBytesError
		if errors.As(err, &limit) {
			return model.CategoryPayload{}, err
		}
		return model.CategoryPayload{}, invalid
	}
	var payload model.CategoryPayload
	for name, value := range fields {
		switch name {
		case "name":
			if err := json.Unmarshal(value, &payload.Name); err != nil {
				return model.CategoryPayload{}, invalid
			}
			if payload.Name == nil {
				return model.CategoryPayload{}, invalid
			}
		case "status":
			if err := json.Unmarshal(value, &payload.Status); err != nil {
				return model.CategoryPayload{}, invalid
			}
			if payload.Status == nil {
				return model.CategoryPayload{}, invalid
			}
		case "sort":
			if err := json.Unmarshal(value, &payload.Sort); err != nil {
				return model.CategoryPayload{}, invalid
			}
			if payload.Sort == nil {
				return model.CategoryPayload{}, invalid
			}
		case "remark":
			if err := json.Unmarshal(value, &payload.Remark); err != nil {
				return model.CategoryPayload{}, invalid
			}

		default:
			return model.CategoryPayload{}, invalid
		}
	}
	return payload, nil
}

func (h *CategoryHandler) handleCategoryError(ctx *gin.Context, err error) {
	var limit *http.MaxBytesError
	switch {
	case errors.As(err, &limit):
		resp.HandleError(ctx, http.StatusRequestEntityTooLarge, 1, "请求体过大", nil)
	case errors.Is(err, model.ErrCategoryInvalid):
		resp.HandleError(ctx, http.StatusBadRequest, 1, err.Error(), nil)
	case errors.Is(err, model.ErrCategoryNotFound):
		resp.HandleError(ctx, http.StatusNotFound, 1, "记录不存在", nil)
	default:
		h.logger.WithContext(ctx).Error("业务接口失败", zap.String("resource", "category"), zap.String("method", ctx.Request.Method), zap.Error(err))
		resp.HandleError(ctx, http.StatusInternalServerError, 1, "操作失败，请稍后重试", nil)
	}
}
