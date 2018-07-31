package nmea

import (
	"fmt"
	"strconv"
)

// parser provides a simple way of accessing and parsing
// sentence fields
type parser struct {
	BaseSentence
	prefix string
	err    error
}

// newParser constructor
func newParser(s BaseSentence, prefix string) *parser {
	p := &parser{BaseSentence: s, prefix: prefix}
	if p.Type != prefix {
		p.SetErr("prefix", p.Type)
	}
	return p
}

// Err returns the first error encountered during the parser's usage.
func (p *parser) Err() error {
	return p.err
}

// SetErr assigns an error. Calling this method has no
// effect if there is already an error.
func (p *parser) SetErr(context, value string) {
	if p.err == nil {
		p.err = fmt.Errorf("nmea: %s invalid %s: %s", p.prefix, context, value)
	}
}

// String returns the field value at the specified index.
func (p *parser) String(i int, context string) string {
	if p.err != nil {
		return ""
	}
	if i < 0 || i >= len(p.Fields) {
		p.SetErr(context, "index out of range")
		return ""
	}
	return p.Fields[i]
}

// EnumString returns the field value at the specified index.
// An error occurs if the value is not one of the options and not empty.
func (p *parser) EnumString(i int, context string, options ...string) string {
	s := p.String(i, context)
	if p.err != nil || s == "" {
		return ""
	}
	for _, o := range options {
		if o == s {
			return s
		}
	}
	p.SetErr(context, s)
	return ""
}

// Int64 returns the int64 value at the specified index.
// If the value is an empty string, 0 is returned.
func (p *parser) Int64(i int, context string) int64 {
	s := p.String(i, context)
	if p.err != nil {
		return 0
	}
	if s == "" {
		return 0
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		p.SetErr(context, s)
	}
	return v
}

// Float64 returns the float64 value at the specified index.
// If the value is an empty string, 0 is returned.
func (p *parser) Float64(i int, context string) float64 {
	s := p.String(i, context)
	if p.err != nil {
		return 0
	}
	if s == "" {
		return 0
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		p.SetErr(context, s)
	}
	return v
}

// Time returns the Time value at the specified index.
// If the value is empty, the Time is marked as invalid.
func (p *parser) Time(i int, context string) Time {
	s := p.String(i, context)
	if p.err != nil {
		return Time{}
	}
	v, err := ParseTime(s)
	if err != nil {
		p.SetErr(context, s)
	}
	return v
}

// Date returns the Date value at the specified index.
// If the value is empty, the Date is marked as invalid.
func (p *parser) Date(i int, context string) Date {
	s := p.String(i, context)
	if p.err != nil {
		return Date{}
	}
	v, err := ParseDate(s)
	if err != nil {
		p.SetErr(context, s)
	}
	return v
}

// LatLong returns the coordinate value of the specified fields.
func (p *parser) LatLong(i, j int, context string) float64 {
	a := p.String(i, context)
	b := p.String(j, context)
	if p.err != nil {
		return 0
	}
	s := fmt.Sprintf("%s %s", a, b)
	v, err := ParseLatLong(s)
	if err != nil {
		p.SetErr(context, err.Error())
	}
	return v
}
