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
		langName := *language

		// Check for language in parameters (e.g., --nats-micro_opt=language=typescript)
		for _, param := range strings.Split(gen.Request.GetParameter(), ",") {
			if strings.HasPrefix(param, "language=") {
				langName = strings.TrimPrefix(param, "language=")
			} else if strings.HasPrefix(param, "lang=") {
				langName = strings.TrimPrefix(param, "lang=")
			}
		}

		// Resolve language once — used for all files
		lang, err := generator.GetLanguage(langName)
		if err != nil {
			return fmt.Errorf("get language: %w", err)
		}

		// Track which packages have had shared files generated
		generatedShared := make(map[string]bool)

		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}

			// For Go-like languages: Use GeneratedFilenamePrefix (derived from go_package)
			// For others: Use the proto source path (e.g., "auth/v1/auth.proto" -> "auth/v1")
			var filenameBase string
			if lang.IsGoLike() {
				filenameBase = f.GeneratedFilenamePrefix
			} else {
				filenameBase = strings.TrimSuffix(f.Proto.GetName(), ".proto")
			}
			pkgDir := filenameBase
			lastSlash := strings.LastIndex(pkgDir, "/")
			if lastSlash > 0 {
				pkgDir = pkgDir[:lastSlash]
			}

			// For Go-like: use the import path (e.g., "github.com/example/gen/order/v1")
			// For others: use the directory path (e.g., "gen/order/v1")
			pkgKey := string(f.GoImportPath)
			if !lang.IsGoLike() {
				pkgKey = pkgDir
			}

			if len(f.Services) > 0 && !generatedShared[pkgKey] {
				generatedShared[pkgKey] = true

				// Only Go-like languages use the Go import path for generated files
				var importPath protogen.GoImportPath
				if lang.IsGoLike() {
					importPath = f.GoImportPath
				}

				// Use the package directory + "/shared" for the filename
				sharedFilename := pkgDir + "/shared" + lang.FileExtension()
				sharedFile := gen.NewGeneratedFile(sharedFilename, importPath)

				// Generate shared content through the Language interface
				if err := lang.GenerateShared(sharedFile, f); err != nil {
					return fmt.Errorf("generate shared: %w", err)
				}

				// Allow language-specific post-generation (e.g., Python __init__.py)
				if err := lang.PostGenerate(gen, f, pkgDir); err != nil {
					return fmt.Errorf("post generate: %w", err)
				}
			}

			if err := generator.GenerateFile(gen, f, lang); err != nil {
				return fmt.Errorf("generate file %s: %w", f.Desc.Path(), err)
			}
		}
		return nil
	})
}
