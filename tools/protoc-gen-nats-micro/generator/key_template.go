package generator

import (
	"fmt"
	"regexp"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
)

var keyTemplatePlaceholderRe = regexp.MustCompile(`\{(\w+)\}`)

// ValidateKeyTemplate checks that every {field} placeholder in the template
// refers to an actual field on the method's input message. Returns an error
// with a clear message listing available fields if a placeholder is invalid.
func ValidateKeyTemplate(template string, method *protogen.Method) error {
	matches := keyTemplatePlaceholderRe.FindAllStringSubmatch(template, -1)
	if len(matches) == 0 {
		return nil // No placeholders, nothing to validate
	}

	// Build a set of valid field names from the input message
	validFields := make(map[string]bool)
	var fieldNames []string
	for _, f := range method.Input.Fields {
		name := string(f.Desc.Name())
		validFields[name] = true
		fieldNames = append(fieldNames, name)
	}

	// Check each placeholder
	for _, m := range matches {
		fieldName := m[1]
		if !validFields[fieldName] {
			return fmt.Errorf(
				"key_template %q references field {%s} which does not exist on input message %s (available fields: [%s])",
				template,
				fieldName,
				method.Input.GoIdent.GoName,
				strings.Join(fieldNames, ", "),
			)
		}
	}
	return nil
}

// ResolveKeyTemplateGo converts a key template like "user.{id}" into Go code:
// fmt.Sprintf("user.%v", msg.GetId())
// Panics at code-gen time if a placeholder references a nonexistent field.
func ResolveKeyTemplateGo(template string, method *protogen.Method) string {
	if err := ValidateKeyTemplate(template, method); err != nil {
		panic(fmt.Sprintf("protoc-gen-nats-micro: %v", err))
	}

	matches := keyTemplatePlaceholderRe.FindAllStringSubmatch(template, -1)
	if len(matches) == 0 {
		return fmt.Sprintf("%q", template)
	}

	format := keyTemplatePlaceholderRe.ReplaceAllString(template, "%v")
	var args []string
	for _, m := range matches {
		fieldName := m[1]
		goFieldName := fieldNameToGoGetter(fieldName)
		args = append(args, fmt.Sprintf("msg.Get%s()", goFieldName))
	}

	return fmt.Sprintf("fmt.Sprintf(%q, %s)", format, strings.Join(args, ", "))
}

// ResolveKeyTemplateTS converts a key template like "user.{id}" into TypeScript code:
// `user.${req.id}`
// Panics at code-gen time if a placeholder references a nonexistent field.
func ResolveKeyTemplateTS(template string, method *protogen.Method) string {
	if err := ValidateKeyTemplate(template, method); err != nil {
		panic(fmt.Sprintf("protoc-gen-nats-micro: %v", err))
	}

	result := keyTemplatePlaceholderRe.ReplaceAllStringFunc(template, func(match string) string {
		fieldName := match[1 : len(match)-1] // strip { }
		tsFieldName := fieldNameToTSAccessor(fieldName)
		return fmt.Sprintf("${req.%s}", tsFieldName)
	})
	return fmt.Sprintf("`%s`", result)
}

// ResolveKeyTemplatePy converts a key template like "user.{id}" into Python code:
// f"user.{request_msg.id}"
// Panics at code-gen time if a placeholder references a nonexistent field.
func ResolveKeyTemplatePy(template string, method *protogen.Method) string {
	if err := ValidateKeyTemplate(template, method); err != nil {
		panic(fmt.Sprintf("protoc-gen-nats-micro: %v", err))
	}

	result := keyTemplatePlaceholderRe.ReplaceAllStringFunc(template, func(match string) string {
		fieldName := match[1 : len(match)-1] // strip { }
		return fmt.Sprintf("{request_msg.%s}", fieldName)
	})
	return fmt.Sprintf("f\"%s\"", result)
}

// GetInputFields returns a list of field names from the method's input message type
func GetInputFields(method *protogen.Method) []string {
	var fields []string
	for _, f := range method.Input.Fields {
		fields = append(fields, string(f.Desc.Name()))
	}
	return fields
}

// fieldNameToGoGetter converts a proto field name (snake_case) to a Go getter name
// e.g., "user_id" -> "UserId", "id" -> "Id"
func fieldNameToGoGetter(name string) string {
	parts := strings.Split(name, "_")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, "")
}

// fieldNameToTSAccessor converts a proto field name to a TypeScript accessor
// Proto uses snake_case, TS/JS generated code uses camelCase
// e.g., "user_id" -> "userId", "id" -> "id"
func fieldNameToTSAccessor(name string) string {
	parts := strings.Split(name, "_")
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}
