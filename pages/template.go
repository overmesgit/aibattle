package pages

import (
	"fmt"
	"github.com/Masterminds/sprig"
	"github.com/pocketbase/pocketbase/core"
	"html/template"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func Render(e *core.RequestEvent, templ *template.Template, filename string, data any) error {
	_, file, _, _ := runtime.Caller(1)
	clone, cloneErr := templ.Clone()
	if cloneErr != nil {
		return cloneErr
	}
	dir := filepath.Dir(file)
	splitPath := strings.Split(dir, "aibattle/")
	if len(splitPath) < 2 {
		return fmt.Errorf("invalid dir for template: %s", dir)
	}
	absolute := splitPath[1]
	_, parseErr := clone.ParseFiles(filepath.Join(absolute, filename))
	if parseErr != nil {
		return parseErr
	}
	return clone.ExecuteTemplate(e.Response, filename, data)
}

func ParseTemplates() (*template.Template, error) {
	templ := template.New("").Funcs(sprig.FuncMap())

	// Parse all .html files in templates directory
	err := filepath.Walk(
		"pages/layout", func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() && filepath.Ext(path) == ".gohtml" {
				_, err = templ.ParseFiles(path)
				if err != nil {
					return err
				}
			}
			return nil
		},
	)

	return templ, err
}
