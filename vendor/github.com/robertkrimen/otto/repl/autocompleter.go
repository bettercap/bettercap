package repl

import (
	"regexp"
	"strings"

	"github.com/robertkrimen/otto"
)

type autoCompleter struct {
	vm *otto.Otto
}

var lastExpressionRegex = regexp.MustCompile(`[a-zA-Z0-9]([a-zA-Z0-9\.]*[a-zA-Z0-9])?\.?$`)

func (a *autoCompleter) Do(line []rune, pos int) ([][]rune, int) {
	lastExpression := lastExpressionRegex.FindString(string(line))

	bits := strings.Split(lastExpression, ".")

	first := bits[:len(bits)-1]
	last := bits[len(bits)-1]

	var l []string

	if len(first) == 0 {
		c := a.vm.Context()

		l = make([]string, len(c.Symbols))

		i := 0
		for k := range c.Symbols {
			l[i] = k
			i++
		}
	} else {
		r, err := a.vm.Eval(strings.Join(bits[:len(bits)-1], "."))
		if err != nil {
			return nil, 0
		}

		if o := r.Object(); o != nil {
			for _, v := range o.KeysByParent() {
				l = append(l, v...)
			}
		}
	}

	var r [][]rune
	for _, s := range l {
		if strings.HasPrefix(s, last) {
			r = append(r, []rune(strings.TrimPrefix(s, last)))
		}
	}

	return r, len(last)
}
