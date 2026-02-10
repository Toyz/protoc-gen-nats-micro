package generator

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
)

// Config holds generator configuration
type Config struct {
	Language string // Target language (default: "go")
}

// GenerateFile generates NATS microservice code for a protobuf file
func GenerateFile(gen *protogen.Plugin, file *protogen.File, cfg Config) error {
	if len(file.Services) == 0 {
		return nil
	}

	// Default to Go if not specified
	if cfg.Language == "" {
		cfg.Language = "go"
	}

	// Get language generator
	lang, err := GetLanguage(cfg.Language)
	if err != nil {
		return fmt.Errorf("get language: %w", err)
	}

	// For non-Go languages, don't use Go import path
	isGo := cfg.Language == "go" || cfg.Language == "golang"
	var importPath protogen.GoImportPath
	if isGo {
		importPath = file.GoImportPath
	}

	// Generate main file with all services
	// For Go, use GeneratedFilenamePrefix (derived from go_package)
	// For non-Go, use the proto source path (e.g., "auth/v1/auth.proto" -> "auth/v1/auth")
	filenamePrefix := file.GeneratedFilenamePrefix
	if !isGo {
		filenamePrefix = strings.TrimSuffix(file.Proto.GetName(), ".proto")
	}
	filename := filenamePrefix + lang.FileExtension()
	g := gen.NewGeneratedFile(filename, importPath)

	// Generate header (package, imports)
	if goLang, ok := lang.(*GoLanguage); ok {
		if err := goLang.GenerateHeader(g, file); err != nil {
			return fmt.Errorf("generate header: %w", err)
		}
	} else if tsLang, ok := lang.(*TypeScriptLanguage); ok {
		if err := tsLang.GenerateHeader(g, file); err != nil {
			return fmt.Errorf("generate header: %w", err)
		}
	} else if pyLang, ok := lang.(*PythonLanguage); ok {
		if err := pyLang.GenerateHeader(g, file); err != nil {
			return fmt.Errorf("generate header: %w", err)
		}
	} else if webTsLang, ok := lang.(*WebTSLanguage); ok {
		if err := webTsLang.GenerateHeader(g, file); err != nil {
			return fmt.Errorf("generate header: %w", err)
		}
	}

	// Generate each service
	for _, service := range file.Services {
		opts := GetServiceOptions(service)

		// Skip this service if skip is set to true
		if opts.Skip {
			continue
		}

		if err := lang.Generate(g, file, service, opts); err != nil {
			return fmt.Errorf("generate service %s: %w", service.GoName, err)
		}
	}

	return nil
}

// ToSnakeCase converts CamelCase to snake_case
func ToSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteByte('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// ToLowerFirst converts first character to lowercase
func ToLowerFirst(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToLower(s[:1]) + s[1:]
}
