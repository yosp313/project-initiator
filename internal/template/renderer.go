// Package template provides template rendering functionality.
package template

import (
	"bytes"
	"fmt"
	"text/template"
)

// Renderer handles template rendering.
type Renderer struct {
	funcMap template.FuncMap
}

// NewRenderer creates a new template renderer.
func NewRenderer() *Renderer {
	return &Renderer{
		funcMap: template.FuncMap{},
	}
}

// Render parses and executes a template with the given data.
func (r *Renderer) Render(source string, data any) (string, error) {
	tmpl, err := template.New("template").Funcs(r.funcMap).Parse(source)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	return buf.String(), nil
}
