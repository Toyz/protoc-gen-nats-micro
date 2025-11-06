package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/toyz/protoc-gen-nats-micro/tools/protoc-gen-nats-micro/generator"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

const version = "0.3.0"

func main() {
	showVersion := flag.Bool("version", false, "print the version and exit")
	language := flag.String("lang", "go", "target language (go, rust, etc.)")
	flag.Parse()
	if *showVersion {
		fmt.Printf("protoc-gen-nats-micro %v\n", version)
		return
	}

	protogen.Options{}.Run(func(gen *protogen.Plugin) error {
		gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)

		// Parse language from plugin parameters
		cfg := generator.Config{
			Language: *language,
		}

		// Check for language in parameters (e.g., --nats-micro_opt=language=typescript)
		for _, param := range strings.Split(gen.Request.GetParameter(), ",") {
			if strings.HasPrefix(param, "language=") {
				cfg.Language = strings.TrimPrefix(param, "language=")
			} else if strings.HasPrefix(param, "lang=") {
				cfg.Language = strings.TrimPrefix(param, "lang=")
			}
		}

		// Track which packages have had shared files generated
		generatedShared := make(map[string]bool)

		// Get language generator
		lang, err := generator.GetLanguage(cfg.Language)
		if err != nil {
			return fmt.Errorf("get language: %w", err)
		}

		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}
			
			// Determine package key for shared file tracking
			// For Go: use the import path (e.g., "github.com/example/gen/order/v1")
			// For TypeScript: use the directory path (e.g., "gen/order/v1")
			pkgKey := string(f.GoImportPath)
			pkgDir := f.GeneratedFilenamePrefix
			
			if cfg.Language == "typescript" || cfg.Language == "ts" {
				// For TypeScript, extract directory from the filename prefix
				lastSlash := strings.LastIndex(pkgDir, "/")
				if lastSlash > 0 {
					pkgKey = pkgDir[:lastSlash]
				} else {
					pkgKey = "."
				}
			} else {
				// For Go, also extract the directory for the shared filename
				lastSlash := strings.LastIndex(pkgDir, "/")
				if lastSlash > 0 {
					pkgDir = pkgDir[:lastSlash]
				}
			}
			
			// Generate shared file once per package
			if !generatedShared[pkgKey] {
				generatedShared[pkgKey] = true
				
				// For non-Go languages, don't use Go import path
				var importPath protogen.GoImportPath
				if cfg.Language == "go" || cfg.Language == "golang" {
					importPath = f.GoImportPath
				}
				
				// Use the package directory + "/shared" for the filename
				sharedFilename := pkgDir + "/shared" + lang.FileExtension()
				sharedFile := gen.NewGeneratedFile(sharedFilename, importPath)
				
				// Generate shared content
				if goLang, ok := lang.(*generator.GoLanguage); ok {
					if err := goLang.GenerateShared(sharedFile, f); err != nil {
						return fmt.Errorf("generate shared: %w", err)
					}
				} else if tsLang, ok := lang.(*generator.TypeScriptLanguage); ok {
					if err := tsLang.GenerateShared(sharedFile, f); err != nil {
						return fmt.Errorf("generate shared: %w", err)
					}
				}
			}
			
			if err := generator.GenerateFile(gen, f, cfg); err != nil {
				return fmt.Errorf("generate file %s: %w", f.Desc.Path(), err)
			}
		}
		return nil
	})
}
