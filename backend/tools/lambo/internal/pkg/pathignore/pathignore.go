package pathignore

import (
	"fmt"
	"path"
	"runtime"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

// Matcher 按 gitignore 路径语义匹配逗号分隔的规则。
// 无斜杠规则匹配任意层级名称，有斜杠规则相对项目根目录匹配。
type Matcher struct {
	rules []rule
}

type rule struct {
	pattern       string
	negated       bool
	directoryOnly bool
	basenameOnly  bool
	contentsOnly  bool
	contentsBase  string
}

// CompileCSV 编译 lambo --excludeDir 接受的忽略规则。
func CompileCSV(value string) (*Matcher, error) {
	parts := strings.Split(value, ",")
	rules := make([]rule, 0, len(parts))
	for index, part := range parts {
		pattern := strings.TrimSpace(part)
		if pattern == "" || strings.HasPrefix(pattern, "#") {
			continue
		}

		negated := strings.HasPrefix(pattern, "!")
		if negated {
			pattern = strings.TrimSpace(strings.TrimPrefix(pattern, "!"))
		}
		if pattern == "" {
			return nil, fmt.Errorf("exclude pattern %d has no value after the negation marker", index+1)
		}

		// 统一路径分隔符，让同一份配置可在不同系统的命令行中复用。
		pattern = strings.ReplaceAll(pattern, `\`, "/")
		directoryOnly := strings.HasSuffix(pattern, "/")
		pattern = strings.TrimSuffix(pattern, "/")
		pattern = strings.TrimPrefix(pattern, "/")
		pattern = strings.TrimPrefix(pattern, "./")
		pattern = path.Clean(pattern)
		if pattern == "." || pattern == "" {
			return nil, fmt.Errorf("exclude pattern %d does not identify a path", index+1)
		}
		if !doublestar.ValidatePattern(pattern) {
			return nil, fmt.Errorf("invalid exclude pattern %q", part)
		}

		contentsBase := strings.TrimSuffix(pattern, "/**")
		rules = append(rules, rule{
			pattern:       normalizeCase(pattern),
			negated:       negated,
			directoryOnly: directoryOnly,
			basenameOnly:  !strings.Contains(pattern, "/"),
			contentsOnly:  contentsBase != pattern,
			contentsBase:  normalizeCase(contentsBase),
		})
	}
	return &Matcher{rules: rules}, nil
}

// Match 判断项目内相对路径是否应忽略；目录规则需要正确的 isDir。
func (m *Matcher) Match(name string, isDir bool) bool {
	if m == nil {
		return false
	}
	name = normalizePath(name)
	if name == "" {
		return false
	}

	ignored := false
	for _, rule := range m.rules {
		if rule.matches(name, isDir) {
			ignored = !rule.negated
		}
	}
	return ignored
}

func (r rule) matches(name string, isDir bool) bool {
	parts := strings.Split(name, "/")
	if r.basenameOnly {
		for index, part := range parts {
			candidateIsDir := index < len(parts)-1 || isDir
			if r.directoryOnly && !candidateIsDir {
				continue
			}
			if doublestar.MatchUnvalidated(r.pattern, part) {
				return true
			}
		}
		return false
	}

	for index := range parts {
		candidate := strings.Join(parts[:index+1], "/")
		candidateIsDir := index < len(parts)-1 || isDir
		if r.directoryOnly && !candidateIsDir {
			continue
		}
		// dir/** 只忽略目录内容，便于 !dir/keep.go 重新包含直接子文件。
		if r.contentsOnly && candidate == r.contentsBase {
			continue
		}
		if doublestar.MatchUnvalidated(r.pattern, candidate) {
			return true
		}
	}
	return false
}

func normalizePath(name string) string {
	name = strings.ReplaceAll(name, `\`, "/")
	name = strings.TrimPrefix(name, "./")
	name = strings.TrimPrefix(name, "/")
	name = path.Clean(name)
	if name == "." {
		return ""
	}
	return normalizeCase(name)
}

func normalizeCase(value string) string {
	if runtime.GOOS == "windows" {
		return strings.ToLower(value)
	}
	return value
}
