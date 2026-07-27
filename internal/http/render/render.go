package render

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var funcMap = template.FuncMap{
	"TitleCase":    TitleCase,
	"MarshallJson": MarshallJson,
}

type TemplateManager struct {
	templates map[string]*template.Template
}

// NewTemplateManager compiles all pages alongside layouts and fragments
func NewTemplateManager(templatesDir string) (*TemplateManager, error) {
	cache := make(map[string]*template.Template)

	layoutsDir := filepath.Join(templatesDir, "layouts")
	fragmentsDir := filepath.Join(templatesDir, "fragments")
	pagesDir := filepath.Join(templatesDir, "pages")

	// Collect all shared files (layouts and fragments)
	var sharedFiles []string
	for _, dir := range []string{layoutsDir, fragmentsDir} {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}

		err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() && strings.HasSuffix(d.Name(), ".html") {
				sharedFiles = append(sharedFiles, p)
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("error walking shared files in %s: %w", dir, err)
		}
	}

	// Walk the pages/ directory and compile each page template
	err := filepath.WalkDir(pagesDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".html") {
			return nil
		}

		// Get relative key inside pages (e.g., "auth/login.html" or "dashboard/index.html")
		relPath, err := filepath.Rel(pagesDir, p)
		if err != nil {
			return err
		}
		key := filepath.ToSlash(relPath) // Normalize slashes for cross-platform compatibility (Windows/Linux)

		// Files for this page: the page itself + all shared layouts and fragments
		files := append([]string{p}, sharedFiles...)

		// Important: Register funcMap BEFORE ParseFiles
		tmpl := template.New(filepath.Base(p)).Funcs(funcMap)
		tmpl, err = tmpl.ParseFiles(files...)
		if err != nil {
			return fmt.Errorf("error parsing template %s: %w", key, err)
		}

		cache[key] = tmpl
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error loading pages: %w", err)
	}

	return &TemplateManager{templates: cache}, nil
}

// RenderPage executes the full page template using the "base" block
func (tm *TemplateManager) RenderPage(w io.Writer, name string, data any) error {
	tmpl, ok := tm.templates[name]
	if !ok {
		return fmt.Errorf("template not found: %s", name)
	}
	// Executes the template defining the main block (e.g., "base" or "base_layout")
	return tmpl.ExecuteTemplate(w, "base", data)
}

// RenderPartial executes only a specific fragment
func (tm *TemplateManager) RenderPartial(w io.Writer, page string, partial string, data any) error {
	tmpl, ok := tm.templates[page]
	if !ok {
		return fmt.Errorf("partial template not found for page: %s", page)
	}
	return tmpl.ExecuteTemplate(w, partial, data)
}

// ===========================================================
// =                 Template Helper Functions               =
// ===========================================================

func MarshallJson(v any) template.JS {
	b, err := json.Marshal(v)
	if err != nil {
		return "[]"
	}
	return template.JS(b)
}

func TitleCase(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	first := strings.ToUpper(string(r[0]))
	rest := string(r[1:])
	return first + rest
}
