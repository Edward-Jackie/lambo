package new

import (
	"bytes"
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Edward-Jackie/lambo/config"
	"github.com/Edward-Jackie/lambo/internal/pkg/helper"
	"github.com/spf13/cobra"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
)

var repoURL, modulePath string
var CmdNew = &cobra.Command{
	Use: "new <项目名>", Short: "从 Lambo 仓库创建前后端项目（不安装依赖）",
	Example: "lambo new shop --module github.com/example/shop",
	Args:    cobra.ExactArgs(1), RunE: run,
}

func init() {
	CmdNew.Flags().StringVarP(&repoURL, "repo-url", "r", config.RepoBase, "模板 Git 地址或本地仓库")
	CmdNew.Flags().StringVar(&modulePath, "module", "", "新项目 Go 模块路径；默认使用项目名")
}

func run(cmd *cobra.Command, args []string) error {
	name := args[0]
	if name == "." || name == ".." || strings.HasPrefix(name, "-") || filepath.Base(name) != name || strings.ContainsAny(name, `\/:`) {
		return fmt.Errorf("项目名必须是当前目录下的新目录名")
	}
	mod := modulePath
	if mod == "" {
		mod = name
	}
	if err := module.CheckImportPath(mod); err != nil {
		return fmt.Errorf("无效模块路径：%w", err)
	}
	if _, err := os.Lstat(name); !os.IsNotExist(err) {
		return fmt.Errorf("目标 %s 已存在或不可访问，不会覆盖", name)
	}
	// 先完整准备临时项目，失败时不留下半成品，也不删除现有目录。
	stage, err := os.MkdirTemp(".", ".lambo-new-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stage)
	target := filepath.Join(stage, name)
	clone := exec.CommandContext(cmd.Context(), "git", "clone", "--depth", "1", "--", repoURL, target)
	if output, err := clone.CombinedOutput(); err != nil {
		return fmt.Errorf("克隆模板：%w\n%s", err, output)
	}
	if info, err := os.Stat(filepath.Join(target, "frontend", "apps", "web-ele", "package.json")); err != nil || info.IsDir() {
		return fmt.Errorf("模板缺少 Vben Element Plus 应用")
	}
	for _, dir := range []string{".", "backend"} {
		base := filepath.Join(target, dir)
		oldModule, err := helper.ReadModulePath(base)
		if err != nil {
			return fmt.Errorf("模板必须包含 CLI 和 backend 两个 Go 模块：%w", err)
		}
		newModule := mod
		if dir == "backend" {
			newModule += "/backend"
		}
		if err := replaceModule(base, oldModule, newModule); err != nil {
			return err
		}
	}
	if err := os.RemoveAll(filepath.Join(target, ".git")); err != nil {
		return err
	}
	if _, err := os.Lstat(name); !os.IsNotExist(err) {
		return fmt.Errorf("目标 %s 已存在，不会覆盖", name)
	}
	if err := os.Rename(target, name); err != nil {
		return fmt.Errorf("发布新目录：%w", err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "已创建 %s。请先阅读 README.md，安装依赖并配置数据库；本命令未启动服务。\n", name)
	return nil
}

func replaceModule(root, oldModule, newModule string) error {
	return filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if path != root && (entry.Name() == ".git" || entry.Name() == "backend" || entry.Name() == "frontend" || entry.Name() == "node_modules") {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return nil
		}
		if filepath.Ext(path) != ".go" && entry.Name() != "go.mod" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if !bytes.Contains(data, []byte(oldModule)) {
			return nil
		}
		updated := data
		if entry.Name() == "go.mod" {
			parsed, err := modfile.Parse(path, data, nil)
			if err != nil {
				return err
			}
			if err := parsed.AddModuleStmt(newModule); err != nil {
				return err
			}
			updated, err = parsed.Format()
			if err != nil {
				return err
			}
		} else {
			// 只替换 Go 导入，不能改动模板来源和 CLI 升级地址等字符串。
			parsed, err := parser.ParseFile(token.NewFileSet(), path, data, parser.ImportsOnly)
			if err != nil {
				return fmt.Errorf("解析导入 %s：%w", path, err)
			}
			for i := len(parsed.Imports) - 1; i >= 0; i-- {
				item := parsed.Imports[i].Path
				value, err := strconv.Unquote(item.Value)
				if err != nil {
					return err
				}
				if value != oldModule && !strings.HasPrefix(value, oldModule+"/") {
					continue
				}
				replacement := []byte(strconv.Quote(newModule + strings.TrimPrefix(value, oldModule)))
				begin, end := int(item.Pos())-1, int(item.End())-1
				updated = append(append(append([]byte{}, updated[:begin]...), replacement...), updated[end:]...)
			}
		}
		return os.WriteFile(path, updated, 0644)
	})
}
