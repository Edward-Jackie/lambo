package create

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerationContract(t *testing.T) {
	spec, err := loadSpec("category", "../../../examples/category.json")
	if err != nil {
		t.Fatal(err)
	}
	files, err := renderFiles("all", templateData{Spec: spec, Module: "example.com/shop/backend"})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 9 {
		t.Fatalf("产物数 %d", len(files))
	}
	root := t.TempDir()
	out := &bytes.Buffer{}
	if err := publishFiles(root, files, true, false, out); err != nil {
		t.Fatal(err)
	}
	if entries, _ := os.ReadDir(root); len(entries) != 0 {
		t.Fatal("预览写入了目录")
	}
	if err := publishFiles(root, files, false, true, out); err == nil {
		t.Fatal("缺失产物检查应失败")
	}
	if err := publishFiles(root, files, false, false, out); err != nil {
		t.Fatal(err)
	}
	if err := publishFiles(root, files, false, true, out); err != nil {
		t.Fatal(err)
	}
	if err := publishFiles(root, files, false, false, out); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, files[1].Path)
	if err := os.WriteFile(path, []byte("手写内容"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(root, files[0].Path)); err != nil {
		t.Fatal(err)
	}
	if err := publishFiles(root, files, false, false, out); err == nil {
		t.Fatal("应拒绝覆盖")
	}
	if data, _ := os.ReadFile(path); string(data) != "手写内容" {
		t.Fatal("手写文件被覆盖")
	}
	if _, err := os.Stat(filepath.Join(root, files[0].Path)); !os.IsNotExist(err) {
		t.Fatal("冲突时生成了部分文件")
	}
}

