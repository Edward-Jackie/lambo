package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/Edward-Jackie/lambo/backend/cmd/server/wire"
	sysrepository "github.com/Edward-Jackie/lambo/backend/internal/system/repository"
	"github.com/Edward-Jackie/lambo/backend/pkg/log"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func seedMenu(args []string) error {
	flags := flag.NewFlagSet("seed-menu", flag.ContinueOnError)
	confirmNew := flags.Bool("new", false, "确认旧库从未安装该模块")
	if err := flags.Parse(args); err != nil {
		return err
	}
	const maxSeedBytes = 1 << 20
	data, err := io.ReadAll(io.LimitReader(os.Stdin, maxSeedBytes+1))
	if err != nil {
		return fmt.Errorf("读取菜单种子：%w", err)
	}
	if len(data) > maxSeedBytes {
		return fmt.Errorf("菜单种子超过 1 MiB")
	}
	data = bytes.TrimSpace(bytes.TrimPrefix(data, []byte{0xef, 0xbb, 0xbf}))
	header, lines, _ := bytes.Cut(data, []byte("\n"))
	if string(bytes.TrimSpace(header)) != "-- Lambo 菜单种子 v3" {
		return fmt.Errorf("需要第三版 Lambo 菜单种子，请更新旧生成文件")
	}
	definition, _, _ := bytes.Cut(lines, []byte("\n"))
	const definitionPrefix = "-- Lambo 菜单定义 "
	if !bytes.HasPrefix(definition, []byte(definitionPrefix)) {
		return fmt.Errorf("菜单种子缺少模块定义")
	}
	var spec sysrepository.MenuSeedDefinition
	if err := json.Unmarshal(definition[len(definitionPrefix):], &spec); err != nil {
		return fmt.Errorf("解析菜单种子定义：%w", err)
	}
	conf := viper.New()
	conf.Set("data.postgres.dsn", os.Getenv("LAMBO_POSTGRES_DSN"))
	logger := &log.Logger{Logger: zap.NewNop()}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	status, err := wire.ApplyMenuSeed(ctx, conf, logger, spec, *confirmNew)
	if err == nil {
		fmt.Println(status)
	}
	return err
}
