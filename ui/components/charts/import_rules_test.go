package charts_test

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const (
	modulePath       = "github.com/wwsheng009/mint"
	chartsImportBase = modulePath + "/ui/components/charts"
	uiImportPath     = modulePath + "/ui"
	runtimeImport    = modulePath + "/runtime/"
	frameworkTheme   = modulePath + "/framework/theme"
	uiInternalImport = modulePath + "/ui/components/internal/"
	internalImport   = chartsImportBase + "/internal/"
	modelImport      = chartsImportBase + "/model"
)

func TestChartsRootHasNoCompilableGoFiles(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read charts root: %v", err)
	}

	var offenders []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		offenders = append(offenders, name)
	}

	if len(offenders) > 0 {
		t.Fatalf("charts 根目录不应包含可编译 Go 文件，发现: %s", strings.Join(offenders, ", "))
	}
}

func TestChartsImportBoundaries(t *testing.T) {
	componentNames, err := discoverConcreteComponents(".")
	if err != nil {
		t.Fatalf("discover chart components: %v", err)
	}

	fset := token.NewFileSet()
	var violations []string

	err = filepath.WalkDir(".", func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		relPath := filepath.ToSlash(strings.TrimPrefix(path, "."+string(filepath.Separator)))
		category := classifyChartsGoFile(relPath, componentNames)
		if category == fileCategoryOther {
			return nil
		}

		file, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}

		for _, spec := range file.Imports {
			importPath, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				return err
			}

			switch category {
			case fileCategoryComponent:
				if importPath == uiImportPath {
					violations = append(violations, relPath+" 反向依赖 ui 顶层包 "+uiImportPath)
				}
				if component := importedConcreteComponent(importPath, componentNames); component != "" && component != currentConcreteComponent(relPath) {
					violations = append(violations, relPath+" 反向依赖具体图表组件 "+component)
				}
			case fileCategoryInternal:
				if component := importedConcreteComponent(importPath, componentNames); component != "" {
					violations = append(violations, relPath+" 反向依赖具体图表组件 "+component)
				}
				if !isAllowedInternalImport(importPath) {
					violations = append(violations, relPath+" 使用了不在白名单内的模块内依赖 "+importPath)
				}
			}
		}

		return nil
	})
	if err != nil {
		t.Fatalf("walk charts tree: %v", err)
	}

	if len(violations) > 0 {
		t.Fatalf("charts 依赖约束检查失败:\n- %s", strings.Join(violations, "\n- "))
	}
}

type chartsFileCategory int

const (
	fileCategoryOther chartsFileCategory = iota
	fileCategoryComponent
	fileCategoryInternal
)

func classifyChartsGoFile(relPath string, componentNames map[string]struct{}) chartsFileCategory {
	parts := strings.Split(relPath, "/")
	if len(parts) == 0 {
		return fileCategoryOther
	}

	switch parts[0] {
	case "internal":
		return fileCategoryInternal
	case "model":
		return fileCategoryOther
	default:
		if _, ok := componentNames[parts[0]]; ok {
			return fileCategoryComponent
		}
		return fileCategoryOther
	}
}

func discoverConcreteComponents(root string) (map[string]struct{}, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}

	componentNames := make(map[string]struct{})
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		switch name {
		case "internal", "model", "testdata":
			continue
		}
		componentNames[name] = struct{}{}
	}
	return componentNames, nil
}

func currentConcreteComponent(relPath string) string {
	parts := strings.Split(relPath, "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}

func importedConcreteComponent(importPath string, componentNames map[string]struct{}) string {
	if !strings.HasPrefix(importPath, chartsImportBase+"/") {
		return ""
	}

	rel := strings.TrimPrefix(importPath, chartsImportBase+"/")
	if rel == "" {
		return ""
	}

	component := strings.Split(rel, "/")[0]
	if _, ok := componentNames[component]; !ok {
		return ""
	}
	return component
}

func isAllowedInternalImport(importPath string) bool {
	if !strings.HasPrefix(importPath, modulePath+"/") {
		return true
	}

	switch {
	case importPath == frameworkTheme:
		return true
	case importPath == modelImport:
		return true
	case strings.HasPrefix(importPath, runtimeImport):
		return true
	case strings.HasPrefix(importPath, uiInternalImport):
		return true
	case strings.HasPrefix(importPath, internalImport):
		return true
	default:
		return false
	}
}
