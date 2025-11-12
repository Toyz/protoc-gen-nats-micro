package generator

import (
	"bytes"
	"fmt"
	"text/template"

	"google.golang.org/protobuf/compiler/protogen"
)

// TypeScriptLanguage implements Language for TypeScript code generation
type TypeScriptLanguage struct {
	templates *template.Template
}

// NewTypeScriptLanguage creates a new TypeScript language generator
func NewTypeScriptLanguage() *TypeScriptLanguage {
	tmpl := template.Must(template.New("ts").Funcs(FuncMap()).ParseFS(templatesFS, "templates/ts/*.tmpl"))
	return &TypeScriptLanguage{
		templates: tmpl,
	}
}

func (l *TypeScriptLanguage) Name() string {
	return "typescript"
}

func (l *TypeScriptLanguage) FileExtension() string {
	return "_nats.pb.ts"
}

func (l *TypeScriptLanguage) Generate(g *protogen.GeneratedFile, file *protogen.File, service *protogen.Service, opts ServiceOptions) error {
	data := TemplateData{
		File:    file,
		Service: service,
		Options: opts,
	}

	// Generate error types for this service
	var errorsBuf bytes.Buffer
	if err := l.templates.ExecuteTemplate(&errorsBuf, "errors.ts.tmpl", data); err != nil {
		return fmt.Errorf("execute errors template: %w", err)
	}
	g.P(errorsBuf.String())
	g.P()

	// Generate service
	var serviceBuf bytes.Buffer
	if err := l.templates.ExecuteTemplate(&serviceBuf, "service.ts.tmpl", data); err != nil {
		return fmt.Errorf("execute service template: %w", err)
	}
	g.P(serviceBuf.String())
	g.P()

	// Generate client
	var clientBuf bytes.Buffer
	if err := l.templates.ExecuteTemplate(&clientBuf, "client.ts.tmpl", data); err != nil {
		return fmt.Errorf("execute client template: %w", err)
	}
	g.P(clientBuf.String())
	g.P()

	return nil
}

// GenerateHeader generates the file header (imports)
func (l *TypeScriptLanguage) GenerateHeader(g *protogen.GeneratedFile, file *protogen.File) error {
	data := TemplateData{
		File: file,
	}

	var headerBuf bytes.Buffer
	if err := l.templates.ExecuteTemplate(&headerBuf, "header.ts.tmpl", data); err != nil {
		return fmt.Errorf("execute header template: %w", err)
	}
	g.P(headerBuf.String())
	g.P()

	return nil
}

// GenerateShared generates shared types once per proto file
func (l *TypeScriptLanguage) GenerateShared(g *protogen.GeneratedFile, file *protogen.File) error {
	data := TemplateData{
		File: file,
	}

	// Generate minimal header for shared file
	var headerBuf bytes.Buffer
	if err := l.templates.ExecuteTemplate(&headerBuf, "shared_header.ts.tmpl", data); err != nil {
		return fmt.Errorf("execute shared header template: %w", err)
	}
	g.P(headerBuf.String())
	g.P()

	// Generate shared types (currently minimal for TypeScript)
	var sharedBuf bytes.Buffer
	if err := l.templates.ExecuteTemplate(&sharedBuf, "shared.ts.tmpl", data); err != nil {
		return fmt.Errorf("execute shared template: %w", err)
	}
	g.P(sharedBuf.String())
	g.P()

	return nil
}
