package generator

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
)

// GenerateFile generates NATS microservice code for a protobuf file.
// The Language must be resolved by the caller (main.go).
func GenerateFile(gen *protogen.Plugin, file *protogen.File, lang Language) error {
	if len(file.Services) == 0 {
		return nil
	}

	// Only Go-like languages use Go import paths
	var importPath protogen.GoImportPath
	if lang.IsGoLike() {
		importPath = file.GoImportPath
	}

	// Go-like: use GeneratedFilenamePrefix (derived from go_package)
	// Others: use the proto source path (e.g., "auth/v1/auth.proto" -> "auth/v1/auth")
	filenamePrefix := file.GeneratedFilenamePrefix
	if !lang.IsGoLike() {
		filenamePrefix = strings.TrimSuffix(file.Proto.GetName(), ".proto")
	}
	filename := filenamePrefix + lang.FileExtension()
	g := gen.NewGeneratedFile(filename, importPath)

	// Generate header (package, imports)
	if err := lang.GenerateHeader(g, file); err != nil {
		return fmt.Errorf("generate header: %w", err)
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

// ToSnakeCase converts CamelCase to snake_case, handling acronyms correctly.
// e.g., "HTTPServer" -> "http_server", "getHTTPSURL" -> "get_https_url"
func ToSnakeCase(s string) string {
	var result strings.Builder
	runes := []rune(s)
	for i, r := range runes {
		if i > 0 {
			prev := runes[i-1]
			if r >= 'A' && r <= 'Z' {
				if prev >= 'a' && prev <= 'z' {
					// lowercase → uppercase: "getH" → "get_h"
					result.WriteByte('_')
				} else if prev >= '0' && prev <= '9' {
					// digit → uppercase: "V2O" → "v2_o"
					result.WriteByte('_')
				} else if prev >= 'A' && prev <= 'Z' && i+1 < len(runes) && runes[i+1] >= 'a' && runes[i+1] <= 'z' {
					// End of acronym before lowercase: "HTTPSe" → "http_se"
					result.WriteByte('_')
				}
			} else if r >= '0' && r <= '9' && prev >= 'a' && prev <= 'z' {
				// lowercase → digit: "Order2" → kept as-is (no separator needed, handled by context)
			}
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
