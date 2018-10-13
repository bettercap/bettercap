package plugin

import (
	// "unicode"
	"github.com/robertkrimen/otto"
)

func (p *Plugin) compile() (err error) {
	// create a new vm
	p.vm = otto.New()
	// track objects already defined by Otto
	predefined := map[string]bool{}
	for name := range p.vm.Context().Symbols {
		predefined[name] = true
	}
	// run the code once in order to define all the functions
	// and validate the syntax, then get the callbacks
	if _, err = p.vm.Run(p.Code); err != nil {
		return
	}
	// every uppercase object is considered exported
	for name, sym := range p.vm.Context().Symbols {
		// ignore predefined objects
		if _, found := predefined[name]; !found {
			// ignore lowercase global objects
			// if unicode.IsUpper(rune(name[0])) {
			if sym.IsFunction() {
				p.callbacks[name] = sym
			} else {
				p.objects[name] = sym
			}
			// }
		}
	}
	return nil
}
