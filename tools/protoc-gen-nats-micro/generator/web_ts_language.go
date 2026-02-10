package generator

import (
	"bytes"
	"fmt"
	"text/template"

	"google.golang.org/protobuf/compiler/protogen"
)

// WebTSLanguage implements Language for web-focused TypeScript code generation
// It generates client-only code compatible with protoc-gen-es v2 (@bufbuild/protobuf)
type WebTSLanguage struct {
	templates *template.Template
}

// NewWebTSLanguage creates a new Web TypeScript language generator
func NewWebTSLanguage() *WebTSLanguage {
	tmpl := template.Must(template.New("web-ts").Funcs(FuncMap()).ParseFS(templatesFS, "templates/web-ts/*.tmpl"))
	return &WebTSLanguage{
		templates: tmpl,
	}
}

func (l *WebTSLanguage) Name() string {
	return "web-ts"
}

func (l *WebTSLanguage) FileExtension() string {
	return "_nats.pb.ts"
}

func (l *WebTSLanguage) Generate(g *protogen.GeneratedFile, file *protogen.File, service *protogen.Service, opts ServiceOptions) error {
	data := TemplateData{
		File:    file,
		Service: service,
		Options: opts,
	}

	// Generate error types
	var errorsBuf bytes.Buffer
	if err := l.templates.ExecuteTemplate(&errorsBuf, "errors.web-ts.tmpl", data); err != nil {
		return fmt.Errorf("execute errors template: %w", err)
	}
	g.P(errorsBuf.String())
	g.P()

	// Generate client only (web-ts is always client-only)
	var clientBuf bytes.Buffer
	if err := l.templates.ExecuteTemplate(&clientBuf, "client.web-ts.tmpl", data); err != nil {
		return fmt.Errorf("execute client template: %w", err)
	}
	g.P(clientBuf.String())
	g.P()

	return nil
}

// GenerateHeader generates the file header (imports)
func (l *WebTSLanguage) GenerateHeader(g *protogen.GeneratedFile, file *protogen.File) error {
	data := TemplateData{
		File: file,
	}

	var headerBuf bytes.Buffer
	if err := l.templates.ExecuteTemplate(&headerBuf, "header.web-ts.tmpl", data); err != nil {
		return fmt.Errorf("execute header template: %w", err)
	}
	g.P(headerBuf.String())
	g.P()

	return nil
}

// GenerateShared generates shared types once per proto file
func (l *WebTSLanguage) GenerateShared(g *protogen.GeneratedFile, file *protogen.File) error {
	data := TemplateData{
		File: file,
	}

	// Generate minimal header for shared file
	var headerBuf bytes.Buffer
	if err := l.templates.ExecuteTemplate(&headerBuf, "shared_header.web-ts.tmpl", data); err != nil {
		return fmt.Errorf("execute shared header template: %w", err)
	}
	g.P(headerBuf.String())
	g.P()

	// Generate shared client types
	var sharedBuf bytes.Buffer
	if err := l.templates.ExecuteTemplate(&sharedBuf, "shared.web-ts.tmpl", data); err != nil {
		return fmt.Errorf("execute shared template: %w", err)
	}
	g.P(sharedBuf.String())
	g.P()

	return nil
}
