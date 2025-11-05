package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type OpenAPISpec struct {
	Swagger     string                 `yaml:"swagger" json:"swagger"`
	Info        map[string]interface{} `yaml:"info" json:"info"`
	Tags        []interface{}          `yaml:"tags,omitempty" json:"tags,omitempty"`
	Paths       map[string]interface{} `yaml:"paths" json:"paths"`
	Definitions map[string]interface{} `yaml:"definitions,omitempty" json:"definitions,omitempty"`
	Consumes    []string               `yaml:"consumes,omitempty" json:"consumes,omitempty"`
	Produces    []string               `yaml:"produces,omitempty" json:"produces,omitempty"`
}

func main() {
	inputDir := flag.String("input", "gen", "Input directory containing swagger files")
	outputFile := flag.String("output", "api.swagger.json", "Output merged file")
	format := flag.String("format", "json", "Output format: json or yaml")
	flag.Parse()

	merged := &OpenAPISpec{
		Swagger: "2.0",
		Info: map[string]interface{}{
			"title":       "Microservices API",
			"version":     "1.0.0",
			"description": "Combined API documentation for all microservices",
		},
		Paths:       make(map[string]interface{}),
		Definitions: make(map[string]interface{}),
		Tags:        make([]interface{}, 0),
		Consumes:    []string{"application/json"},
		Produces:    []string{"application/json"},
	}

	serviceFiles := []string{
		"order/v1/service.swagger.yaml",
		"order/v2/service.swagger.yaml",
		"product/v1/service.swagger.yaml",
		"user/v1/service.swagger.yaml",
	}

	for _, file := range serviceFiles {
		fullPath := filepath.Join(*inputDir, file)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			log.Printf("Skipping %s (not found)", file)
			continue
		}

		log.Printf("Merging %s", file)
		if err := mergeSpec(merged, fullPath); err != nil {
			log.Printf("Warning: Failed to merge %s: %v", file, err)
			continue
		}
	}

	// Write output
	var data []byte
	var err error
	if *format == "yaml" {
		data, err = yaml.Marshal(merged)
	} else {
		data, err = json.MarshalIndent(merged, "", "  ")
	}
	if err != nil {
		log.Fatal(err)
	}

	if err := os.WriteFile(*outputFile, data, 0644); err != nil {
		log.Fatal(err)
	}

	log.Printf("✓ Merged OpenAPI spec written to %s", *outputFile)
	log.Printf("  - %d paths", len(merged.Paths))
	log.Printf("  - %d definitions", len(merged.Definitions))
	log.Printf("  - %d tags", len(merged.Tags))
}

func mergeSpec(merged *OpenAPISpec, filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	var spec OpenAPISpec
	if strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".yml") {
		err = yaml.Unmarshal(data, &spec)
	} else {
		err = json.Unmarshal(data, &spec)
	}
	if err != nil {
		return err
	}

	// Merge paths
	for path, ops := range spec.Paths {
		merged.Paths[path] = ops
	}

	// Merge definitions
	for name, def := range spec.Definitions {
		merged.Definitions[name] = def
	}

	// Merge tags (deduplicate)
	for _, tag := range spec.Tags {
		if !containsTag(merged.Tags, tag) {
			merged.Tags = append(merged.Tags, tag)
		}
	}

	return nil
}

func containsTag(tags []interface{}, newTag interface{}) bool {
	newMap, ok := newTag.(map[string]interface{})
	if !ok {
		return false
	}
	newName, ok := newMap["name"].(string)
	if !ok {
		return false
	}

	for _, tag := range tags {
		if tagMap, ok := tag.(map[string]interface{}); ok {
			if name, ok := tagMap["name"].(string); ok && name == newName {
				return true
			}
		}
	}
	return false
}
