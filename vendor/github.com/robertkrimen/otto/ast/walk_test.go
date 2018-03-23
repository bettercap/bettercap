package ast_test

import (
	"log"
	"testing"

	"github.com/robertkrimen/otto/ast"
	"github.com/robertkrimen/otto/file"
	"github.com/robertkrimen/otto/parser"
)

type walker struct {
	stack  []ast.Node
	source string
	shift  file.Idx
}

// push and pop below are to prove the symmetry of Enter/Exit calls

func (w *walker) push(n ast.Node) {
	w.stack = append(w.stack, n)
}

func (w *walker) pop(n ast.Node) {
	size := len(w.stack)
	if size <= 0 {
		panic("pop of empty stack")
	}

	toPop := w.stack[size-1]
	if toPop != n {
		panic("pop: nodes do not equal")
	}

	w.stack[size-1] = nil
	w.stack = w.stack[:size-1]
}

func (w *walker) Enter(n ast.Node) ast.Visitor {
	w.push(n)

	if id, ok := n.(*ast.Identifier); ok && id != nil {
		idx := n.Idx0() + w.shift - 1
		s := w.source[:idx] + "new_" + w.source[idx:]
		w.source = s
		w.shift += 4
	}
	if v, ok := n.(*ast.VariableExpression); ok && v != nil {
		idx := n.Idx0() + w.shift - 1
		s := w.source[:idx] + "varnew_" + w.source[idx:]
		w.source = s
		w.shift += 7
	}

	return w
}

func (w *walker) Exit(n ast.Node) {
	w.pop(n)
}

func TestVisitorRewrite(t *testing.T) {
	source := `var b = function() {test(); try {} catch(e) {} var test = "test(); var test = 1"} // test`
	program, err := parser.ParseFile(nil, "", source, 0)
	if err != nil {
		log.Fatal(err)
	}

	w := &walker{source: source}

	ast.Walk(w, program)

	xformed := `var varnew_b = function() {new_test(); try {} catch(new_e) {} var varnew_test = "test(); var test = 1"} // test`

	if w.source != xformed {
		t.Errorf("source is `%s` not `%s`", w.source, xformed)
	}

	if len(w.stack) != 0 {
		t.Errorf("stack should be empty, but is length: %d", len(w.stack))
	}
}
