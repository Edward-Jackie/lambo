package create

import (
	"fmt"
	"path/filepath"

	"github.com/Edward-Jackie/lambo/internal/pkg/helper"
	"github.com/spf13/cobra"
)

var rootPath, schemaPath, templatePath string
var dryRun, checkOnly bool

var CmdCreate = &cobra.Command{
	Use: "create", Short: "生成后端分层组件或 Vben Element Plus 业务页面",
	Example: "lambo create frontend category --schema examples/category.json --dry-run",
	Args:    cobra.NoArgs,
	RunE:    func(cmd *cobra.Command, _ []string) error { return cmd.Help() },
}

func init() {
	CmdCreate.PersistentFlags().StringVar(&rootPath, "root", "", "Lambo 项目根目录；默认向上查找")
	CmdCreate.PersistentFlags().StringVar(&schemaPath, "schema", "", "字段定义 JSON；省略时生成名称字段示例")
	CmdCreate.PersistentFlags().StringVarP(&templatePath, "tpl-path", "t", "", "自定义模板目录（可只替换部分模板）")
	CmdCreate.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "预览文件内容，不写入磁盘")
	CmdCreate.PersistentFlags().BoolVar(&checkOnly, "check", false, "只检查产物是否与模板一致，不写入磁盘")
	for _, kind := range []string{"handler", "service", "repository", "model", "backend", "frontend", "all"} {
		kind := kind
		CmdCreate.AddCommand(&cobra.Command{
			Use: kind + " <名称>", Short: "生成 " + kind,
			Args: cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error { return runCreate(cmd, kind, args[0]) },
		})
	}
}

func runCreate(cmd *cobra.Command, kind, name string) error {
	if dryRun && checkOnly {
		return fmt.Errorf("--dry-run 和 --check 不能同时使用")
	}
	root, err := helper.FindLamboRoot(rootPath)
	if err != nil {
		return err
	}
	module, err := helper.ReadModulePath(filepath.Join(root, "backend"))
	if err != nil {
		return err
	}
	spec, err := loadSpec(name, schemaPath)
	if err != nil {
		return err
	}
	data := templateData{Spec: spec, Module: module}
	if kind == "frontend" || kind == "all" {
		if err := checkRouteName(root, spec); err != nil {
			return err
		}
	}
	files, err := renderFiles(kind, data)
	if err != nil {
		return err
	}
	if err := publishFiles(root, files, dryRun, checkOnly, cmd.OutOrStdout()); err != nil {
		return err
	}
	if !dryRun && !checkOnly {
		fmt.Fprintln(cmd.OutOrStdout(), "生成完成。已有手写文件不会覆盖；后端组件需接入 Wire、路由和真实业务实现。")
		if kind == "frontend" || kind == "all" {
			fmt.Fprintf(cmd.OutOrStdout(), "前端路由：/business/%s；接口契约：%s。请配置相应权限码并联调真实后端。\n", spec.Name, spec.API)
		}
	}
	return nil
}
