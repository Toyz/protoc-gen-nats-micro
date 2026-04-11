package generator

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os/exec"
	"strings"
	"testing"

	natspb "github.com/franchb/protoc-gen-nats-micro/tools/protoc-gen-nats-micro/nats/micro"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/descriptorpb"
)

// findPython3 locates a Python 3 interpreter on PATH.
// Tries "python3" first, then falls back to "python" if it reports version 3.x.
func findPython3(t *testing.T) (string, bool) {
	t.Helper()
	if bin, err := exec.LookPath("python3"); err == nil {
		return bin, true
	}
	bin, err := exec.LookPath("python")
	if err != nil {
		return "", false
	}
	out, err := exec.Command(bin, "--version").CombinedOutput()
	if err != nil || !strings.HasPrefix(string(out), "Python 3") {
		return "", false
	}
	return bin, true
}

// assertPythonSyntax validates that code is syntactically correct Python by
// running ast.parse() via the given interpreter. Code is passed through stdin
// to avoid shell-escaping issues with large generated output.
func assertPythonSyntax(t *testing.T, pythonBin, filename, code string) {
	t.Helper()
	cmd := exec.Command(pythonBin, "-c", "import ast,sys; ast.parse(sys.stdin.read())")
	cmd.Stdin = strings.NewReader(code)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("Python syntax error in generated file %q:\n%s", filename, string(out))
	}
}

// TestPythonGeneratedCodeSyntax generates Python code that exercises every
// template code path (unary, server-streaming, client-streaming, chunked
// download, chunked upload) and validates it as syntactically correct Python.
func TestPythonGeneratedCodeSyntax(t *testing.T) {
	pythonBin, found := findPython3(t)
	if !found {
		t.Skip("python3 not found on PATH; skipping Python syntax validation")
	}

	// Build a comprehensive proto that exercises every Python code path.
	file := buildTestFile(t, []*descriptorpb.DescriptorProto{
		messageDescriptor("GetRequest", stringField("id", 1)),
		messageDescriptor("GetResponse", stringField("value", 1)),
		messageDescriptor("ListRequest", stringField("filter", 1)),
		messageDescriptor("ListResponse", stringField("item", 1)),
		messageDescriptor("SumRequest", stringField("value", 1)),
		messageDescriptor("SumResponse", stringField("total", 1)),
		messageDescriptor("SnapshotChunk", bytesField("data", 1)),
		messageDescriptor("DownloadRequest", stringField("id", 1)),
		messageDescriptor("UploadResponse", stringField("id", 1)),
	}, []*descriptorpb.MethodDescriptorProto{
		// Unary
		methodDescriptor("Get", "GetRequest", "GetResponse", false, false, nil),
		// Server-streaming (plain)
		methodDescriptor("ListItems", "ListRequest", "ListResponse", false, true, nil),
		// Client-streaming (plain)
		methodDescriptor("Sum", "SumRequest", "SumResponse", true, false, nil),
		// Server-streaming + chunked_io
		methodDescriptor("Download", "DownloadRequest", "SnapshotChunk", false, true, &natspb.ChunkedIOOptions{
			ChunkField:       "data",
			DefaultChunkSize: 65536,
		}),
		// Client-streaming + chunked_io
		methodDescriptor("Upload", "SnapshotChunk", "UploadResponse", true, false, &natspb.ChunkedIOOptions{
			ChunkField:       "data",
			DefaultChunkSize: 65536,
		}),
	})

	gen, target := newTestPlugin(t, file)
	lang := NewPythonLanguage()

	shared := gen.NewGeneratedFile("test/shared_nats_pb2.py", "")
	if err := lang.GenerateShared(shared, target); err != nil {
		t.Fatalf("GenerateShared() error = %v", err)
	}

	if err := GenerateFile(gen, target, lang); err != nil {
		t.Fatalf("GenerateFile() error = %v", err)
	}

	// Collect all generated Python files and validate syntax.
	responseFiles := gen.Response().File
	if len(responseFiles) == 0 {
		t.Fatal("no files were generated")
	}

	var found_py bool
	for _, f := range responseFiles {
		if strings.HasSuffix(f.GetName(), ".py") {
			found_py = true
			name := f.GetName()
			content := f.GetContent()
			t.Run(name, func(t *testing.T) {
				assertPythonSyntax(t, pythonBin, name, content)
			})
		}
	}
	if !found_py {
		t.Fatal("no Python files found in generated output")
	}
}

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
		// Unary
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
	shared := gen.NewGeneratedFile("test/shared_nats.pb.go", target.GoImportPath)
	if err := lang.GenerateShared(shared, target); err != nil {
		t.Fatalf("GenerateShared() error = %v", err)
	}

	// Generate the service file (header + service templates)
	if err := GenerateFile(gen, target, lang); err != nil {
		t.Fatalf("GenerateFile() error = %v", err)
	}

	// Collect and validate all generated .go files.
	// Skip "os" check for chunked files — they only contain SendBytes
	// which doesn't use os.
	validateGoFiles(t, gen, true)
}

// TestGoGeneratedCodeCompiles_UnaryOnly reproduces issue #4 exactly:
// a proto with only unary methods generates service_nats.pb.go that
// references os.Stderr but does not import "os".
func TestGoGeneratedCodeCompiles_UnaryOnly(t *testing.T) {
	file := buildTestFile(t, []*descriptorpb.DescriptorProto{
		messageDescriptor("GetItemRequest", stringField("id", 1)),
		messageDescriptor("GetItemResponse", bytesField("data", 1)),
	}, []*descriptorpb.MethodDescriptorProto{
		methodDescriptor("GetItem", "GetItemRequest", "GetItemResponse", false, false, nil),
	})

	gen, target := newTestPlugin(t, file)
	lang := NewGoLanguage()

	shared := gen.NewGeneratedFile("test/shared_nats.pb.go", target.GoImportPath)
	if err := lang.GenerateShared(shared, target); err != nil {
		t.Fatalf("GenerateShared() error = %v", err)
	}

	if err := GenerateFile(gen, target, lang); err != nil {
		t.Fatalf("GenerateFile() error = %v", err)
	}

	validateGoFiles(t, gen, false)
}

// validateGoFiles checks all generated .go files parse correctly and
// that service files include the "os" import (regression guard for issue #4).
// Set skipChunked to true to exclude chunked send files from the "os" check.
func validateGoFiles(t *testing.T, gen *protogen.Plugin, skipChunked bool) {
	t.Helper()

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
			// Syntax validation via go/parser
			parsed, err := parser.ParseFile(fset, name, content, parser.AllErrors|parser.ParseComments)
			if err != nil {
				t.Errorf("Go parse error in %q:\n%s\n\nGenerated content:\n%s", name, err, content)
				return
			}

			// Import regression guard for issue #4
			// Service files (not shared_*) must import "os"
			if strings.Contains(name, "shared") {
				return
			}
			if skipChunked && strings.Contains(name, "chunked") {
				return
			}
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
