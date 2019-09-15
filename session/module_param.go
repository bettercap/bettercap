package session

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"

	"github.com/evilsocket/islazy/tui"
)

type ParamType int

const (
	STRING ParamType = iota
	BOOL             = iota
	INT              = iota
	FLOAT            = iota
)

type ModuleParam struct {
	Name        string
	Type        ParamType
	Value       string
	Description string

	Validator *regexp.Regexp
}

func NewModuleParameter(name string, def_value string, t ParamType, validator string, desc string) *ModuleParam {
	p := &ModuleParam{
		Name:        name,
		Type:        t,
		Description: desc,
		Value:       def_value,
		Validator:   nil,
	}

	if validator != "" {
		p.Validator = regexp.MustCompile(validator)
	}

	return p
}

func NewStringParameter(name string, def_value string, validator string, desc string) *ModuleParam {
	return NewModuleParameter(name, def_value, STRING, validator, desc)
}

func NewBoolParameter(name string, def_value string, desc string) *ModuleParam {
	return NewModuleParameter(name, def_value, BOOL, "^(true|false)$", desc)
}

func NewIntParameter(name string, def_value string, desc string) *ModuleParam {
	return NewModuleParameter(name, def_value, INT, `^[\-\+]?[\d]+$`, desc)
}

func NewDecimalParameter(name string, def_value string, desc string) *ModuleParam {
	return NewModuleParameter(name, def_value, FLOAT, "^[\\d]+(\\.\\d+)?$", desc)
}

func (p ModuleParam) validate(value string) (error, interface{}) {
	if p.Validator != nil {
		if !p.Validator.MatchString(value) {
			return fmt.Errorf("Parameter %s not valid: '%s' does not match rule '%s'.", tui.Bold(p.Name), value, p.Validator.String()), nil
		}
	}

	switch p.Type {
	case STRING:
		return nil, value
	case BOOL:
		lvalue := strings.ToLower(value)
		if lvalue == "true" {
			return nil, true
		} else if lvalue == "false" {
			return nil, false
		} else {
			return fmt.Errorf("Can't typecast '%s' to boolean.", value), nil
		}
	case INT:
		i, err := strconv.Atoi(value)
		return err, i
	case FLOAT:
		i, err := strconv.ParseFloat(value, 64)
		return err, i
	}

	return fmt.Errorf("Unhandled module parameter type %d.", p.Type), nil
}

const ParamIfaceName = "<interface name>"
const ParamIfaceAddress = "<interface address>"
const ParamIfaceAddress6 = "<interface address6>"
const ParamSubnet = "<entire subnet>"
const ParamRandomMAC = "<random mac>"

func (p ModuleParam) parse(s *Session, v string) string {
	switch v {
	case ParamIfaceName:
		v = s.Interface.Name()
	case ParamIfaceAddress:
		v = s.Interface.IpAddress
	case ParamIfaceAddress6:
		v = s.Interface.Ip6Address
	case ParamSubnet:
		v = s.Interface.CIDR()
	case ParamRandomMAC:
		hw := make([]byte, 6)
		rand.Read(hw)
		v = net.HardwareAddr(hw).String()
	}
	return v

}

func (p ModuleParam) getUnlocked(s *Session) string {
	_, v := s.Env.GetUnlocked(p.Name)
	return p.parse(s, v)
}

func (p ModuleParam) Get(s *Session) (error, interface{}) {
	_, v := s.Env.Get(p.Name)
	v = p.parse(s, v)
	return p.validate(v)
}

func (p ModuleParam) Help(padding int) string {
	return fmt.Sprintf("  "+tui.YELLOW+"%"+strconv.Itoa(padding)+"s"+tui.RESET+
		" : "+
		"%s "+tui.DIM+"(default=%s"+tui.RESET+")\n", p.Name, p.Description, p.Value)
}

func (p ModuleParam) Register(s *Session) {
	s.Env.Set(p.Name, p.Value)
}

func (p ModuleParam) RegisterObserver(s *Session, cb EnvironmentChangedCallback) {
	s.Env.WithCallback(p.Name, p.Value, cb)
}

type JSONModuleParam struct {
	Name        string    `json:"name"`
	Type        ParamType `json:"type"`
	Description string    `json:"description"`
	Value       string    `json:"default_value"`
	Current     string    `json:"current_value"`
	Validator   string    `json:"validator"`
}

func (p ModuleParam) MarshalJSON() ([]byte, error) {
	j := JSONModuleParam{
		Name:        p.Name,
		Type:        p.Type,
		Description: p.Description,
		Value:       p.Value,
		Current:     p.getUnlocked(I), // if we're here, Env is already locked
	}
	if p.Validator != nil {
		j.Validator = p.Validator.String()
	}
	return json.Marshal(j)
}
