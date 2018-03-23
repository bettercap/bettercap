package parser

import (
	"fmt"
	"github.com/robertkrimen/otto/ast"
	"reflect"
	"testing"
)

func checkComments(actual []*ast.Comment, expected []string, position ast.CommentPosition) error {
	var comments []*ast.Comment
	for _, c := range actual {
		if c.Position == position {
			comments = append(comments, c)
		}
	}

	if len(comments) != len(expected) {
		return fmt.Errorf("the number of comments is not correct. %v != %v", len(comments), len(expected))
	}

	for i, v := range comments {
		if v.Text != expected[i] {
			return fmt.Errorf("comments do not match. \"%v\" != \"%v\"\n", v.Text, expected[i])
		}
		if v.Position != position {
			return fmt.Errorf("the comment is not %v, was %v\n", position, v.Position)
		}
	}

	return nil
}

func displayStatements(statements []ast.Statement) {
	fmt.Printf("Printing statements (%v)\n", len(statements))
	for k, v := range statements {
		fmt.Printf("Type of line %v: %v\n", k, reflect.TypeOf(v))
	}
}

func displayComments(m ast.CommentMap) {
	fmt.Printf("Displaying comments:\n")
	for n, comments := range m {
		fmt.Printf("%v %v:\n", reflect.TypeOf(n), n)
		for i, comment := range comments {
			fmt.Printf(" [%v] %v @ %v\n", i, comment.Text, comment.Position)
		}
	}
}

