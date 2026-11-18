package db

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Edward-Jackie/lambo/tools/lambo/internal/pkg/helper"
	"github.com/duke-git/lancet/v2/strutil"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var rootPath, dsnFlag string
var dryRun, confirmNew bool

var CmdDB = &cobra.Command{
	Use: "db", Short: "数据库运维命令",
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error { return cmd.Help() },
}

var CmdSeed = &cobra.Command{
	Use:   "seed <名称>",
	Short: "执行生成模块的菜单初始化 SQL",
	Long: "按名称读取第三版菜单种子，通过当前项目后端的 IAM 校验与事务执行，无需启动服务。\n" +
		"旧库没有完整的种子历史，确认目标模块从未安装后可使用 --new。",
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
	CmdSeed.Flags().BoolVar(&confirmNew, "new", false, "确认旧库从未安装该模块")
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
	script = bytes.TrimSpace(bytes.TrimPrefix(script, []byte{0xef, 0xbb, 0xbf}))
	header, _, _ := bytes.Cut(script, []byte("\n"))
	if string(bytes.TrimSpace(header)) != "-- Lambo 菜单种子 v3" {
		return fmt.Errorf("%s 不是第三版 Lambo 菜单种子，请用新版模板更新后再执行", file)
	}
	dsn, err := resolveDSN(root)
	if err != nil {
		return err
	}
	args := []string{"-C", filepath.Join(root, "backend"), "run", "./cmd/server", "seed-menu"}
	if confirmNew {
		args = append(args, "-new")
	}
	runner := exec.CommandContext(cmd.Context(), "go", args...)
	runner.Env = append(os.Environ(), "LAMBO_POSTGRES_DSN="+dsn)
	runner.Stdin = bytes.NewReader(script)
	runner.Stdout, runner.Stderr = out, cmd.ErrOrStderr()
	if err := runner.Run(); err != nil {
		return fmt.Errorf("应用菜单种子 %s：%w", filepath.Base(file), err)
	}
	return nil
}

// locateSeed 拒绝两个域同名的脚本，避免把权限写入错误模块。
func locateSeed(root, name string) (string, error) {
	// 与 create 的 SnakeName 保持一致：先规范成结构体名再转下划线。
	snake := strutil.SnakeCase(strutil.UpperFirst(strutil.CamelCase(name)))
	var match string
	for _, domain := range []string{"business", "system"} {
		path := filepath.Join(root, "backend", "internal", domain, "seed", snake+".sql")
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			if match != "" {
				return "", fmt.Errorf("业务域和系统域均有 %s.sql，无法确定种子", snake)
			}
			match = path
		}
	}
	if match != "" {
		return match, nil
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
		var conf map[string]interface{}
		if err := yaml.Unmarshal(data, &conf); err != nil {
			return "", fmt.Errorf("无法解析 %s 的 YAML 数据库配置", path)
		}
		value, err := lookupConfigKey(conf, "data", "postgres", "dsn")
		if err != nil {
			return "", fmt.Errorf("读取数据库配置 %s：%w", path, err)
		}
		rawDSN, _ := value.(string)
		if dsn := strings.TrimSpace(rawDSN); dsn != "" {
			return dsn, nil
		}
	}
	return "", fmt.Errorf("未提供数据库连接：使用 --dsn、设置 LAMBO_POSTGRES_DSN，或填写 %s 的 data.postgres.dsn", path)
}

// 后端 Viper 按大小写不敏感的路径取值。存在多个同名键时拒绝选择连接串。
func lookupConfigKey(root map[string]interface{}, keys ...string) (interface{}, error) {
	var value interface{} = root
	for _, key := range keys {
		entries, ok := value.(map[string]interface{})
		if !ok {
			return nil, nil
		}
		found := false
		value = nil
		for name, entry := range entries {
			if !strings.EqualFold(name, key) {
				continue
			}
			if found {
				return nil, fmt.Errorf("配置键 %s 有重复的大小写形式", key)
			}
			value, found = entry, true
		}
	}
	return value, nil
}
