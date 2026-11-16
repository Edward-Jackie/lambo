package httpserver

import (
	"github.com/Edward-Jackie/lambo/backend/pkg/helper/resp"
	"github.com/gin-gonic/gin"
)

// NewEngine 只搭建 HTTP 底座（中间件与基础路由），业务与系统路由由装配层挂载。
func NewEngine() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(
		CORSMiddleware(),
	)
	// 仅表示 HTTP 进程可响应，不代替数据库和业务就绪检查。
	r.GET("/healthz", func(ctx *gin.Context) {
		resp.HandleSuccess(ctx, map[string]interface{}{"status": "ok"})
	})
	r.GET("/", func(ctx *gin.Context) {
		resp.HandleSuccess(ctx, map[string]interface{}{
			"say": "Hi Lambo!",
		})
	})
	return r
}
