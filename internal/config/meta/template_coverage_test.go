package meta_test

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/artalkjs/artalk/v2/internal/config"
	"github.com/artalkjs/artalk/v2/internal/config/meta"
	"github.com/artalkjs/artalk/v2/internal/pkged"
)

var templateLocales = []string{"en", "zh-CN", "zh-TW", "tr", "simple"}

func TestConfigTemplatesCoverConfigLeafPaths(t *testing.T) {
	setProjectFS(t)

	configPaths := configLeafPathSet()
	for _, locale := range templateLocales {
		t.Run(locale, func(t *testing.T) {
			templatePaths := templateLeafPathSet(t, locale)

			for path := range configPaths {
				if isTemplateCoverageAllowedMissing(path) {
					continue
				}
				if _, ok := templatePaths[path]; !ok {
					t.Errorf("template %q is missing config path %q", locale, path)
				}
			}
		})
	}
}

func TestConfigTemplateLeafPathsMatchAcrossLocales(t *testing.T) {
	setProjectFS(t)

	baseline := filteredTemplateLeafPathSet(t, "en")
	for _, locale := range templateLocales[1:] {
		t.Run(locale, func(t *testing.T) {
			current := filteredTemplateLeafPathSet(t, locale)

			for _, path := range sortedDiff(baseline, current) {
				t.Errorf("template %q is missing path %q from English template", locale, path)
			}
			for _, path := range sortedDiff(current, baseline) {
				t.Errorf("template %q has extra path %q not present in English template", locale, path)
			}
		})
	}
}

func TestConfigTemplateSelectorOptionsMatchAcrossLocales(t *testing.T) {
	setProjectFS(t)

	baseline := templateSelectorOptions(t, "en")
	for _, locale := range []string{"zh-CN", "zh-TW", "tr"} {
		t.Run(locale, func(t *testing.T) {
			current := templateSelectorOptions(t, locale)

			for path, expected := range baseline {
				actual, ok := current[path]
				if !ok {
					t.Errorf("template %q selector %q is missing", locale, path)
					continue
				}
				if !slices.Equal(expected, actual) {
					t.Errorf("template %q selector %q = %#v, expected %#v", locale, path, actual, expected)
				}
			}
		})
	}
}

func setProjectFS(t *testing.T) {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve test file path")
	}

	rootDir := filepath.Clean(filepath.Join(filepath.Dir(filename), "../../.."))
	pkged.SetFS(os.DirFS(rootDir))
}

func configLeafPathSet() map[string]struct{} {
	paths := map[string]struct{}{}
	collectConfigLeafPaths(reflect.TypeOf(config.Config{}), "", paths)
	return paths
}

func collectConfigLeafPaths(t reflect.Type, rootPath string, paths map[string]struct{}) {
	t = derefType(t)
	if t.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		name := jsonFieldName(field)
		if name == "" || name == "-" {
			continue
		}

		path := name
		if rootPath != "" {
			path = rootPath + "." + name
		}

		fieldType := derefType(field.Type)
		switch fieldType.Kind() {
		case reflect.Struct:
			collectConfigLeafPaths(fieldType, path, paths)
		case reflect.Slice, reflect.Array:
			elemType := derefType(fieldType.Elem())
			if elemType.Kind() == reflect.Struct {
				collectConfigLeafPaths(elemType, path+".$$", paths)
			} else {
				paths[path] = struct{}{}
			}
		default:
			paths[path] = struct{}{}
		}
	}
}

func derefType(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

func jsonFieldName(field reflect.StructField) string {
	return strings.Split(field.Tag.Get("json"), ",")[0]
}

func templateLeafPathSet(t *testing.T, locale string) map[string]struct{} {
	t.Helper()

	metas, err := meta.GetOptionsMetaData(config.Template(locale))
	if err != nil {
		t.Fatalf("parse template %q: %v", locale, err)
	}

	paths := map[string]struct{}{}
	for _, item := range metas {
		if item.AllowsSet {
			paths[item.Path] = struct{}{}
		}
	}
	return paths
}

func filteredTemplateLeafPathSet(t *testing.T, locale string) map[string]struct{} {
	t.Helper()

	paths := templateLeafPathSet(t, locale)
	for path := range paths {
		if isTemplateFieldSetAllowedDifferent(path) {
			delete(paths, path)
		}
	}
	return paths
}

func templateSelectorOptions(t *testing.T, locale string) map[string][]string {
	t.Helper()

	metas, err := meta.GetOptionsMetaData(config.Template(locale))
	if err != nil {
		t.Fatalf("parse template %q: %v", locale, err)
	}

	options := map[string][]string{}
	for _, item := range metas {
		if len(item.Options) > 0 {
			options[item.Path] = item.Options
		}
	}
	return options
}

func isTemplateCoverageAllowedMissing(path string) bool {
	return path == "frontend" || path == "admin_users" || strings.HasPrefix(path, "admin_users.")
}

func isTemplateFieldSetAllowedDifferent(path string) bool {
	return path == "admin_users" || strings.HasPrefix(path, "admin_users.")
}

func sortedDiff(a map[string]struct{}, b map[string]struct{}) []string {
	diff := []string{}
	for path := range a {
		if _, ok := b[path]; !ok {
			diff = append(diff, path)
		}
	}
	slices.Sort(diff)
	return diff
}
