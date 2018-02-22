package session

import (
	"crypto/rand"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"

	"github.com/bettercap/bettercap/core"
)

type ParamType int

const (
	STRING ParamType = iota
	BOOL             = iota
	INT              = iota
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
	return NewModuleParameter(name, def_value, INT, "^[\\d]+$", desc)
}

func (p ModuleParam) Validate(value string) (error, interface{}) {
	if p.Validator != nil {
		if p.Validator.MatchString(value) == false {
			return fmt.Errorf("Parameter %s not valid: '%s' does not match rule '%s'.", core.Bold(p.Name), value, p.Validator.String()), nil
		}
	}

	if p.Type == STRING {
		return nil, value
	} else if p.Type == BOOL {
		lvalue := strings.ToLower(value)
		if lvalue == "true" {
			return nil, true
		} else if lvalue == "false" {
			return nil, false
		} else {
			return fmt.Errorf("Can't typecast '%s' to boolean.", value), nil
		}
	} else if p.Type == INT {
		i, err := strconv.Atoi(value)
		return err, i
	}

	return fmt.Errorf("Unhandled module parameter type %d.", p.Type), nil
}

const ParamIfaceName = "<interface name>"
const ParamIfaceAddress = "<interface address>"
const ParamSubnet = "<entire subnet>"
const ParamRandomMAC = "<random mac>"

func (p ModuleParam) Get(s *Session) (error, interface{}) {
	var v string
	var found bool
	var obj interface{}
	var err error

	if found, v = s.Env.Get(p.Name); found == false {
		v = ""
	}

	if v == ParamIfaceName {
		v = s.Interface.Name()
	} else if v == ParamIfaceAddress {
		v = s.Interface.IpAddress
	} else if v == ParamSubnet {
		v = s.Interface.CIDR()
	} else if v == ParamRandomMAC {
		hw := make([]byte, 6)
		rand.Read(hw)
		v = net.HardwareAddr(hw).String()
	}

	err, obj = p.Validate(v)
	return err, obj

}

func (p ModuleParam) Dump(padding int) string {
	return fmt.Sprintf("  "+core.YELLOW+"%"+strconv.Itoa(padding)+"s"+core.RESET+
		" : %s\n", p.Name, p.Value)
}

func (p ModuleParam) Help(padding int) string {
	return fmt.Sprintf("  "+core.YELLOW+"%"+strconv.Itoa(padding)+"s"+core.RESET+
		" : "+
		"%s "+core.DIM+"(default=%s"+core.RESET+")\n", p.Name, p.Description, p.Value)
}

func (p ModuleParam) Register(s *Session) {
	s.Env.Set(p.Name, p.Value)
}
