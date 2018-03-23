package nmea

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var parsertests = []struct {
	name     string
	fields   []string
	expected interface{}
	hasErr   bool
	parse    func(p *parser) interface{}
}{
	{
		name:     "String",
		fields:   []string{"foo", "bar"},
		expected: "bar",
		parse: func(p *parser) interface{} {
			return p.String(1, "")
		},
	},
	{
		name:     "String out of range",
		fields:   []string{"wot"},
		expected: "",
		hasErr:   true,
		parse: func(p *parser) interface{} {
			return p.String(5, "thing")
		},
	},
	{
		name:     "String with existing error",
		expected: "",
		hasErr:   true,
		parse: func(p *parser) interface{} {
			p.SetErr("context", "value")
			return p.String(123, "blah")
		},
	},
	{
		name:     "EnumString",
		fields:   []string{"a", "b", "c"},
		expected: "b",
		parse: func(p *parser) interface{} {
			return p.EnumString(1, "context", "b", "d")
		},
	},
	{
		name:     "EnumString invalid",
		fields:   []string{"a", "b", "c"},
		expected: "",
		hasErr:   true,
		parse: func(p *parser) interface{} {
			return p.EnumString(1, "context", "x", "y")
		},
	},
	{
		name:     "EnumString with existing error",
		fields:   []string{"a", "b", "c"},
		expected: "",
		hasErr:   true,
		parse: func(p *parser) interface{} {
			p.SetErr("context", "value")
			return p.EnumString(1, "context", "a", "b")
		},
	},
	{
		name:     "Int64",
		fields:   []string{"123"},
		expected: int64(123),
		parse: func(p *parser) interface{} {
			return p.Int64(0, "context")
		},
	},
	{
		name:     "Int64 empty field is zero",
		fields:   []string{""},
		expected: int64(0),
		parse: func(p *parser) interface{} {
			return p.Int64(0, "context")
		},
	},
	{
		name:     "Int64 invalid",
		fields:   []string{"abc"},
		expected: int64(0),
		hasErr:   true,
		parse: func(p *parser) interface{} {
			return p.Int64(0, "context")
		},
	},
	{
		name:     "Int64 with existing error",
		fields:   []string{"123"},
		expected: int64(0),
		hasErr:   true,
		parse: func(p *parser) interface{} {
			p.SetErr("context", "value")
			return p.Int64(0, "context")
		},
	},
	{
		name:     "Float64",
		fields:   []string{"123.123"},
		expected: float64(123.123),
		parse: func(p *parser) interface{} {
			return p.Float64(0, "context")
		},
	},
	{
		name:     "Float64 empty field is zero",
		fields:   []string{""},
		expected: float64(0),
		parse: func(p *parser) interface{} {
			return p.Float64(0, "context")
		},
	},
	{
		name:     "Float64 invalid",
		fields:   []string{"abc"},
		expected: float64(0),
		hasErr:   true,
		parse: func(p *parser) interface{} {
			return p.Float64(0, "context")
		},
	},
	{
		name:     "Float64 with existing error",
		fields:   []string{"123.123"},
		expected: float64(0),
		hasErr:   true,
		parse: func(p *parser) interface{} {
			p.SetErr("context", "value")
			return p.Float64(0, "context")
		},
	},
	{
		name:     "Time",
		fields:   []string{"123456"},
		expected: Time{true, 12, 34, 56, 0},
		parse: func(p *parser) interface{} {
			return p.Time(0, "context")
		},
	},
	{
		name:     "Time empty field is zero",
		fields:   []string{""},
		expected: Time{},
		parse: func(p *parser) interface{} {
			return p.Time(0, "context")
		},
	},
	{
		name:     "Time with existing error",
		fields:   []string{"123456"},
		expected: Time{},
		hasErr:   true,
		parse: func(p *parser) interface{} {
			p.SetErr("context", "value")
			return p.Time(0, "context")
		},
	},
	{
		name:     "Time invalid",
		fields:   []string{"wrong"},
		expected: Time{},
		hasErr:   true,
		parse: func(p *parser) interface{} {
			return p.Time(0, "context")
		},
	},
	{
		name:     "Date",
		fields:   []string{"010203"},
		expected: Date{true, 1, 2, 3},
		parse: func(p *parser) interface{} {
			return p.Date(0, "context")
		},
	},
	{
		name:     "Date empty field is zero",
		fields:   []string{""},
		expected: Date{},
		parse: func(p *parser) interface{} {
			return p.Date(0, "context")
		},
	},
	{
		name:     "Date invalid",
		fields:   []string{"Hello"},
		expected: Date{},
		hasErr:   true,
		parse: func(p *parser) interface{} {
			return p.Date(0, "context")
		},
	},
	{
		name:     "Date with existing error",
		fields:   []string{"010203"},
		expected: Date{},
		hasErr:   true,
		parse: func(p *parser) interface{} {
			p.SetErr("context", "value")
			return p.Date(0, "context")
		},
	},
}

func TestParser(t *testing.T) {
	for _, tt := range parsertests {
		t.Run(tt.name, func(t *testing.T) {
			p := newParser(Sent{
				Type:   "type",
				Fields: tt.fields,
			}, "type")
			assert.Equal(t, tt.expected, tt.parse(p))
			if tt.hasErr {
				assert.Error(t, p.Err())
			} else {
				assert.NoError(t, p.Err())
			}
		})
	}
}
