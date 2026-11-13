package server

import (
	"github.com/gin-gonic/gin"
	"github.com/Edward-Jackie/lambo/backend/internal/handler"
	"github.com/Edward-Jackie/lambo/backend/internal/middleware"
	"github.com/Edward-Jackie/lambo/backend/pkg/helper/resp"
	"github.com/Edward-Jackie/lambo/backend/pkg/log"
)

func NewServerHTTP(
	logger *log.Logger,
	userHandler *handler.UserHandler,
) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(
		middleware.CORSMiddleware(),
	)
	r.GET("/", func(ctx *gin.Context) {
		resp.HandleSuccess(ctx, map[string]interface{}{
			"say": "Hi Lambo!",
		})
	})
	r.GET("/user", userHandler.GetUserById)

	return r
}
