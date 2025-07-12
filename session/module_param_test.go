package session

import (
	"regexp"
	"strings"
	"testing"
)

func TestNewModuleParameter(t *testing.T) {
	tests := []struct {
		name      string
		paramName string
		defValue  string
		paramType ParamType
		validator string
		desc      string
	}{
		{
			name:      "string parameter with validator",
			paramName: "test.param",
			defValue:  "default",
			paramType: STRING,
			validator: "^[a-z]+$",
			desc:      "A test parameter",
		},
		{
			name:      "int parameter without validator",
			paramName: "test.int",
			defValue:  "42",
			paramType: INT,
			validator: "",
			desc:      "An integer parameter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewModuleParameter(tt.paramName, tt.defValue, tt.paramType, tt.validator, tt.desc)

			if p.Name != tt.paramName {
				t.Errorf("expected name %s, got %s", tt.paramName, p.Name)
			}
			if p.Value != tt.defValue {
				t.Errorf("expected value %s, got %s", tt.defValue, p.Value)
			}
			if p.Type != tt.paramType {
				t.Errorf("expected type %v, got %v", tt.paramType, p.Type)
			}
			if p.Description != tt.desc {
				t.Errorf("expected description %s, got %s", tt.desc, p.Description)
			}

			if tt.validator != "" && p.Validator == nil {
				t.Error("expected validator to be set")
			}
			if tt.validator == "" && p.Validator != nil {
				t.Error("expected validator to be nil")
			}
		})
	}
}

func TestNewStringParameter(t *testing.T) {
	p := NewStringParameter("test.string", "hello", "^[a-z]+$", "A string param")

	if p.Type != STRING {
		t.Errorf("expected type STRING, got %v", p.Type)
	}
	if p.Validator == nil {
		t.Error("expected validator to be set")
	}
}

func TestNewBoolParameter(t *testing.T) {
	p := NewBoolParameter("test.bool", "true", "A boolean param")

	if p.Type != BOOL {
		t.Errorf("expected type BOOL, got %v", p.Type)
	}
	if p.Validator == nil || p.Validator.String() != "^(true|false)$" {
		t.Error("expected boolean validator to be set")
	}
}

func TestNewIntParameter(t *testing.T) {
	p := NewIntParameter("test.int", "123", "An integer param")

	if p.Type != INT {
		t.Errorf("expected type INT, got %v", p.Type)
	}
	if p.Validator == nil {
		t.Error("expected integer validator to be set")
	}
}

func TestNewDecimalParameter(t *testing.T) {
	p := NewDecimalParameter("test.decimal", "3.14", "A decimal param")

	if p.Type != FLOAT {
		t.Errorf("expected type FLOAT, got %v", p.Type)
	}
	if p.Validator == nil {
		t.Error("expected decimal validator to be set")
	}
}

func TestModuleParamValidate(t *testing.T) {
	tests := []struct {
		name      string
		param     *ModuleParam
		value     string
		wantError bool
		expected  interface{}
	}{
		// String tests
		{
			name: "valid string without validator",
			param: &ModuleParam{
				Name: "test",
				Type: STRING,
			},
			value:     "any string",
			wantError: false,
			expected:  "any string",
		},
		{
			name: "valid string with validator",
			param: &ModuleParam{
				Name:      "test",
				Type:      STRING,
				Validator: regexp.MustCompile("^[a-z]+$"),
			},
			value:     "hello",
			wantError: false,
			expected:  "hello",
		},
		{
			name: "invalid string with validator",
			param: &ModuleParam{
				Name:      "test",
				Type:      STRING,
				Validator: regexp.MustCompile("^[a-z]+$"),
			},
			value:     "Hello123",
			wantError: true,
		},
		// Bool tests
		{
			name: "valid bool true",
			param: &ModuleParam{
				Name:      "test",
				Type:      BOOL,
				Validator: regexp.MustCompile("^(true|false)$"),
			},
			value:     "true",
			wantError: false,
			expected:  true,
		},
		{
			name: "valid bool false",
			param: &ModuleParam{
				Name:      "test",
				Type:      BOOL,
				Validator: regexp.MustCompile("^(true|false)$"),
			},
			value:     "false",
			wantError: false,
			expected:  false,
		},
		{
			name: "valid bool uppercase",
			param: &ModuleParam{
				Name: "test",
				Type: BOOL,
			},
			value:     "TRUE",
			wantError: false,
			expected:  true,
		},
		{
			name: "invalid bool",
			param: &ModuleParam{
				Name: "test",
				Type: BOOL,
			},
			value:     "yes",
			wantError: true,
		},
		// Int tests
		{
			name: "valid positive int",
			param: &ModuleParam{
				Name:      "test",
				Type:      INT,
				Validator: regexp.MustCompile(`^[\-\+]?[\d]+$`),
			},
			value:     "123",
			wantError: false,
			expected:  123,
		},
		{
			name: "valid negative int",
			param: &ModuleParam{
				Name:      "test",
				Type:      INT,
				Validator: regexp.MustCompile(`^[\-\+]?[\d]+$`),
			},
			value:     "-456",
			wantError: false,
			expected:  -456,
		},
		{
			name: "valid int with plus",
			param: &ModuleParam{
				Name:      "test",
				Type:      INT,
				Validator: regexp.MustCompile(`^[\-\+]?[\d]+$`),
			},
			value:     "+789",
			wantError: false,
			expected:  789,
		},
		{
			name: "invalid int",
			param: &ModuleParam{
				Name: "test",
				Type: INT,
			},
			value:     "12.34",
			wantError: true,
		},
		// Float tests
		{
			name: "valid float",
			param: &ModuleParam{
				Name:      "test",
				Type:      FLOAT,
				Validator: regexp.MustCompile(`^[\-\+]?[\d]+(\.\d+)?$`),
			},
			value:     "3.14",
			wantError: false,
			expected:  3.14,
		},
		{
			name: "valid float without decimal",
			param: &ModuleParam{
				Name:      "test",
				Type:      FLOAT,
				Validator: regexp.MustCompile(`^[\-\+]?[\d]+(\.\d+)?$`),
			},
			value:     "42",
			wantError: false,
			expected:  42.0,
		},
		{
			name: "valid negative float",
			param: &ModuleParam{
				Name:      "test",
				Type:      FLOAT,
				Validator: regexp.MustCompile(`^[\-\+]?[\d]+(\.\d+)?$`),
			},
			value:     "-2.718",
			wantError: false,
			expected:  -2.718,
		},
		{
			name: "invalid float",
			param: &ModuleParam{
				Name: "test",
				Type: FLOAT,
			},
			value:     "3.14.15",
			wantError: true,
		},
		// Invalid type test
		{
			name: "invalid type",
			param: &ModuleParam{
				Name: "test",
				Type: ParamType(999),
			},
			value:     "anything",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err, result := tt.param.validate(tt.value)

			if tt.wantError {
				if err == nil {
					t.Error("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("expected %v (%T), got %v (%T)", tt.expected, tt.expected, result, result)
				}
			}
		})
	}
}

