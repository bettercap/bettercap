package ast_test

import (
	"fmt"
	"log"

	"github.com/robertkrimen/otto/ast"
	"github.com/robertkrimen/otto/file"
	"github.com/robertkrimen/otto/parser"
)

type walkExample struct {
	source string
	shift  file.Idx
}

func (w *walkExample) Enter(n ast.Node) ast.Visitor {
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

func (w *walkExample) Exit(n ast.Node) {
	// AST node n has had all its children walked. Pop it out of your
	// stack, or do whatever processing you need to do, if any.
}

func ExampleVisitor_codeRewrite() {
	source := `var b = function() {test(); try {} catch(e) {} var test = "test(); var test = 1"} // test`
	program, err := parser.ParseFile(nil, "", source, 0)
	if err != nil {
		log.Fatal(err)
	}

	w := &walkExample{source: source}

	ast.Walk(w, program)

	fmt.Println(w.source)
	// Output: var varnew_b = function() {new_test(); try {} catch(new_e) {} var varnew_test = "test(); var test = 1"} // test
}
