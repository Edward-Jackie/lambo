package db

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Edward-Jackie/lambo/tools/lambo/internal/pkg/helper"
	"github.com/duke-git/lancet/v2/strutil"
	"github.com/jackc/pgx/v5"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var rootPath, dsnFlag string
var dryRun bool

var CmdDB = &cobra.Command{
	Use: "db", Short: "数据库运维命令",
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error { return cmd.Help() },
}

var CmdSeed = &cobra.Command{
	Use:   "seed <名称>",
	Short: "执行生成模块的菜单初始化 SQL",
	Long: "按名称读取 backend/internal/{business,system}/seed/<名称>.sql 并在单个事务内执行。\n" +
		"SQL 为幂等脚本：重复执行不会产生重复菜单，也不会覆盖人工调整的授权。",
	Example: "lambo db seed category --dry-run",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSeed(cmd, args[0])
	},
}

func init() {
	CmdSeed.Flags().StringVar(&rootPath, "root", "", "Lambo 项目根目录；默认向上查找")
	CmdSeed.Flags().StringVar(&dsnFlag, "dsn", "", "PostgreSQL DSN；默认读 LAMBO_POSTGRES_DSN 或 backend/config/local.yml")
	CmdSeed.Flags().BoolVar(&dryRun, "dry-run", false, "只输出 SQL，不连接数据库")
	CmdDB.AddCommand(CmdSeed)
}

func runSeed(cmd *cobra.Command, name string) error {
	root, err := helper.FindLamboRoot(rootPath)
	if err != nil {
		return err
	}
	file, err := locateSeed(root, name)
	if err != nil {
		return err
	}
	script, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("读取 %s：%w", file, err)
	}
	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "脚本 %s\n", file)
	if dryRun {
		fmt.Fprintf(out, "%s\n", script)
		return nil
	}
	if bytes.HasPrefix(script, []byte("-- 由 Lambo 生成：")) {
		return fmt.Errorf("%s 使用旧版菜单种子，请用新版模板更新后再执行，避免恢复已撤销的授权", file)
	}
	dsn, err := resolveDSN(root)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
	defer cancel()
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return fmt.Errorf("连接数据库：%w", err)
	}
	defer func() { _ = conn.Close(context.Background()) }()
	// 整份脚本按简单协议一次执行，避免逐句拆分误伤含分号的字面量。
	if err := pgx.BeginFunc(ctx, conn, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, string(script), pgx.QueryExecModeSimpleProtocol)
		return err
	}); err != nil {
		return fmt.Errorf("应用菜单种子 %s：%w", filepath.Base(file), err)
	}
	fmt.Fprintln(out, "菜单种子执行完成。新增授权将在重新登录后显示。")
	return nil
}

// locateSeed 先找业务区再找系统区，避免两个域同名模块互相覆盖。
func locateSeed(root, name string) (string, error) {
	// 与 create 的 SnakeName 保持一致：先规范成结构体名再转下划线。
	snake := strutil.SnakeCase(strutil.UpperFirst(strutil.CamelCase(name)))
	for _, domain := range []string{"business", "system"} {
		path := filepath.Join(root, "backend", "internal", domain, "seed", snake+".sql")
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path, nil
		}
	}
	return "", fmt.Errorf("未找到菜单脚本 backend/internal/{business,system}/seed/%s.sql；先执行 lambo create frontend %s", snake, name)
}

// resolveDSN 依次取 --dsn、环境变量与开发配置，与后端一致。
func resolveDSN(root string) (string, error) {
	if dsnFlag != "" {
		return dsnFlag, nil
	}
	if dsn := os.Getenv("LAMBO_POSTGRES_DSN"); dsn != "" {
		return dsn, nil
	}
	path := filepath.Join(root, "backend", "config", "local.yml")
	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("读取数据库配置 %s：%w", path, err)
	}
	if err == nil {
		var conf struct {
			Data struct {
				Postgres struct {
					DSN string `yaml:"dsn"`
				} `yaml:"postgres"`
			} `yaml:"data"`
		}
		if err := yaml.Unmarshal(data, &conf); err != nil {
			return "", fmt.Errorf("无法解析 %s 的 YAML 数据库配置", path)
		}
		if dsn := strings.TrimSpace(conf.Data.Postgres.DSN); dsn != "" {
			return dsn, nil
		}
	}
	return "", fmt.Errorf("未提供数据库连接：使用 --dsn、设置 LAMBO_POSTGRES_DSN，或填写 %s 的 data.postgres.dsn", path)
}
