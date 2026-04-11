# Fix Missing `"os"` Import — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix issue #4 — generated `service_nats.pb.go` references `os.Stderr` without importing `"os"`, and add a Go compilation validation test to prevent regressions.

**Architecture:** One-line template fix in `header.go.tmpl` moves the `"os"` import from conditional to unconditional. A new `TestGoGeneratedCodeCompiles` in `syntax_validation_test.go` validates generated Go code parses and has correct imports.

**Tech Stack:** Go templates, `go/parser`, `go/ast`, existing test helpers from `chunked_io_test.go`

---

## File Map

| File | Action | Purpose |
|------|--------|---------|
| `tools/protoc-gen-nats-micro/generator/templates/go/header.go.tmpl` | Modify | Move `"os"` from conditional to unconditional import |
| `tools/protoc-gen-nats-micro/generator/syntax_validation_test.go` | Modify | Add `TestGoGeneratedCodeCompiles` |

---

### Task 1: Write failing test — Go compilation validation

**Files:**
- Modify: `tools/protoc-gen-nats-micro/generator/syntax_validation_test.go`

- [ ] **Step 1: Add `TestGoGeneratedCodeCompiles` to `syntax_validation_test.go`**

Append this test after the existing `TestPythonGeneratedCodeSyntax` function. It uses the same test helpers (`buildTestFile`, `newTestPlugin`, `messageDescriptor`, `methodDescriptor`, `stringField`, `bytesField`) already defined in `chunked_io_test.go` (same package).

```go
// TestGoGeneratedCodeCompiles generates Go code that exercises every
// template code path (unary, server-streaming, client-streaming, bidi,
// chunked upload) and validates it parses as valid Go with correct imports.
func TestGoGeneratedCodeCompiles(t *testing.T) {
	// Build a comprehensive proto exercising every Go code path.
	file := buildTestFile(t, []*descriptorpb.DescriptorProto{
		messageDescriptor("GetRequest", stringField("id", 1)),
		messageDescriptor("GetResponse", stringField("value", 1)),
		messageDescriptor("ListRequest", stringField("filter", 1)),
		messageDescriptor("ListResponse", stringField("item", 1)),
		messageDescriptor("SumRequest", stringField("value", 1)),
		messageDescriptor("SumResponse", stringField("total", 1)),
		messageDescriptor("ChatRequest", stringField("message", 1)),
		messageDescriptor("ChatResponse", stringField("reply", 1)),
		messageDescriptor("SnapshotChunk", bytesField("data", 1)),
		messageDescriptor("UploadResponse", stringField("id", 1)),
	}, []*descriptorpb.MethodDescriptorProto{
		// Unary — the exact reproduction case from issue #4
		methodDescriptor("Get", "GetRequest", "GetResponse", false, false, nil),
		// Server-streaming
		methodDescriptor("ListItems", "ListRequest", "ListResponse", false, true, nil),
		// Client-streaming (plain)
		methodDescriptor("Sum", "SumRequest", "SumResponse", true, false, nil),
		// Bidi-streaming
		methodDescriptor("Chat", "ChatRequest", "ChatResponse", true, true, nil),
		// Client-streaming + chunked_io
		methodDescriptor("Upload", "SnapshotChunk", "UploadResponse", true, false, &natspb.ChunkedIOOptions{
			ChunkField:       "data",
			DefaultChunkSize: 65536,
		}),
	})

	gen, target := newTestPlugin(t, file)
	lang := NewGoLanguage()

	// Generate shared file (shared_header + shared templates)
	shared := gen.NewGeneratedFile("test/shared_nats.pb.go", "")
	if err := lang.GenerateShared(shared, target); err != nil {
		t.Fatalf("GenerateShared() error = %v", err)
	}

	// Generate the service file (header + service templates)
	if err := GenerateFile(gen, target, lang); err != nil {
		t.Fatalf("GenerateFile() error = %v", err)
	}

	// Collect and validate all generated .go files
	responseFiles := gen.Response().File
	if len(responseFiles) == 0 {
		t.Fatal("no files were generated")
	}

	fset := token.NewFileSet()
	var foundGo bool
	for _, f := range responseFiles {
		name := f.GetName()
		if !strings.HasSuffix(name, ".go") {
			continue
		}
		foundGo = true
		content := f.GetContent()

		t.Run(name, func(t *testing.T) {
			// Step 1: Syntax validation via go/parser
			parsed, err := parser.ParseFile(fset, name, content, parser.AllErrors|parser.ParseComments)
			if err != nil {
				t.Errorf("Go parse error in %q:\n%s\n\nGenerated content:\n%s", name, err, content)
				return
			}

			// Step 2: Import regression guard for issue #4
			// Service files (not shared_*) must import "os"
			if !strings.Contains(name, "shared") {
				hasOsImport := false
				for _, imp := range parsed.Imports {
					if imp.Path.Value == `"os"` {
						hasOsImport = true
						break
					}
				}
				if !hasOsImport {
					t.Errorf("service file %q is missing \"os\" import (regression: issue #4).\nImports found: %v",
						name, formatImports(parsed.Imports))
				}
			}
		})
	}
	if !foundGo {
		t.Fatal("no Go files found in generated output")
	}
}

// formatImports returns a human-readable list of import paths from AST import specs.
func formatImports(imports []*ast.ImportSpec) []string {
	paths := make([]string, len(imports))
	for i, imp := range imports {
		paths[i] = imp.Path.Value
	}
	return paths
}
```