func TestModuleParamHelp(t *testing.T) {
	p := &ModuleParam{
		Name:        "test.param",
		Description: "A test parameter",
		Value:       "default",
	}

	help := p.Help(15)

	// Check that help contains the name
	if !strings.Contains(help, "test.param") {
		t.Error("help should contain parameter name")
	}

	// Check that help contains the description
	if !strings.Contains(help, "A test parameter") {
		t.Error("help should contain parameter description")
	}

	// Check that help contains the default value
	if !strings.Contains(help, "default=default") {
		t.Error("help should contain default value")
	}
}

func TestParseSpecialValues(t *testing.T) {
	// Test the special parameter constants
	tests := []struct {
		name      string
		value     string
		isSpecial bool
	}{
		{
			name:      "interface name",
			value:     ParamIfaceName,
			isSpecial: true,
		},
		{
			name:      "interface address",
			value:     ParamIfaceAddress,
			isSpecial: true,
		},
		{
			name:      "interface address6",
			value:     ParamIfaceAddress6,
			isSpecial: true,
		},
		{
			name:      "interface mac",
			value:     ParamIfaceMac,
			isSpecial: true,
		},
		{
			name:      "subnet",
			value:     ParamSubnet,
			isSpecial: true,
		},
		{
			name:      "random mac",
			value:     ParamRandomMAC,
			isSpecial: true,
		},
		{
			name:      "normal value",
			value:     "192.168.1.1",
			isSpecial: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.isSpecial {
				// Special values should be in angle brackets
				if !strings.HasPrefix(tt.value, "<") || !strings.HasSuffix(tt.value, ">") {
					t.Errorf("special value %s should be in angle brackets", tt.value)
				}
			}
		})
	}
}

func TestParamIfaceNameParser(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		matches   bool
		ifaceName string
	}{
		{
			name:      "valid interface name",
			input:     "<eth0>",
			matches:   true,
			ifaceName: "eth0",
		},
		{
			name:      "valid interface with numbers",
			input:     "<wlan1>",
			matches:   true,
			ifaceName: "wlan1",
		},
		{
			name:      "long interface name",
			input:     "<enp0s31f6>",
			matches:   true,
			ifaceName: "enp0s31f6",
		},
		{
			name:    "no angle brackets",
			input:   "eth0",
			matches: false,
		},
		{
			name:    "invalid characters",
			input:   "<eth-0>",
			matches: false,
		},
		{
			name:    "too short",
			input:   "<e>",
			matches: false,
		},
		{
			name:    "too long",
			input:   "<verylonginterfacename>",
			matches: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := ParamIfaceNameParser.FindStringSubmatch(tt.input)

			if tt.matches {
				if len(matches) != 2 {
					t.Errorf("expected to match interface name pattern, got %v", matches)
				} else if matches[1] != tt.ifaceName {
					t.Errorf("expected interface name %s, got %s", tt.ifaceName, matches[1])
				}
			} else {
				if len(matches) > 0 {
					t.Errorf("expected no match, but got %v", matches)
				}
			}
		})
	}
}

func BenchmarkModuleParamValidate(b *testing.B) {
	p := &ModuleParam{
		Name:      "test",
		Type:      STRING,
		Validator: regexp.MustCompile("^[a-z]+$"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.validate("hello")
	}
}

func BenchmarkModuleParamValidateInt(b *testing.B) {
	p := &ModuleParam{
		Name:      "test",
		Type:      INT,
		Validator: regexp.MustCompile(`^[\-\+]?[\d]+$`),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.validate("12345")
	}
}
