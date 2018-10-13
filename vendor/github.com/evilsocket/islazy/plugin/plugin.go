package plugin

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"

	"github.com/robertkrimen/otto"
)

// Defines is a map containing the predefined objects
// and functions for each vm of each plugin.
var Defines = map[string]interface{}{}

// Plugin is an object representing a javascript
// file exporting functions and variables that
// your project can use to extend its functionalities.
type Plugin struct {
	sync.Mutex
	// The basename of the plugin.
	Name string
	// The actual javascript code.
	Code string
	// The full path of the plugin.
	Path string

	vm        *otto.Otto
	callbacks map[string]otto.Value
	objects   map[string]otto.Value
}

// Parse parsesand compiles a plugin given its source code.
func Parse(code string) (*Plugin, error) {
	plugin := &Plugin{
		Code:      code,
		callbacks: make(map[string]otto.Value),
		objects:   make(map[string]otto.Value),
	}

	if err := plugin.compile(); err != nil {
		return nil, err
	}

	for name, val := range Defines {
		if err := plugin.vm.Set(name, val); err != nil {
			return nil, err
		}
	}

	return plugin, nil
}

// Load loads and compiles a plugin given its path.
func Load(path string) (plug *Plugin, err error) {
	if raw, err := ioutil.ReadFile(path); err != nil {
		return nil, err
	} else if plug, err = Parse(string(raw)); err != nil {
		return nil, err
	} else {
		plug.Path = path
		plug.Name = strings.Replace(filepath.Base(path), ".js", "", -1)
	}
	return plug, nil
}

// Clone returns a new instance identical to the plugin.
func (p *Plugin) Clone() (clone *Plugin) {
	var err error
	if p.Path == "" {
		clone, err = Parse(p.Code)
	} else {
		clone, err = Load(p.Path)
	}
	if err != nil {
		panic(err) // this should never happen
	}
	return clone
}

// HasFunc returns true if the function with `name`
// has been declared in the plugin code.
func (p *Plugin) HasFunc(name string) bool {
	_, found := p.callbacks[name]
	return found
}

// Set sets a variable into the VM of this plugin instance.
func (p *Plugin) Set(name string, v interface{}) error {
	p.Lock()
	defer p.Unlock()
	return p.vm.Set(name, v)
}

// Call executes one of the declared callbacks of the plugin by its name.
func (p *Plugin) Call(name string, args ...interface{}) (interface{}, error) {
	p.Lock()
	defer p.Unlock()

	if cb, found := p.callbacks[name]; !found {
		return nil, fmt.Errorf("%s does not name a function", name)
	} else if ret, err := cb.Call(otto.NullValue(), args...); err != nil {
		return nil, err
	} else if !ret.IsUndefined() {
		exported, err := ret.Export()
		if err != nil {
			return nil, err
		}
		return exported, nil
	}
	return nil, nil
}

// Methods returns a list of methods exported from the javascript
func (p *Plugin) Methods() []string {
	methods := []string{}
	for key, _ := range p.callbacks {
		methods = append(methods, key)
	}
	return methods
}

// Objects returns a list of object exported by the javascript
func (p *Plugin) Objects() []string {
	objs := []string{}
	for key, _ := range p.callbacks {
		objs = append(objs, key)
	}
	return objs
}

// GetTypeObject returns the type of the object by its name
func (p *Plugin) GetTypeObject(name string) string {
	if obj, found := p.objects[name]; !found {
		return ""
	} else if obj.IsPrimitive() {
		if obj.IsBoolean() {
			return "BooleanPrimitive"
		} else if obj.IsNumber() {
			return "NumberPrimitive"
		} else if obj.IsString() {
			return "StringPrimitive"
		}
	} else if obj.IsObject() {
		switch obj.Class() {
		case "Array":
			return "ArrayObject"
		case "String":
			return "StringObject"
		case "Boolean":
			return "BooleanObject"
		case "Number":
			return "NumberObject"
		case "Date":
			return "DateObject"
		case "RegExp":
			return "RegExpObject"
		case "Error":
			return "ErrorObject"
		}
	}
	return ""
}

// IsStringPrimitive returns true if the object with a
// given name is a javascript primitive string
func (p *Plugin) IsStringPrimitive(name string) bool {
	return p.GetTypeObject(name) == "StringPrimitive"
}

// IsBooleanPrimitive returns true if the object with a
// given name is a javascript primitive boolean, false otherwise
func (p *Plugin) IsBooleanPrimitive(name string) bool {
	return p.GetTypeObject(name) == "BooleanPrimitive"
}

// IsNumberPrimitive returns true if the object with a
// given name is a javascript primitive number, false otherwise
func (p *Plugin) IsNumberPrimitive(name string) bool {
	return p.GetTypeObject(name) == "NumberPrimitive"
}

// IsArrayObject returns true if the object with a
// given name is a javascript array object, false otherwise
func (p *Plugin) IsArrayObject(name string) bool {
	return p.GetTypeObject(name) == "ArrayObject"
}

// IsStringObject returns true if the object with a
// given name is a javascript string object, false otherwise
func (p *Plugin) IsStringObject(name string) bool {
	return p.GetTypeObject(name) == "StringObject"
}

// IsBooleanObject returns true if the object with a
// given name is a javascript boolean object, false otherwise
func (p *Plugin) IsBooleanObject(name string) bool {
	return p.GetTypeObject(name) == "BooleanObject"
}

// IsNumberObject returns true if the object with a
// given name is a javascript Number object, false otherwise
func (p *Plugin) IsNumberObject(name string) bool {
	return p.GetTypeObject(name) == "NumberObject"
}

// IsDateObject returns true if the object with a
// given name is a javascript Date object, false otherwise
func (p *Plugin) IsDateObject(name string) bool {
	return p.GetTypeObject(name) == "DateObject"
}

// IsRegExpObject returns true if the object with a
// given name is a javascript RegExp object, false otherwise
func (p *Plugin) IsRegExpObject(name string) bool {
	return p.GetTypeObject(name) == "RegExpObject"
}

// IsErrorObject returns true if the object with a
// given name is a javascript error object, false otherwise
func (p *Plugin) IsErrorObject(name string) bool {
	return p.GetTypeObject(name) == "ErrorObject"
}

// GetObject returns an interface containing the value of the object by its name
func (p *Plugin) GetObject(name string) (interface{}, error) {
	if obj, found := p.objects[name]; !found {
		return nil, fmt.Errorf("%s does not name an object", name)
	} else {
		return obj.Export()
	}
}
