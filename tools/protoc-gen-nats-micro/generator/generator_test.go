package generator

import "testing"

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Basic cases
		{"", ""},
		{"a", "a"},
		{"A", "a"},

		// Simple CamelCase
		{"FooBar", "foo_bar"},
		{"fooBar", "foo_bar"},
		{"CreateProduct", "create_product"},
		{"GetOrder", "get_order"},

		// Acronyms
		{"HTTPServer", "http_server"},
		{"APIGateway", "api_gateway"},
		{"DBService", "db_service"},
		{"getHTTPSURL", "get_httpsurl"}, // Adjacent acronyms are ambiguous without a dictionary
		{"XMLParser", "xml_parser"},
		{"parseJSON", "parse_json"},
		{"IOReader", "io_reader"},

		// Already snake_case
		{"foo_bar", "foo_bar"},
		{"already_snake", "already_snake"},

		// Single word
		{"Product", "product"},
		{"order", "order"},

		// Numbers (should pass through)
		{"V2Order", "v2_order"},
		{"OrderV2", "order_v2"},

		// Consecutive uppercase at end
		{"MyAPI", "my_api"},
		{"TestHTTP", "test_http"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ToSnakeCase(tt.input)
			if got != tt.expected {
				t.Errorf("ToSnakeCase(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestToLowerFirst(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"A", "a"},
		{"Hello", "hello"},
		{"helloWorld", "helloWorld"},
		{"HTTPServer", "hTTPServer"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ToLowerFirst(tt.input)
			if got != tt.expected {
				t.Errorf("ToLowerFirst(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestToUpperFirst(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"a", "A"},
		{"hello", "Hello"},
		{"Hello", "Hello"},
		{"hTTPServer", "HTTPServer"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ToUpperFirst(tt.input)
			if got != tt.expected {
				t.Errorf("ToUpperFirst(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"foo_bar", "FooBar"},
		{"create_product", "CreateProduct"},
		{"a", "A"},
		{"hello_world_test", "HelloWorldTest"},
		{"already", "Already"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ToCamelCase(tt.input)
			if got != tt.expected {
				t.Errorf("ToCamelCase(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestToKebabCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"FooBar", "foo-bar"},
		{"CreateProduct", "create-product"},
		{"HTTPServer", "http-server"},
		{"APIGateway", "api-gateway"},
		{"parseJSON", "parse-json"},
		{"Product", "product"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ToKebabCase(tt.input)
			if got != tt.expected {
				t.Errorf("ToKebabCase(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestProtoBasename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"service.proto", "service"},
		{"path/to/service.proto", "service"},
		{"order/v1/order.proto", "order"},
		{"simple", "simple"},
		{"a.proto", "a"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ProtoBasename(tt.input)
			if got != tt.expected {
				t.Errorf("ProtoBasename(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestGetLanguage(t *testing.T) {
	// Valid languages
	validCases := []struct {
		input        string
		expectedName string
	}{
		{"go", "go"},
		{"golang", "go"},
		{"typescript", "typescript"},
		{"ts", "typescript"},
		{"python", "python"},
		{"py", "python"},
		{"web-ts", "web-ts"},
		{"webts", "web-ts"},
	}

	for _, tt := range validCases {
		t.Run(tt.input, func(t *testing.T) {
			lang, err := GetLanguage(tt.input)
			if err != nil {
				t.Fatalf("GetLanguage(%q) returned unexpected error: %v", tt.input, err)
			}
			if lang.Name() != tt.expectedName {
				t.Errorf("GetLanguage(%q).Name() = %q, want %q", tt.input, lang.Name(), tt.expectedName)
			}
		})
	}

	// Invalid languages
	_, err := GetLanguage("rust")
	if err == nil {
		t.Error("GetLanguage(\"rust\") should return error for unsupported language")
	}

	_, err = GetLanguage("java")
	if err == nil {
		t.Error("GetLanguage(\"java\") should return error for unsupported language")
	}
}
