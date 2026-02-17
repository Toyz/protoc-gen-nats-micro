package generator

// GoLanguage implements Language for Go code generation
type GoLanguage struct{ BaseLanguage }

// IsGoLike returns true — Go uses Go import paths and GeneratedFilenamePrefix.
func (g *GoLanguage) IsGoLike() bool { return true }

// NewGoLanguage creates a new Go language generator
func NewGoLanguage() *GoLanguage {
	return &GoLanguage{newBaseLanguage("go", "_nats.pb.go", "templates/go/*.tmpl",
		[]string{"header.go.tmpl"},
		[]string{"shared_header.go.tmpl", "shared.go.tmpl", "stream_helpers.go.tmpl"},
		[]string{"errors.go.tmpl", "service.go.tmpl", "stream.go.tmpl", "client.go.tmpl"},
	)}
}