func TestParser_comments(t *testing.T) {
	tt(t, func() {
		test := func(source string, chk interface{}) (*_parser, *ast.Program) {
			parser, program, err := testParseWithMode(source, StoreComments)
			is(firstErr(err), chk)

			// Check unresolved comments
			//is(len(parser.comments), 0)
			return parser, program
		}

		var err error
		var parser *_parser
		var program *ast.Program

		parser, program = test("q=2;// Hej\nv = 0", nil)
		is(len(program.Body), 2)
		err = checkComments((parser.comments.CommentMap)[program.Body[1]], []string{" Hej"}, ast.LEADING)
		is(err, nil)

		// Assignment
		parser, program = test("i = /*test=*/ 2", nil)
		is(len(program.Body), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0]], []string{}, ast.LEADING), nil)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ExpressionStatement).Expression.(*ast.AssignExpression).Right], []string{"test="}, ast.LEADING), nil)
		is(parser.comments.CommentMap.Size(), 1)

		// Conditional, before consequent
		parser, program = test("i ? /*test?*/ 2 : 3", nil)
		is(len(program.Body), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0]], []string{}, ast.LEADING), nil)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ExpressionStatement).Expression.(*ast.ConditionalExpression).Consequent], []string{"test?"}, ast.LEADING), nil)
		is(parser.comments.CommentMap.Size(), 1)

		// Conditional, after consequent
		parser, program = test("i ? 2 /*test?*/ : 3", nil)
		is(len(program.Body), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0]], []string{}, ast.LEADING), nil)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ExpressionStatement).Expression.(*ast.ConditionalExpression).Consequent], []string{"test?"}, ast.TRAILING), nil)
		is(parser.comments.CommentMap.Size(), 1)

		// Conditional, before alternate
		parser, program = test("i ? 2 : /*test:*/ 3", nil)
		is(len(program.Body), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0]], []string{}, ast.LEADING), nil)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ExpressionStatement).Expression.(*ast.ConditionalExpression).Alternate], []string{"test:"}, ast.LEADING), nil)
		is(parser.comments.CommentMap.Size(), 1)

		// Logical OR
		parser, program = test("i || /*test||*/ 2", nil)
		is(len(program.Body), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0]], []string{}, ast.LEADING), nil)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ExpressionStatement).Expression.(*ast.BinaryExpression).Right], []string{"test||"}, ast.LEADING), nil)
		is(parser.comments.CommentMap.Size(), 1)

		// Logical AND
		parser, program = test("i && /*test&&*/ 2", nil)
		is(len(program.Body), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0]], []string{}, ast.LEADING), nil)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ExpressionStatement).Expression.(*ast.BinaryExpression).Right], []string{"test&&"}, ast.LEADING), nil)
		is(parser.comments.CommentMap.Size(), 1)

		// Bitwise OR
		parser, program = test("i | /*test|*/ 2", nil)
		is(len(program.Body), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0]], []string{}, ast.LEADING), nil)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ExpressionStatement).Expression.(*ast.BinaryExpression).Right], []string{"test|"}, ast.LEADING), nil)
		is(parser.comments.CommentMap.Size(), 1)

		// Exclusive OR
		parser, program = test("i ^ /*test^*/ 2", nil)
		is(len(program.Body), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0]], []string{}, ast.LEADING), nil)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ExpressionStatement).Expression.(*ast.BinaryExpression).Right], []string{"test^"}, ast.LEADING), nil)
		is(parser.comments.CommentMap.Size(), 1)

		// Bitwise AND
		parser, program = test("i & /*test&*/ 2", nil)
		is(len(program.Body), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0]], []string{}, ast.LEADING), nil)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ExpressionStatement).Expression.(*ast.BinaryExpression).Right], []string{"test&"}, ast.LEADING), nil)
		is(parser.comments.CommentMap.Size(), 1)

		// Equality
		parser, program = test("i == /*test==*/ 2", nil)
		is(len(program.Body), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0]], []string{}, ast.LEADING), nil)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ExpressionStatement).Expression.(*ast.BinaryExpression).Right], []string{"test=="}, ast.LEADING), nil)
		is(parser.comments.CommentMap.Size(), 1)

		// Relational, <
		parser, program = test("i < /*test<*/ 2", nil)
		is(len(program.Body), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0]], []string{}, ast.LEADING), nil)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ExpressionStatement).Expression.(*ast.BinaryExpression).Right], []string{"test<"}, ast.LEADING), nil)
		is(parser.comments.CommentMap.Size(), 1)

		// Relational, instanceof
		parser, program = test("i instanceof /*testinstanceof*/ thing", nil)
		is(len(program.Body), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0]], []string{}, ast.LEADING), nil)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ExpressionStatement).Expression.(*ast.BinaryExpression).Right], []string{"testinstanceof"}, ast.LEADING), nil)
		is(parser.comments.CommentMap.Size(), 1)

		// Shift left
		parser, program = test("i << /*test<<*/ 2", nil)
		is(len(program.Body), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0]], []string{}, ast.LEADING), nil)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ExpressionStatement).Expression.(*ast.BinaryExpression).Right], []string{"test<<"}, ast.LEADING), nil)
		is(parser.comments.CommentMap.Size(), 1)

		// +
		parser, program = test("i + /*test+*/ 2", nil)
		is(len(program.Body), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0]], []string{}, ast.LEADING), nil)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ExpressionStatement).Expression.(*ast.BinaryExpression).Right], []string{"test+"}, ast.LEADING), nil)
		is(parser.comments.CommentMap.Size(), 1)

		// *
		parser, program = test("i * /*test**/ 2", nil)
		is(len(program.Body), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0]], []string{}, ast.LEADING), nil)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ExpressionStatement).Expression.(*ast.BinaryExpression).Right], []string{"test*"}, ast.LEADING), nil)
		is(parser.comments.CommentMap.Size(), 1)

		// Unary prefix, ++
		parser, program = test("++/*test++*/i", nil)
		is(len(program.Body), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0]], []string{}, ast.LEADING), nil)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ExpressionStatement).Expression.(*ast.UnaryExpression).Operand], []string{"test++"}, ast.LEADING), nil)
		is(parser.comments.CommentMap.Size(), 1)

		// Unary prefix, delete
		parser, program = test("delete /*testdelete*/ i", nil)
		is(len(program.Body), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0]], []string{}, ast.LEADING), nil)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ExpressionStatement).Expression.(*ast.UnaryExpression).Operand], []string{"testdelete"}, ast.LEADING), nil)
		is(parser.comments.CommentMap.Size(), 1)

		// Unary postfix, ++
		parser, program = test("i/*test++*/++", nil)
		is(len(program.Body), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0]], []string{}, ast.LEADING), nil)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ExpressionStatement).Expression.(*ast.UnaryExpression).Operand], []string{"test++"}, ast.TRAILING), nil)
		is(parser.comments.CommentMap.Size(), 1)

		// + pt 2
		parser, program = test("i /*test+*/ + 2", nil)
		is(len(program.Body), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0]], []string{}, ast.LEADING), nil)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ExpressionStatement).Expression.(*ast.BinaryExpression).Left], []string{"test+"}, ast.TRAILING), nil)
		is(parser.comments.CommentMap.Size(), 1)

		// Multiple comments for a single node
		parser, program = test("i /*test+*/ /*test+2*/ + 2", nil)
		is(len(program.Body), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0]], []string{}, ast.LEADING), nil)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ExpressionStatement).Expression.(*ast.BinaryExpression).Left], []string{"test+", "test+2"}, ast.TRAILING), nil)
		is(parser.comments.CommentMap.Size(), 2)

		// Multiple comments for multiple nodes
		parser, program = test("i /*test1*/ + 2 /*test2*/ + a /*test3*/ * x /*test4*/", nil)
		is(len(program.Body), 1)
		is(parser.comments.CommentMap.Size(), 4)

		// Leading comment
		parser, program = test("/*leadingtest*/i + 2", nil)
		is(len(program.Body), 1)

		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ExpressionStatement)], []string{"leadingtest"}, ast.LEADING), nil)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ExpressionStatement).Expression.(*ast.BinaryExpression).Left], []string{}, ast.TRAILING), nil)
		is(parser.comments.CommentMap.Size(), 1)

		// Leading comment, with semicolon
		parser, program = test("/*leadingtest;*/;i + 2", nil)
		is(len(program.Body), 2)
		is(checkComments((parser.comments.CommentMap)[program.Body[1]], []string{"leadingtest;"}, ast.LEADING), nil)
		is(parser.comments.CommentMap.Size(), 1)

		// Arrays
		parser, program = test("[1, 2 /*test2*/, 3]", nil)
		is(len(program.Body), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ExpressionStatement).Expression.(*ast.ArrayLiteral).Value[1]], []string{"test2"}, ast.TRAILING), nil)
		is(parser.comments.CommentMap.Size(), 1)

		// Function calls
		parser, program = test("fun(a,b) //test", nil)
		is(len(program.Body), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ExpressionStatement).Expression.(*ast.CallExpression)], []string{"test"}, ast.TRAILING), nil)
		is(parser.comments.CommentMap.Size(), 1)

		// Function calls, pt 2
		parser, program = test("fun(a/*test1*/,b)", nil)
		is(len(program.Body), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ExpressionStatement).Expression.(*ast.CallExpression).ArgumentList[0]], []string{"test1"}, ast.TRAILING), nil)
		is(parser.comments.CommentMap.Size(), 1)

		// Function calls, pt 3
		parser, program = test("fun(/*test1*/a,b)", nil)
		is(len(program.Body), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ExpressionStatement).Expression.(*ast.CallExpression).ArgumentList[0]], []string{"test1"}, ast.LEADING), nil)
		is(parser.comments.CommentMap.Size(), 1)

		// Arrays pt 2
		parser, program = test(`["abc".substr(0,1)/*testa*/,
            "abc.substr(0,2)"];`, nil)
		is(len(program.Body), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ExpressionStatement).Expression.(*ast.ArrayLiteral).Value[0]], []string{"testa"}, ast.TRAILING), nil)
		is(parser.comments.CommentMap.Size(), 1)

		// Arrays pt 3
		parser, program = test(`[a, //test
            b];`, nil)
		is(len(program.Body), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ExpressionStatement).Expression.(*ast.ArrayLiteral).Value[1]], []string{"test"}, ast.LEADING), nil)
		is(parser.comments.CommentMap.Size(), 1)

		// Arrays pt 4
		parser, program = test(`[a, //test
		b, c];`, nil)
		is(len(program.Body), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ExpressionStatement).Expression.(*ast.ArrayLiteral).Value[1]], []string{"test"}, ast.LEADING), nil)
		is(parser.comments.CommentMap.Size(), 1)

		// Arrays pt 5
		parser, program = test(`
[
	"a1", // "a"
	"a2", // "ab"
];
        `, nil)
		is(parser.comments.CommentMap.Size(), 2)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ExpressionStatement).Expression.(*ast.ArrayLiteral).Value[1]], []string{" \"a\""}, ast.LEADING), nil)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ExpressionStatement).Expression.(*ast.ArrayLiteral)], []string{" \"ab\""}, ast.FINAL), nil)

		// Arrays pt 6
		parser, program = test(`[a, /*test*/ b, c];`, nil)
		is(len(program.Body), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ExpressionStatement).Expression.(*ast.ArrayLiteral).Value[1]], []string{"test"}, ast.LEADING), nil)
		is(parser.comments.CommentMap.Size(), 1)

		// Arrays pt 7 - Empty node
		parser, program = test(`[a,,/*test2*/,];`, nil)
		is(len(program.Body), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ExpressionStatement).Expression.(*ast.ArrayLiteral).Value[2]], []string{"test2"}, ast.TRAILING), nil)
		is(parser.comments.CommentMap.Size(), 1)

		// Arrays pt 8 - Trailing node
		parser, program = test(`[a,,,/*test2*/];`, nil)
		is(len(program.Body), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ExpressionStatement).Expression.(*ast.ArrayLiteral)], []string{"test2"}, ast.FINAL), nil)
		is(parser.comments.CommentMap.Size(), 1)

		// Arrays pt 9 - Leading node
		parser, program = test(`[/*test2*/a,,,];`, nil)
		is(len(program.Body), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ExpressionStatement).Expression.(*ast.ArrayLiteral).Value[0]], []string{"test2"}, ast.LEADING), nil)
		is(parser.comments.CommentMap.Size(), 1)

		// Object literal
		parser, program = test("obj = {a: 1, b: 2 /*test2*/, c: 3}", nil)
		is(len(program.Body), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ExpressionStatement).Expression.(*ast.AssignExpression).Right.(*ast.ObjectLiteral).Value[1].Value], []string{"test2"}, ast.TRAILING), nil)
		is(parser.comments.CommentMap.Size(), 1)

		// Object literal, pt 2
		parser, program = test("obj = {/*test2*/a: 1, b: 2, c: 3}", nil)
		is(len(program.Body), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ExpressionStatement).Expression.(*ast.AssignExpression).Right.(*ast.ObjectLiteral).Value[0].Value], []string{"test2"}, ast.KEY), nil)
		is(parser.comments.CommentMap.Size(), 1)

		// Object literal, pt 3
		parser, program = test("obj = {x/*test2*/: 1, y: 2, z: 3}", nil)
		is(len(program.Body), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ExpressionStatement).Expression.(*ast.AssignExpression).Right.(*ast.ObjectLiteral).Value[0].Value], []string{"test2"}, ast.COLON), nil)
		is(parser.comments.CommentMap.Size(), 1)

		// Object literal, pt 4
		parser, program = test("obj = {x: /*test2*/1, y: 2, z: 3}", nil)
		is(len(program.Body), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ExpressionStatement).Expression.(*ast.AssignExpression).Right.(*ast.ObjectLiteral).Value[0].Value], []string{"test2"}, ast.LEADING), nil)
		is(parser.comments.CommentMap.Size(), 1)

		// Object literal, pt 5
		parser, program = test("obj = {x: 1/*test2*/, y: 2, z: 3}", nil)
		is(len(program.Body), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ExpressionStatement).Expression.(*ast.AssignExpression).Right.(*ast.ObjectLiteral).Value[0].Value], []string{"test2"}, ast.TRAILING), nil)
		is(parser.comments.CommentMap.Size(), 1)

		// Object literal, pt 6
		parser, program = test("obj = {x: 1, y: 2, z: 3/*test2*/}", nil)
		is(len(program.Body), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ExpressionStatement).Expression.(*ast.AssignExpression).Right.(*ast.ObjectLiteral).Value[2].Value], []string{"test2"}, ast.TRAILING), nil)
		is(parser.comments.CommentMap.Size(), 1)

		// Object literal, pt 7 - trailing comment
		parser, program = test("obj = {x: 1, y: 2, z: 3,/*test2*/}", nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ExpressionStatement).Expression.(*ast.AssignExpression).Right.(*ast.ObjectLiteral)], []string{"test2"}, ast.FINAL), nil)

		// Line breaks
		parser, program = test(`
t1 = "BLA DE VLA"
/*Test*/
t2 = "Nothing happens."
		`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[1]], []string{"Test"}, ast.LEADING), nil)

		// Line breaks pt 2
		parser, program = test(`
t1 = "BLA DE VLA" /*Test*/
t2 = "Nothing happens."
		`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ExpressionStatement).Expression.(*ast.AssignExpression).Right.(*ast.StringLiteral)], []string{"Test"}, ast.TRAILING), nil)
		is(checkComments((parser.comments.CommentMap)[program.Body[1].(*ast.ExpressionStatement)], []string{}, ast.LEADING), nil)

		// Line breaks pt 3
		parser, program = test(`
t1 = "BLA DE VLA" /*Test*/ /*Test2*/
t2 = "Nothing happens."
		`, nil)
		is(parser.comments.CommentMap.Size(), 2)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ExpressionStatement).Expression.(*ast.AssignExpression).Right.(*ast.StringLiteral)], []string{"Test", "Test2"}, ast.TRAILING), nil)
		is(checkComments((parser.comments.CommentMap)[program.Body[1].(*ast.ExpressionStatement)], []string{}, ast.LEADING), nil)

		// Line breaks pt 4
		parser, program = test(`
t1 = "BLA DE VLA" /*Test*/
/*Test2*/
t2 = "Nothing happens."
		`, nil)
		is(parser.comments.CommentMap.Size(), 2)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ExpressionStatement).Expression.(*ast.AssignExpression).Right.(*ast.StringLiteral)], []string{"Test"}, ast.TRAILING), nil)
		is(checkComments((parser.comments.CommentMap)[program.Body[1]], []string{"Test2"}, ast.LEADING), nil)

		// Line breaks pt 5
		parser, program = test(`
t1 = "BLA DE VLA";
/*Test*/
t2 = "Nothing happens."
		`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[1]], []string{"Test"}, ast.LEADING), nil)

		// Line breaks pt 6
		parser, program = test(`
t1 = "BLA DE VLA"; /*Test*/
/*Test2*/
t2 = "Nothing happens."
		`, nil)
		is(parser.comments.CommentMap.Size(), 2)
		is(checkComments((parser.comments.CommentMap)[program.Body[1]], []string{"Test", "Test2"}, ast.LEADING), nil)

		// Misc
		parser, program = test(`
var x = Object.create({y: {
},
// a
});
		`, nil)
		is(parser.comments.CommentMap.Size(), 1)

		// Misc 2
		parser, program = test(`
var x = Object.create({y: {
},
// a
// b
a: 2});
		`, nil)
		is(parser.comments.CommentMap.Size(), 2)

		// Statement blocks
		parser, program = test(`
(function() {
  // Baseline setup
})
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ExpressionStatement).Expression.(*ast.FunctionLiteral).Body], []string{" Baseline setup"}, ast.FINAL), nil)

		// Switches
		parser, program = test(`
switch (switcha) {
  // switch comment
  case "switchb":
  	a
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.SwitchStatement).Body[0]], []string{" switch comment"}, ast.LEADING), nil)

		// Switches pt 2
		parser, program = test(`
switch (switcha) {
  case /*switch comment*/ "switchb":
  	a
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.SwitchStatement).Body[0].Test], []string{"switch comment"}, ast.LEADING), nil)

		// Switches pt 3
		parser, program = test(`
switch (switcha) {
  case "switchb" /*switch comment*/:
  	a
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.SwitchStatement).Body[0].Test], []string{"switch comment"}, ast.TRAILING), nil)

		// Switches pt 4
		parser, program = test(`
switch (switcha) {
  case "switchb": /*switch comment*/
  	a
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.SwitchStatement).Body[0].Consequent[0]], []string{"switch comment"}, ast.LEADING), nil)

		// Switches pt 5 - default
		parser, program = test(`
switch (switcha) {
  default: /*switch comment*/
  	a
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.SwitchStatement).Body[0].Consequent[0]], []string{"switch comment"}, ast.LEADING), nil)

		// Switches pt 6
		parser, program = test(`
switch (switcha) {
  case "switchb":
  	/*switch comment*/a
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.SwitchStatement).Body[0].Consequent[0]], []string{"switch comment"}, ast.LEADING), nil)

		// Switches pt 7
		parser, program = test(`
switch (switcha) {
  case "switchb": /*switch comment*/ {
  	a
  	}
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.SwitchStatement).Body[0].Consequent[0]], []string{"switch comment"}, ast.LEADING), nil)

		// Switches pt 8
		parser, program = test(`
switch (switcha) {
  case "switchb":  {
  	a
  	}/*switch comment*/
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.SwitchStatement).Body[0].Consequent[0]], []string{"switch comment"}, ast.TRAILING), nil)

		// Switches pt 9
		parser, program = test(`
switch (switcha) {
  case "switchb": /*switch comment*/ {
  	a
  	}
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.SwitchStatement).Body[0].Consequent[0]], []string{"switch comment"}, ast.LEADING), nil)

		// Switches pt 10
		parser, program = test(`
switch (switcha) {
  case "switchb": {
  	/*switch comment*/a
  	}
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.SwitchStatement).Body[0].Consequent[0].(*ast.BlockStatement).List[0]], []string{"switch comment"}, ast.LEADING), nil)

		// For loops
		parser, program = test(`
for(/*comment*/i = 0 ; i < 1 ; i++) {
	a
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ForStatement).Initializer.(*ast.SequenceExpression).Sequence[0].(*ast.AssignExpression).Left], []string{"comment"}, ast.LEADING), nil)

		// For loops pt 2
		parser, program = test(`
for(i/*comment*/ = 0 ; i < 1 ; i++) {
	a
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ForStatement).Initializer.(*ast.SequenceExpression).Sequence[0].(*ast.AssignExpression).Left], []string{"comment"}, ast.TRAILING), nil)

		// For loops pt 3
		parser, program = test(`
for(i = 0 ; /*comment*/i < 1 ; i++) {
	a
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ForStatement).Test.(*ast.BinaryExpression).Left], []string{"comment"}, ast.LEADING), nil)

		// For loops pt 4
		parser, program = test(`
for(i = 0 ;i /*comment*/ < 1 ; i++) {
	a
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ForStatement).Test.(*ast.BinaryExpression).Left], []string{"comment"}, ast.TRAILING), nil)

		// For loops pt 5
		parser, program = test(`
for(i = 0 ;i < 1 /*comment*/ ; i++) {
	a
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ForStatement).Test.(*ast.BinaryExpression).Right], []string{"comment"}, ast.TRAILING), nil)

		// For loops pt 6
		parser, program = test(`
for(i = 0 ;i < 1 ; /*comment*/ i++) {
	a
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ForStatement).Update.(*ast.UnaryExpression).Operand], []string{"comment"}, ast.LEADING), nil)

		// For loops pt 7
		parser, program = test(`
for(i = 0 ;i < 1 ; i++) /*comment*/ {
	a
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ForStatement).Body], []string{"comment"}, ast.LEADING), nil)

		// For loops pt 8
		parser, program = test(`
for(i = 0 ;i < 1 ; i++)  {
	a
}/*comment*/
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ForStatement).Body], []string{"comment"}, ast.TRAILING), nil)

		// For loops pt 9
		parser, program = test(`
for(i = 0 ;i < 1 ; /*comment*/i++)  {
	a
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ForStatement).Update.(*ast.UnaryExpression).Operand], []string{"comment"}, ast.LEADING), nil)

		// For loops pt 10
		parser, program = test(`
for(i = 0 ;i < 1 ; i/*comment*/++)  {
	a
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ForStatement).Update.(*ast.UnaryExpression).Operand.(*ast.Identifier)], []string{"comment"}, ast.TRAILING), nil)

		// For loops pt 11
		parser, program = test(`
for(i = 0 ;i < 1 ; i++/*comment*/)  {
	a
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ForStatement).Update.(*ast.UnaryExpression)], []string{"comment"}, ast.TRAILING), nil)

		// ForIn
		parser, program = test(`
for(/*comment*/var i = 0 in obj)  {
	a
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ForInStatement).Into], []string{"comment"}, ast.LEADING), nil)

		// ForIn pt 2
		parser, program = test(`
for(var i = 0 /*comment*/in obj)  {
	a
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ForInStatement).Into.(*ast.VariableExpression).Initializer], []string{"comment"}, ast.TRAILING), nil)

		// ForIn pt 3
		parser, program = test(`
for(var i = 0 in /*comment*/ obj)  {
	a
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ForInStatement).Source], []string{"comment"}, ast.LEADING), nil)

		// ForIn pt 4
		parser, program = test(`
for(var i = 0 in  obj/*comment*/)  {
	a
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ForInStatement).Source], []string{"comment"}, ast.TRAILING), nil)

		// ForIn pt 5
		parser, program = test(`
for(var i = 0 in  obj) /*comment*/ {
	a
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ForInStatement).Body], []string{"comment"}, ast.LEADING), nil)

		// ForIn pt 6
		parser, program = test(`
for(var i = 0 in  obj) {
	a
}/*comment*/
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ForInStatement).Body], []string{"comment"}, ast.TRAILING), nil)

		// ForIn pt 7
		parser, program = test(`
for(var i = 0 in  obj) {
	a
}
// comment
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program], []string{" comment"}, ast.TRAILING), nil)

		// ForIn pt 8
		parser, program = test(`
for(var i = 0 in  obj) {
	a
}
// comment
c
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[1]], []string{" comment"}, ast.LEADING), nil)

		// Block
		parser, program = test(`
		/*comment*/{
			a
		}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.BlockStatement)], []string{"comment"}, ast.LEADING), nil)

		// Block pt 2
		parser, program = test(`
		{
			a
		}/*comment*/
			`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.BlockStatement)], []string{"comment"}, ast.TRAILING), nil)

		// If then else
		parser, program = test(`
/*comment*/
if(a) {
	b
} else {
	c
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.IfStatement)], []string{"comment"}, ast.LEADING), nil)

		// If then else pt 2
		parser, program = test(`
if/*comment*/(a) {
	b
} else {
	c
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.IfStatement)], []string{"comment"}, ast.IF), nil)

		// If then else pt 3
		parser, program = test(`
if(/*comment*/a) {
	b
} else {
	c
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.IfStatement).Test], []string{"comment"}, ast.LEADING), nil)

		// If then else pt 4
		parser, program = test(`
if(a/*comment*/) {
	b
} else {
	c
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.IfStatement).Test], []string{"comment"}, ast.TRAILING), nil)

		// If then else pt 4
		parser, program = test(`
if(a)/*comment*/ {
	b
} else {
	c
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.IfStatement).Consequent], []string{"comment"}, ast.LEADING), nil)

		// If then else pt 5
		parser, program = test(`
if(a) {
	b
} /*comment*/else {
	c
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.IfStatement).Consequent], []string{"comment"}, ast.TRAILING), nil)

		// If then else pt 6
		parser, program = test(`
if(a) {
	b
} else/*comment*/ {
	c
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.IfStatement).Alternate], []string{"comment"}, ast.LEADING), nil)

		// If then else pt 7
		parser, program = test(`
if(a) {
	b
} else {
	c
}/*comment*/
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.IfStatement).Alternate], []string{"comment"}, ast.TRAILING), nil)

		// If then else pt 8
		parser, program = test(`
if
/*comment*/
(a) {
	b
} else {
	c
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)

		// If then else pt 9
		parser, program = test(`
if
(a)
 /*comment*/{
	b
} else {
	c
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)

		// If then else pt 10
		parser, program = test(`
if(a){
	b
}
/*comment*/
else {
	c
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)

		// Do while
		parser, program = test(`
/*comment*/do {
	a
} while(b)
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.DoWhileStatement)], []string{"comment"}, ast.LEADING), nil)

		// Do while pt 2
		parser, program = test(`
do /*comment*/ {
	a
} while(b)
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.DoWhileStatement)], []string{"comment"}, ast.DO), nil)

		// Do while pt 3
		parser, program = test(`
do {
	a
} /*comment*/ while(b)
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.DoWhileStatement).Body], []string{"comment"}, ast.TRAILING), nil)

		// Do while pt 4
		parser, program = test(`
do {
	a
} while/*comment*/(b)
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.DoWhileStatement)], []string{"comment"}, ast.WHILE), nil)

		// Do while pt 5
		parser, program = test(`
do {
	a
} while(b)/*comment*/
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.DoWhileStatement)], []string{"comment"}, ast.TRAILING), nil)

		// While
		parser, program = test(`
/*comment*/while(a) {
	b
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.WhileStatement)], []string{"comment"}, ast.LEADING), nil)

		// While pt 2
		parser, program = test(`
while/*comment*/(a) {
	b
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.WhileStatement)], []string{"comment"}, ast.WHILE), nil)

		// While pt 3
		parser, program = test(`
while(/*comment*/a) {
	b
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.WhileStatement).Test], []string{"comment"}, ast.LEADING), nil)

		// While pt 4
		parser, program = test(`
while(a/*comment*/) {
	b
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.WhileStatement).Test], []string{"comment"}, ast.TRAILING), nil)

		// While pt 5
		parser, program = test(`
while(a) /*comment*/ {
	c
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.WhileStatement).Body], []string{"comment"}, ast.LEADING), nil)

		// While pt 6
		parser, program = test(`
while(a) {
	c
}/*comment*/
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.WhileStatement).Body], []string{"comment"}, ast.TRAILING), nil)

		// While pt 7
		parser, program = test(`
while(a) {
	c/*comment*/
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.WhileStatement).Body.(*ast.BlockStatement).List[0].(*ast.ExpressionStatement).Expression.(*ast.Identifier)], []string{"comment"}, ast.TRAILING), nil)

		// While pt 7
		parser, program = test(`
while(a) {
	/*comment*/
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.WhileStatement).Body.(*ast.BlockStatement)], []string{"comment"}, ast.FINAL), nil)

		// While pt 8
		parser, program = test(`
while
/*comment*/(a) {

}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)

		// While pt 9
		parser, program = test(`
while
(a)
 /*comment*/{

}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)

		// Break
		parser, program = test(`
while(a) {
	break/*comment*/;
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.WhileStatement).Body.(*ast.BlockStatement).List[0].(*ast.BranchStatement)], []string{"comment"}, ast.TRAILING), nil)

		// Break pt 2
		parser, program = test(`
while(a) {
	next/*comment*/:
	break next;
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.WhileStatement).Body.(*ast.BlockStatement).List[0].(*ast.LabelledStatement).Label], []string{"comment"}, ast.TRAILING), nil)

		// Break pt 3
		parser, program = test(`
while(a) {
	next:/*comment*/
	break next;
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.WhileStatement).Body.(*ast.BlockStatement).List[0].(*ast.LabelledStatement)], []string{"comment"}, ast.LEADING), nil)

		// Break pt 4
		parser, program = test(`
while(a) {
	next:
	break /*comment*/next;
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.WhileStatement).Body.(*ast.BlockStatement).List[0].(*ast.LabelledStatement).Statement.(*ast.BranchStatement).Label], []string{"comment"}, ast.LEADING), nil)

		// Break pt 5
		parser, program = test(`
while(a) {
	next:
	break next/*comment*/;
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.WhileStatement).Body.(*ast.BlockStatement).List[0].(*ast.LabelledStatement).Statement.(*ast.BranchStatement).Label], []string{"comment"}, ast.TRAILING), nil)

		// Debugger
		parser, program = test(`
debugger // comment
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.DebuggerStatement)], []string{" comment"}, ast.TRAILING), nil)

		// Debugger pt 2
		parser, program = test(`
debugger; // comment
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program], []string{" comment"}, ast.TRAILING), nil)

		// Debugger pt 3
		parser, program = test(`
debugger;
// comment
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program], []string{" comment"}, ast.TRAILING), nil)

		// With
		parser, program = test(`
/*comment*/with(a) {
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.WithStatement)], []string{"comment"}, ast.LEADING), nil)

		// With pt 2
		parser, program = test(`
with/*comment*/(a) {
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.WithStatement)], []string{"comment"}, ast.WITH), nil)

		// With pt 3
		parser, program = test(`
with(/*comment*/a) {
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.WithStatement).Object], []string{"comment"}, ast.LEADING), nil)

		// With pt 4
		parser, program = test(`
with(a/*comment*/) {
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.WithStatement).Object], []string{"comment"}, ast.TRAILING), nil)

		// With pt 5
		parser, program = test(`
with(a) /*comment*/ {
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.WithStatement).Body], []string{"comment"}, ast.LEADING), nil)

		// With pt 6
		parser, program = test(`
with(a)  {
}/*comment*/
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.WithStatement).Body], []string{"comment"}, ast.TRAILING), nil)

		// With pt 7
		parser, program = test(`
with
/*comment*/(a)  {
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)

		// With pt 8
		parser, program = test(`
with
(a)
  /*comment*/{
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)

		// Var
		parser, program = test(`
/*comment*/var a
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.VariableStatement)], []string{"comment"}, ast.LEADING), nil)

		// Var pt 2
		parser, program = test(`
var/*comment*/ a
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.VariableStatement).List[0]], []string{"comment"}, ast.LEADING), nil)

		// Var pt 3
		parser, program = test(`
var a/*comment*/
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.VariableStatement).List[0]], []string{"comment"}, ast.TRAILING), nil)

		// Var pt 4
		parser, program = test(`
var a/*comment*/, b
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.VariableStatement).List[0].(*ast.VariableExpression)], []string{"comment"}, ast.TRAILING), nil)

		// Var pt 5
		parser, program = test(`
var a, /*comment*/b
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.VariableStatement).List[1].(*ast.VariableExpression)], []string{"comment"}, ast.LEADING), nil)

		// Var pt 6
		parser, program = test(`
var a, b/*comment*/
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.VariableStatement).List[1]], []string{"comment"}, ast.TRAILING), nil)

		// Var pt 7
		parser, program = test(`
var a, b;
/*comment*/
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program], []string{"comment"}, ast.TRAILING), nil)

		// Return
		parser, program = test(`
		function f() {
/*comment*/return o
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)

		// Try catch
		parser, program = test(`
/*comment*/try {
	a
} catch(b) {
	c
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.TryStatement)], []string{"comment"}, ast.LEADING), nil)

		// Try catch pt 2
		parser, program = test(`
try/*comment*/ {
	a
} catch(b) {
	c
} finally {
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.TryStatement).Body], []string{"comment"}, ast.LEADING), nil)

		// Try catch pt 3
		parser, program = test(`
try {
	a
}/*comment*/ catch(b) {
	c
} finally {
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.TryStatement).Body], []string{"comment"}, ast.TRAILING), nil)

		// Try catch pt 4
		parser, program = test(`
try {
	a
} catch(/*comment*/b) {
	c
} finally {
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.TryStatement).Catch.Parameter], []string{"comment"}, ast.LEADING), nil)

		// Try catch pt 5
		parser, program = test(`
try {
	a
} catch(b/*comment*/) {
	c
} finally {
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.TryStatement).Catch.Parameter], []string{"comment"}, ast.TRAILING), nil)

		// Try catch pt 6
		parser, program = test(`
try {
	a
} catch(b) /*comment*/{
	c
} finally {
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.TryStatement).Catch.Body], []string{"comment"}, ast.LEADING), nil)

		// Try catch pt 7
		parser, program = test(`
try {
	a
} catch(b){
	c
} /*comment*/ finally {
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.TryStatement).Catch.Body], []string{"comment"}, ast.TRAILING), nil)

		// Try catch pt 8
		parser, program = test(`
try {
	a
} catch(b){
	c
} finally /*comment*/ {
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.TryStatement).Finally], []string{"comment"}, ast.LEADING), nil)

		// Try catch pt 9
		parser, program = test(`
try {
	a
} catch(b){
	c
} finally  {
}/*comment*/
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.TryStatement).Finally], []string{"comment"}, ast.TRAILING), nil)

		// Try catch pt 11
		parser, program = test(`
try {
	a
}
/*comment*/
 catch(b){
	c
} finally {
 d
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.TryStatement).Body], []string{"comment"}, ast.TRAILING), nil)

		// Throw
		parser, program = test(`
throw a/*comment*/
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ThrowStatement).Argument], []string{"comment"}, ast.TRAILING), nil)

		// Throw pt 2
		parser, program = test(`
/*comment*/throw a
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ThrowStatement)], []string{"comment"}, ast.LEADING), nil)

		// Throw pt 3
		parser, program = test(`
throw /*comment*/a
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ThrowStatement).Argument], []string{"comment"}, ast.LEADING), nil)

		// Try catch pt 10
		parser, program = test(`
try {
	a
} catch(b){
	c
}
 /*comment*/finally  {
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.TryStatement).Catch.Body], []string{"comment"}, ast.TRAILING), nil)

		// Try catch pt 11
		parser, program = test(`
try {
	a
} catch(b){
	c
}
 finally
 /*comment*/
 {
 d
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.TryStatement).Finally], []string{"comment"}, ast.LEADING), nil)

		// Switch / comment
		parser, program = test(`
var volvo = 1
//comment
switch(abra) {
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)

		// Switch / comment
		parser, program = test(`
f("string",{
   key: "val"
   //comment
});
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)

		// Switch / comment
		parser, program = test(`
function f() {
   /*comment*/if(true){a++}
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		n := program.Body[0].(*ast.FunctionStatement).Function.Body.(*ast.BlockStatement).List[0]
		is(checkComments((parser.comments.CommentMap)[n], []string{"comment"}, ast.LEADING), nil)

		// Function in function
		parser, program = test(`
function f() {
   /*comment*/function f2() {
   }
}
	`, nil)
		is(parser.comments.CommentMap.Size(), 1)
		n = program.Body[0].(*ast.FunctionStatement).Function.Body.(*ast.BlockStatement).List[0]
		is(checkComments((parser.comments.CommentMap)[n], []string{"comment"}, ast.LEADING), nil)

		parser, program = test(`
a + /*comment1*/
/*comment2*/
b/*comment3*/;
/*comment4*/c
	`, nil)
		is(parser.comments.CommentMap.Size(), 4)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ExpressionStatement).Expression.(*ast.BinaryExpression).Right], []string{"comment1", "comment2"}, ast.LEADING), nil)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ExpressionStatement).Expression.(*ast.BinaryExpression).Right], []string{"comment3"}, ast.TRAILING), nil)
		is(checkComments((parser.comments.CommentMap)[program.Body[1]], []string{"comment4"}, ast.LEADING), nil)

		parser, program = test(`
a + /*comment1*/
/*comment2*/
b/*comment3*/
/*comment4*/c
	`, nil)
		is(parser.comments.CommentMap.Size(), 4)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ExpressionStatement).Expression.(*ast.BinaryExpression).Right], []string{"comment1", "comment2"}, ast.LEADING), nil)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ExpressionStatement).Expression.(*ast.BinaryExpression).Right], []string{"comment3"}, ast.TRAILING), nil)
		is(checkComments((parser.comments.CommentMap)[program.Body[1]], []string{"comment4"}, ast.LEADING), nil)

		// New
		parser, program = test(`
a = /*comment1*/new /*comment2*/ obj/*comment3*/()
	`, nil)
		is(parser.comments.CommentMap.Size(), 3)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ExpressionStatement).Expression.(*ast.AssignExpression).Right], []string{"comment1"}, ast.LEADING), nil)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ExpressionStatement).Expression.(*ast.AssignExpression).Right.(*ast.NewExpression).Callee], []string{"comment2"}, ast.LEADING), nil)
		is(checkComments((parser.comments.CommentMap)[program.Body[0].(*ast.ExpressionStatement).Expression.(*ast.AssignExpression).Right.(*ast.NewExpression).Callee], []string{"comment3"}, ast.TRAILING), nil)

	})
}

func TestParser_comments2(t *testing.T) {
	tt(t, func() {
		test := func(source string, chk interface{}) (*_parser, *ast.Program) {
			parser, program, err := testParseWithMode(source, StoreComments)
			is(firstErr(err), chk)

			// Check unresolved comments
			is(len(parser.comments.Comments), 0)
			return parser, program
		}

		parser, program := test(`
a = /*comment1*/new /*comment2*/ obj/*comment3*/()
`, nil)
		n := program.Body[0]
		fmt.Printf("FOUND NODE: %v, number of comments: %v\n", reflect.TypeOf(n), len(parser.comments.CommentMap[n]))
		displayComments(parser.comments.CommentMap)

	})
}
