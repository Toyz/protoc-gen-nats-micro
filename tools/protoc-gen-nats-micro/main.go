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

		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}
			if err := generator.GenerateFile(gen, f, cfg); err != nil {
				return fmt.Errorf("generate file %s: %w", f.Desc.Path(), err)
			}
		}
		return nil
	})
}
