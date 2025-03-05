package pages

import (
	"embed"
	"fmt"
	"github.com/Masterminds/sprig"
	"github.com/pocketbase/pocketbase/core"
	"html/template"
	"path/filepath"
)

//go:embed auth/*.gohtml battle/*.gohtml index/*.gohtml layout/*.gohtml leader/*.gohtml prompt/*.gohtml
var templates embed.FS

func Render(e *core.RequestEvent, templ *template.Template, filename string, data any) error {
	clone, cloneErr := templ.Clone()
	if cloneErr != nil {
		return cloneErr
	}

	// Read template content from embedded filesystem
	content, err := templates.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read template %s: %v", filename, err)
	}

	// Parse the template content
	temp, parseErr := clone.Parse(string(content))
	if parseErr != nil {
		return fmt.Errorf("parse error %w", parseErr)
	}

	return temp.Execute(e.Response, data)
}

func ParseTemplates() (*template.Template, error) {
	templ := template.New("").Funcs(sprig.FuncMap())

	// Parse layout templates from embedded filesystem
	layoutFiles, err := templates.ReadDir("layout")
	if err != nil {
		return nil, err
	}

	for _, file := range layoutFiles {
		if filepath.Ext(file.Name()) == ".gohtml" {
			content, err := templates.ReadFile(filepath.Join("layout", file.Name()))
			if err != nil {
				return nil, err
			}
			templ, err = templ.New(file.Name()).Parse(string(content))
			if err != nil {
				return nil, err
			}
		}
	}

	return templ, nil
}