- [ ] **Step 2: Add required imports to `syntax_validation_test.go`**

Add `"go/ast"`, `"go/parser"`, and `"go/token"` to the import block at the top of the file. The existing imports are:

```go
import (
	"os/exec"
	"strings"
	"testing"

	natspb "github.com/franchb/protoc-gen-nats-micro/tools/protoc-gen-nats-micro/nats/micro"
	"google.golang.org/protobuf/types/descriptorpb"
)
```

Update to:

```go
import (
	"go/ast"
	"go/parser"
	"go/token"
	"os/exec"
	"strings"
	"testing"

	natspb "github.com/franchb/protoc-gen-nats-micro/tools/protoc-gen-nats-micro/nats/micro"
	"google.golang.org/protobuf/types/descriptorpb"
)
```

- [ ] **Step 3: Run the test to verify it fails**

Run:
```bash
cd tools/protoc-gen-nats-micro && go test ./generator/ -run TestGoGeneratedCodeCompiles -v
```

Expected: FAIL — the test should report that the service file is missing the `"os"` import, because `header.go.tmpl` currently only imports `"os"` conditionally for chunked-send methods.

The unary-only service file won't have `"os"` in its imports, triggering the regression guard.

- [ ] **Step 4: Commit the failing test**

```bash
git add tools/protoc-gen-nats-micro/generator/syntax_validation_test.go
git commit -m "test: add Go compilation validation for generated code (issue #4)

Validates generated Go code parses correctly and checks that service
files include the \"os\" import. Currently fails — the fix is in the
next commit."
```

---

### Task 2: Fix the template — make `"os"` unconditional

**Files:**
- Modify: `tools/protoc-gen-nats-micro/generator/templates/go/header.go.tmpl`

- [ ] **Step 1: Edit `header.go.tmpl` to move `"os"` to unconditional imports**

Current content (lines 23-33):

```
import (
  "context"
  "errors"
  "fmt"
{{- if or $needsChunkedRecvImports $needsChunkedSendImports}}
  "io"
{{- end}}
{{- if $needsChunkedSendImports}}
  "os"
{{- end}}
```

Change to:

```
import (
  "context"
  "errors"
  "fmt"
{{- if or $needsChunkedRecvImports $needsChunkedSendImports}}
  "io"
{{- end}}
  "os"
```

This removes the `{{- if $needsChunkedSendImports}}` guard around `"os"`, making it unconditional. The `"os"` import is always needed because every service file uses `os.Stderr` in at least the response error path.

- [ ] **Step 2: Run the test to verify it passes**

Run:
```bash
cd tools/protoc-gen-nats-micro && go test ./generator/ -run TestGoGeneratedCodeCompiles -v
```

Expected: PASS — the generated service file now includes `"os"` unconditionally.

- [ ] **Step 3: Run the full test suite to check for regressions**

Run:
```bash
cd tools/protoc-gen-nats-micro && go test ./generator/ -v
```

Expected: All tests pass, including the existing `TestPythonGeneratedCodeSyntax`, `TestToSnakeCase`, etc.

- [ ] **Step 4: Commit the fix**

```bash
git add tools/protoc-gen-nats-micro/generator/templates/go/header.go.tmpl
git commit -m "fix: always import \"os\" in generated service files (issue #4)

The service template uses fmt.Fprintf(os.Stderr, ...) in the response
error path for every unary method, plus KV/ObjectStore warning paths
and client-streaming error paths. The \"os\" import was previously
conditional on chunked-send methods only, causing compilation errors
for services without chunked IO.

Fixes #4"
```
