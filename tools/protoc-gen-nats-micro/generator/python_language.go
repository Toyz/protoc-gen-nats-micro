package generator

import (
	"bytes"
	"fmt"
	"text/template"

	"google.golang.org/protobuf/compiler/protogen"
)

// PythonLanguage implements Language for Python code generation
type PythonLanguage struct {
	templates *template.Template
}

// NewPythonLanguage creates a new Python language generator
func NewPythonLanguage() *PythonLanguage {
	tmpl := template.Must(template.New("python").Funcs(FuncMap()).ParseFS(templatesFS, "templates/python/*.tmpl"))
	return &PythonLanguage{
		templates: tmpl,
	}
}

func (l *PythonLanguage) Name() string {
	return "python"
}

func (l *PythonLanguage) FileExtension() string {
	return "_nats_pb2.py"
}

func (l *PythonLanguage) Generate(g *protogen.GeneratedFile, file *protogen.File, service *protogen.Service, opts ServiceOptions) error {
	data := TemplateData{
		File:    file,
		Service: service,
		Options: opts,
	}

	// Generate error types for this service
	var errorsBuf bytes.Buffer
	if err := l.templates.ExecuteTemplate(&errorsBuf, "errors.py.tmpl", data); err != nil {
		return fmt.Errorf("execute errors template: %w", err)
	}
	g.P(errorsBuf.String())
	g.P()

	// Generate service
	var serviceBuf bytes.Buffer
	if err := l.templates.ExecuteTemplate(&serviceBuf, "service.py.tmpl", data); err != nil {
		return fmt.Errorf("execute service template: %w", err)
	}
	g.P(serviceBuf.String())
	g.P()

	// Generate client
	var clientBuf bytes.Buffer
	if err := l.templates.ExecuteTemplate(&clientBuf, "client.py.tmpl", data); err != nil {
		return fmt.Errorf("execute client template: %w", err)
	}
	g.P(clientBuf.String())
	g.P()

	return nil
}

// GenerateHeader generates the file header (imports)
func (l *PythonLanguage) GenerateHeader(g *protogen.GeneratedFile, file *protogen.File) error {
	data := TemplateData{
		File: file,
	}

	var headerBuf bytes.Buffer
	if err := l.templates.ExecuteTemplate(&headerBuf, "header.py.tmpl", data); err != nil {
		return fmt.Errorf("execute header template: %w", err)
	}
	g.P(headerBuf.String())
	g.P()

	return nil
}

// GenerateShared generates shared types once per proto file
func (l *PythonLanguage) GenerateShared(g *protogen.GeneratedFile, file *protogen.File) error {
	data := TemplateData{
		File: file,
	}

	// Generate minimal header for shared file
	var headerBuf bytes.Buffer
	if err := l.templates.ExecuteTemplate(&headerBuf, "shared_header.py.tmpl", data); err != nil {
		return fmt.Errorf("execute shared header template: %w", err)
	}
	g.P(headerBuf.String())
	g.P()

	// Generate shared types (error codes, interceptors, headers)
	var sharedBuf bytes.Buffer
	if err := l.templates.ExecuteTemplate(&sharedBuf, "shared.py.tmpl", data); err != nil {
		return fmt.Errorf("execute shared template: %w", err)
	}
	g.P(sharedBuf.String())
	g.P()

	return nil
}
