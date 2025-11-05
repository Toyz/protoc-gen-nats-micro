package generator

import (
	"bytes"
	"fmt"
	"text/template"

	"google.golang.org/protobuf/compiler/protogen"
)

// GoLanguage implements Language for Go code generation
type GoLanguage struct {
	templates *template.Template
}

// NewGoLanguage creates a new Go language generator
func NewGoLanguage() *GoLanguage {
	tmpl := template.Must(template.New("go").Funcs(FuncMap()).ParseFS(templatesFS, "templates/go/*.tmpl"))
	return &GoLanguage{
		templates: tmpl,
	}
}

func (l *GoLanguage) Name() string {
	return "go"
}

func (l *GoLanguage) FileExtension() string {
	return "_nats.pb.go"
}

func (l *GoLanguage) Generate(g *protogen.GeneratedFile, service *protogen.Service, opts ServiceOptions) error {
	data := TemplateData{
		Service: service,
		Options: opts,
	}

	// Generate service
	var serviceBuf bytes.Buffer
	if err := l.templates.ExecuteTemplate(&serviceBuf, "service.go.tmpl", data); err != nil {
		return fmt.Errorf("execute service template: %w", err)
	}
	g.P(serviceBuf.String())
	g.P()

	// Generate client
	var clientBuf bytes.Buffer
	if err := l.templates.ExecuteTemplate(&clientBuf, "client.go.tmpl", data); err != nil {
		return fmt.Errorf("execute client template: %w", err)
	}
	g.P(clientBuf.String())
	g.P()

	return nil
}

// GenerateHeader generates the file header (package declaration and imports)
func (l *GoLanguage) GenerateHeader(g *protogen.GeneratedFile, file *protogen.File) error {
	data := TemplateData{
		File: file,
	}

	var headerBuf bytes.Buffer
	if err := l.templates.ExecuteTemplate(&headerBuf, "header.go.tmpl", data); err != nil {
		return fmt.Errorf("execute header template: %w", err)
	}
	g.P(headerBuf.String())
	g.P()

	return nil
}
