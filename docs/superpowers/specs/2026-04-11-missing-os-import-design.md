# Fix: Missing `"os"` import in generated service files

**Issue:** [franchb/protoc-gen-nats-micro#4](https://github.com/franchb/protoc-gen-nats-micro/issues/4)
**Date:** 2026-04-11

## Problem

`protoc-gen-nats-micro` generates `service_nats.pb.go` files that reference `os.Stderr` but do not include `"os"` in the import block. The `"os"` import is only present in `shared_nats.pb.go`, which is a separate file — Go imports are per-file, not per-package.

This causes a compilation error:

```
service_nats.pb.go:471:16: undefined: os
```

### Root cause

`header.go.tmpl` (the service file header) only imports `"os"` conditionally — when `$needsChunkedSendImports` is true. But `service.go.tmpl` uses `fmt.Fprintf(os.Stderr, ...)` unconditionally in the response-send error path for every unary method (lines 512, 517), and additionally in KV/ObjectStore warning paths and client-streaming error paths. There are 10 `os.Stderr` references in the service template.

Any proto with at least one service method triggers the bug if it doesn't also have client-streaming + chunked_io methods (which is the only condition that currently enables the `"os"` import).

## Fix

### Template change

In `tools/protoc-gen-nats-micro/generator/templates/go/header.go.tmpl`, add `"os"` to the unconditional imports block:

```diff
 import (
   "context"
   "errors"
   "fmt"
+  "os"
 {{- if or $needsChunkedRecvImports $needsChunkedSendImports}}
   "io"
 {{- end}}
-{{- if $needsChunkedSendImports}}
-  "os"
-{{- end}}
```

The `"os"` import moves from conditional (chunked-send-only) to unconditional because every generated service file uses `os.Stderr` in at least the response error path. This matches `shared_header.go.tmpl` which already imports `"os"` unconditionally.

## Test: Go compilation validation

### Purpose

Prevent regressions where generated Go code fails to compile due to missing imports or syntax errors. Modeled after the existing `TestPythonGeneratedCodeSyntax`.

### Location

`tools/protoc-gen-nats-micro/generator/syntax_validation_test.go` — alongside the existing Python syntax validation test.

### Design

`TestGoGeneratedCodeCompiles` will:

1. Use existing test helpers (`buildTestFile`, `newTestPlugin`, `messageDescriptor`, `methodDescriptor`) to build a proto exercising all Go code paths
2. Run the generator (`GenerateFile` + `GenerateShared`)
3. Validate every generated `.go` file parses with `go/parser.ParseFile`
4. Check that service files include `"os"` in their imports (regression guard for issue #4)

### Test scenarios

The proto will include:

- **Plain unary RPC** — the exact reproduction case from issue #4
- **Server-streaming RPC** — exercises the server-streaming handler template
- **Client-streaming RPC** — exercises client-streaming handler (uses `os.Stderr` for error logging)
- **Bidi-streaming RPC** — exercises bidirectional handler
- **Client-streaming + chunked_io** — exercises the chunked send import path (previously the only path that imported `"os"`)

All methods will be in a single service to match realistic usage and exercise the combined template output.

### Validation steps

For each generated `.go` file:
1. **Syntax validation:** Parse with `go/parser.ParseFile(fset, name, content, parser.AllErrors)` — catches template rendering bugs (unclosed braces, invalid Go syntax, malformed declarations)
2. **Import regression guard:** Walk the parsed AST's import declarations and verify that expected packages are present. For service files (not `shared_*`), assert `"os"` is in the imports — direct regression guard for issue #4

Note: `go/parser` validates syntax only, not semantics. It will NOT catch "undefined: os" errors. Step 2 compensates by explicitly checking import declarations via AST inspection.

### No external dependencies

Unlike a full `go/types` check, `go/parser` + AST inspection doesn't require resolving external packages (`nats.go`, `protobuf`, etc.). It validates syntax and import structure, which catches both template rendering bugs and the specific missing-import class of bug reported in this issue. This matches the Python test's approach of using `ast.parse()` for syntax validation.
