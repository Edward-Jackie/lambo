package create

import (
	"bytes"
	"errors"
	"fmt"
	"go/format"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/Edward-Jackie/lambo/tpl"
)

type outputFile struct {
	Path string
	Data []byte
}

// checkRouteName 避免不同拼写生成相同的 Vben 路由名称。
func checkRouteName(root string, spec Spec) error {
	base := filepath.Join(root, "frontend/apps/web-ele/src/router/routes/modules")
	target := filepath.Join(base, "business", spec.Name+".ts")
	pattern := regexp.MustCompile(`\bname\s*:\s*['"]Business` + regexp.QuoteMeta(spec.StructName()) + `['"]`)
	return filepath.WalkDir(base, func(path string, entry fs.DirEntry, err error) error {
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return err
		}
		if entry.IsDir() || path == target || filepath.Ext(path) != ".ts" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if pattern.Match(data) {
			return fmt.Errorf("路由名称 Business%s 已由 %s 使用", spec.StructName(), path)
		}
		return nil
	})
}

func renderFiles(kind string, data templateData) ([]outputFile, error) {
	kinds := []string{kind}
	if kind == "backend" || kind == "all" {
		kinds = []string{"model", "repository", "service", "handler"}
	}
	if kind == "all" {
		kinds = append(kinds, "frontend")
	}
	files := []outputFile{}
	for _, current := range kinds {
		paths := [][2]string{}
		if current == "frontend" {
			src := "frontend/apps/web-ele/src/"
			view := src + "views/business/" + data.Name + "/"
			paths = [][2]string{
				{"frontend/api.ts.tpl", src + "api/business/" + data.Name + ".ts"},
				{"frontend/data.ts.tpl", view + "data.ts"},
				{"frontend/form.vue.tpl", view + "form.vue"},
				{"frontend/index.vue.tpl", view + "index.vue"},
				{"frontend/route.ts.tpl", src + "router/routes/modules/business/" + data.Name + ".ts"},
			}
		} else {
			paths = [][2]string{{"create/" + current + ".tpl", "backend/internal/" + current + "/" + data.SnakeName() + ".go"}}
		}
		for _, pair := range paths {
			source, err := tpl.CreateTemplateFS.ReadFile(pair[0])
			if err != nil {
				return nil, err
			}
			if templatePath != "" {
				path := filepath.Join(templatePath, filepath.FromSlash(pair[0]))
				custom, err := os.ReadFile(path)
				if err == nil {
					source = custom
				} else if !os.IsNotExist(err) {
					return nil, fmt.Errorf("读取自定义模板：%w", err)
				}
			}
			t, err := template.New(pair[0]).Funcs(template.FuncMap{"json": jsonValue}).Option("missingkey=error").Parse(string(source))
			if err != nil {
				return nil, fmt.Errorf("解析模板 %s：%w", pair[0], err)
			}
			var rendered bytes.Buffer
			if err := t.Execute(&rendered, data); err != nil {
				return nil, fmt.Errorf("执行模板 %s：%w", pair[0], err)
			}
			content := rendered.Bytes()
			if strings.HasSuffix(pair[1], ".go") {
				content, err = format.Source(content)
				if err != nil {
					return nil, fmt.Errorf("生成 Go 语法错误 %s：%w", pair[1], err)
				}
			}
			files = append(files, outputFile{Path: pair[1], Data: content})
		}
	}
	return files, nil
}

// publishFiles 先检查整批产物，再以独占创建写入；冲突不会覆盖，写入失败撤销本批新文件。
func publishFiles(root string, files []outputFile, preview, check bool, out io.Writer) error {
	pending := []outputFile{}
	var problems []error
	for _, file := range files {
		if err := safePath(root, file.Path); err != nil {
			return err
		}
		existing, err := os.ReadFile(filepath.Join(root, file.Path))
		state := "新增"
		switch {
		case err == nil && bytes.Equal(existing, file.Data):
			state = "一致"
		case err == nil:
			state = "冲突（保留原文件）"
			problems = append(problems, fmt.Errorf("%s 已有不同内容，请手动合并或调整名称", file.Path))
		case os.IsNotExist(err):
			pending = append(pending, file)
			if check {
				problems = append(problems, fmt.Errorf("缺少生成文件 %s", file.Path))
			}
		default:
			return fmt.Errorf("读取目标 %s：%w", file.Path, err)
		}
		fmt.Fprintf(out, "%s %s\n", state, file.Path)
		if preview {
			fmt.Fprintf(out, "--- %s ---\n%s\n", file.Path, file.Data)
		}
	}
	if len(problems) > 0 {
		return errors.Join(problems...)
	}
	if preview || check {
		return nil
	}
	created := []string{}
	success := false
	defer func() {
		if !success {
			for _, path := range created {
				_ = os.Remove(path)
			}
		}
	}()
	for _, file := range pending {
		if err := safePath(root, file.Path); err != nil {
			return err
		}
		path := filepath.Join(root, file.Path)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
		// 再检查一次新建目录，拒绝经过符号链接写到项目之外。
		if err := safePath(root, file.Path); err != nil {
			return err
		}
		handle, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
		if err != nil {
			return fmt.Errorf("独占创建 %s：%w", file.Path, err)
		}
		created = append(created, path)
		_, writeErr := handle.Write(file.Data)
		closeErr := handle.Close()
		if err := errors.Join(writeErr, closeErr); err != nil {
			return fmt.Errorf("写入 %s：%w", file.Path, err)
		}
	}
	success = true
	return nil
}

func safePath(root, relative string) error {
	if filepath.IsAbs(relative) || !filepath.IsLocal(relative) {
		return fmt.Errorf("拒绝项目外路径 %s", relative)
	}
	current := root
	for _, part := range strings.Split(filepath.FromSlash(relative), string(filepath.Separator)) {
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("目标路径包含符号链接：%s", current)
		}
	}
	return nil
}
