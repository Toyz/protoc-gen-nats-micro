package generator

import (
	"embed"
	"fmt"
	"strings"
	"text/template"

	"google.golang.org/protobuf/compiler/protogen"
)

//go:embed templates/*
var templatesFS embed.FS

// Language represents a target programming language for code generation
type Language interface {
	// Name returns the language name (e.g., "go", "rust")
	Name() string

	// FileExtension returns the file extension (e.g., ".go", ".rs")
	FileExtension() string

	// GenerateShared generates shared code once per proto file (e.g., RegisterOption types)
	GenerateShared(g *protogen.GeneratedFile, file *protogen.File) error

	// Generate generates code for the given service
	Generate(g *protogen.GeneratedFile, service *protogen.Service, opts ServiceOptions) error
}

// TemplateData holds data passed to templates
type TemplateData struct {
	File    *protogen.File
	Service *protogen.Service
	Options ServiceOptions
}

// FuncMap returns template helper functions
func FuncMap() template.FuncMap {
	return template.FuncMap{
		"ToSnakeCase":        ToSnakeCase,
		"ToLowerFirst":       ToLowerFirst,
		"ToUpperFirst":       ToUpperFirst,
		"ToCamelCase":        ToCamelCase,
		"ToKebabCase":        ToKebabCase,
		"GetEndpointOptions": GetEndpointOptions,
	}
}

// ToUpperFirst converts first character to uppercase
func ToUpperFirst(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// ToCamelCase converts snake_case to CamelCase
func ToCamelCase(s string) string {
	parts := strings.Split(s, "_")
	for i, part := range parts {
		parts[i] = ToUpperFirst(part)
	}
	return strings.Join(parts, "")
}

// ToKebabCase converts CamelCase to kebab-case
func ToKebabCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteByte('-')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// GetLanguage returns a language generator by name
func GetLanguage(name string) (Language, error) {
	switch strings.ToLower(name) {
	case "go", "golang":
		return NewGoLanguage(), nil
	case "typescript", "ts":
		return NewTypeScriptLanguage(), nil
	// Future languages:
	// case "rust":
	//   return NewRustLanguage(), nil
	default:
		return nil, fmt.Errorf("unsupported language: %s", name)
	}
}
