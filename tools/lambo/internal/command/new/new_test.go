package new

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReplaceModulesWithoutTouchingOtherTrees(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{
		"go.mod":         "module example.com/lambo\n\ngo 1.25\n",
		"main.go":        "package main\nimport _ \"example.com/lambo/cmd/lambo\"\n",
		"backend/go.mod": "module example.com/lambo/backend\n",
		"config.go": `package main
const Upgrade = "example.com/lambo/cmd/lambo@latest"
`,
		"frontend/keep.go": "example.com/lambo",
		".git/keep.go":     "example.com/lambo",
	}
	for name, content := range files {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err := replaceModule(root, "example.com/lambo", "example.com/shop"); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(filepath.Join(root, "main.go"))
	if !strings.Contains(string(data), "example.com/shop/cmd/lambo") {
		t.Fatal("导入未改名")
	}
	for _, name := range []string{"backend/go.mod", "frontend/keep.go", ".git/keep.go", "config.go"} {
		data, _ := os.ReadFile(filepath.Join(root, name))
		if string(data) != files[name] {
			t.Fatalf("越界修改 %s", name)
		}
	}
}

func TestReplaceModuleSkipsNestedToolModule(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "tools", "lambo")
	if err := os.MkdirAll(nested, 0755); err != nil {
		t.Fatal(err)
	}
	original := "module example.com/lambo/tools/lambo\n\ngo 1.25.8\n"
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/lambo/backend\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nested, "go.mod"), []byte(original), 0644); err != nil {
		t.Fatal(err)
	}
	if err := replaceModule(root, "example.com/lambo/backend", "example.com/shop/backend"); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(nested, "go.mod"))
	if err != nil || string(data) != original {
		t.Fatalf("嵌套模块被业务模块改写：%s，%v", data, err)
	}
}
