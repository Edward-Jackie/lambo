package main

import (
	"fmt"

	"github.com/Edward-Jackie/lambo/backend/cmd/server/wire"
	"github.com/Edward-Jackie/lambo/backend/pkg/config"
	"github.com/Edward-Jackie/lambo/backend/pkg/http"
	"github.com/Edward-Jackie/lambo/backend/pkg/log"
	"go.uber.org/zap"
)

func main() {
	conf := config.NewConfig()
	logger := log.NewLog(conf)

	app, cleanup, err := wire.NewWire(conf, logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()
	logger.Info("服务已就绪", zap.String("host", "http://127.0.0.1:"+conf.GetString("http.port")))

	http.Run(app, fmt.Sprintf(":%d", conf.GetInt("http.port")))
}
