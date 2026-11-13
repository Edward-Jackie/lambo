package helper

import (
	"fmt"
	"os"
	"path/filepath"
)

// FindLamboRoot 查找同时包含前端应用与后端模块的项目，避免把 CLI 模块当成后端。
func FindLamboRoot(start string) (string, error) {
	if start == "" {
		start = "."
	}
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		backend, backendErr := os.Stat(filepath.Join(dir, "backend", "go.mod"))
		frontend, frontendErr := os.Stat(filepath.Join(dir, "frontend", "apps", "web-ele", "package.json"))
		if backendErr == nil && frontendErr == nil && !backend.IsDir() && !frontend.IsDir() {
			return filepath.EvalSymlinks(dir)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("从 %s 未找到 Lambo 项目（需要 backend/go.mod 与 frontend/apps/web-ele）", start)
		}
		dir = parent
	}
}