func TestSchemaValidation(t *testing.T) {
	cases := []struct{ name, spec string }{
		{"unknown", `{"fields":[{"name":"name","label":"名称","type":"string","wat":true}]}`},
		{"type", `{"fields":[{"name":"name","label":"名称","type":"object"}]}`},
		{"duplicate", `{"fields":[{"name":"name","label":"名称","type":"string"},{"name":"name","label":"名称","type":"string"}]}`},
		{"system", `{"fields":[{"name":"id","label":"名称","type":"string"}]}`},
		{"fieldpath", `{"fields":[{"name":"../name","label":"名称","type":"string"}]}`},
		{"default", `{"fields":[{"name":"status","label":"状态","type":"number","default":"1"}]}`},
		{"options", `{"fields":[{"name":"status","label":"状态","type":"number","options":[{"label":"启用","value":1},{"label":"停用","value":1}]}]}`},
		{"numericduplicate", `{"fields":[{"name":"status","label":"状态","type":"number","options":[{"label":"启用","value":1},{"label":"停用","value":1.0}]}]}`},
		{"unsafeinteger", `{"fields":[{"name":"count","label":"数量","type":"number","default":9007199254740992}]}`},
		{"optiondefault", `{"fields":[{"name":"status","label":"状态","type":"number","default":2,"options":[{"label":"启用","value":1}]}]}`},
		{"readonly", `{"fields":[{"name":"name","label":"名称","type":"string","readOnly":true}]}`},
		{"api", `{"api":"/api/../users","fields":[{"name":"name","label":"名称","type":"string"}]}`},
		{"trailing", `{} {}`},
		{"name", `{"name":"other"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "spec.json")
			if err := os.WriteFile(path, []byte(tc.spec), 0644); err != nil {
				t.Fatal(err)
			}
			if _, err := loadSpec("category", path); err == nil {
				t.Fatal("无效定义应失败")
			}
		})
	}
	for _, name := range []string{"../evil", "/tmp/evil", "user.go", "User", ""} {
		if _, err := loadSpec(name, ""); err == nil {
			t.Fatalf("无效名称 %q", name)
		}
	}
}

func TestTypedValuesAndReadOnly(t *testing.T) {
	path := filepath.Join(t.TempDir(), "spec.json")
	data := `{"idType":"string","title":"</script>","fields":[{"name":"enabled","label":"启用","type":"boolean","required":true,"search":true,"default":false},{"name":"count","label":"数量","type":"number","default":0},{"name":"code","label":"业务编号","type":"string","readOnly":true}]}`
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}
	spec, err := loadSpec("sample-item", path)
	if err != nil {
		t.Fatal(err)
	}
	if spec.Fields[0].DefaultJSON() != "false" || spec.Fields[1].DefaultJSON() != "0" {
		t.Fatal("丢失零值")
	}
	files, err := renderFiles("all", templateData{Spec: spec, Module: "example.com/shop/backend"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(files[5].Data), "</script>") {
		t.Fatal("未转义脚本结束标记")
	}
	if strings.Contains(string(files[5].Data), `fieldName: "code"`) {
		t.Fatal("只读字段进入表单")
	}
	var values interface{}
	if err := json.Unmarshal([]byte(jsonValue(spec.Fields[0].Choices())), &values); err != nil {
		t.Fatal(err)
	}
}

func TestPathBoundaries(t *testing.T) {
	root, outside := t.TempDir(), t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "frontend")); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"../escape", "/tmp/escape", "frontend/escape.ts"} {
		if err := publishFiles(root, []outputFile{{Path: path, Data: []byte("bad")}}, false, false, &bytes.Buffer{}); err == nil {
			t.Fatalf("允许非法路径 %s", path)
		}
	}
	if entries, _ := os.ReadDir(outside); len(entries) != 0 {
		t.Fatal("越界写入")
	}
}

func TestRouteNameCollision(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "frontend/apps/web-ele/src/router/routes/modules/business/user-profile.ts")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(`export default [{ name: 'BusinessUserProfile' }];`), 0644); err != nil {
		t.Fatal(err)
	}
	original, err := loadSpec("user-profile", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := checkRouteName(root, original); err != nil {
		t.Fatal(err)
	}
	other, err := loadSpec("user_profile", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := checkRouteName(root, other); err == nil {
		t.Fatal("未拒绝重复路由名称")
	}
}

func TestBackendDeclarationAndAPICollision(t *testing.T) {
	spec, err := loadSpec("category", "../../../examples/category.json")
	if err != nil {
		t.Fatal(err)
	}
	files, err := renderFiles("backend", templateData{Spec: spec, Module: "example.com/shop/backend"})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ path, content string }{
		{"backend/internal/business/model/shared.go", "package model\ntype CategoryPayload struct{}\n"},
		{"backend/internal/business/handler/other.go", "package handler\nconst OtherAPIPath = \"/api/categories\"\n"},
	} {
		root := t.TempDir()
		path := filepath.Join(root, tc.path)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(tc.content), 0644); err != nil {
			t.Fatal(err)
		}
		if err := checkGoConflicts(root, files); err == nil {
			t.Fatalf("未拒绝命名或 API 冲突：%s", tc.path)
		}
	}
}

func TestCRUDSchemaValidation(t *testing.T) {
	for _, value := range []string{
		`{"fields":[{"name":"validate","label":"校验","type":"string"}]}`,
		`{"fields":[{"name":"applyDefaults","label":"默认值","type":"string"}]}`,
		`{"fields":[{"name":"name","label":"名称","type":"string","required":true,"default":" "}]}`,
		`{"fields":[{"name":"name","label":"名称","type":"string","maxLength":1,"options":[{"label":"表情","value":"😀"}]}]}`,
	} {
		path := filepath.Join(t.TempDir(), "schema.json")
		if err := os.WriteFile(path, []byte(value), 0644); err != nil {
			t.Fatal(err)
		}
		if _, err := loadSpec("sample", path); err == nil {
			t.Fatalf("应拒绝无效定义：%s", value)
		}
	}
	spec := Spec{Name: "sample", IDType: "number", Fields: []Field{{Name: "name", Label: "名称", Type: "string"}, {Name: "state", Label: "状态", Type: "string", ReadOnly: true, Required: true}}}
	if _, err := renderFiles("backend", templateData{Spec: spec, Module: "example.com/shop/backend"}); err == nil {
		t.Fatal("只读必填字段缺少初始值时应阻止生成完整 CRUD")
	}
}
