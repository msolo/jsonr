package jsonr

import (
	"reflect"
	"testing"

	"github.com/pmezard/go-difflib/difflib"
)

func TestAstParse(t *testing.T) {
	checkParsedVal := func(input string, expectedVal interface{}) {
		v, err := (&astParser{}).parse(input)
		if err != nil {
			t.Fatal(input, err)
		}
		if !reflect.DeepEqual(v, expectedVal) {
			diff := difflib.UnifiedDiff{
				A:        difflib.SplitLines(prettyFmt(expectedVal)),
				B:        difflib.SplitLines(prettyFmt(v)),
				FromFile: "Original",
				ToFile:   "Current",
				Context:  3,
			}
			text, err := difflib.GetUnifiedDiffString(diff)
			if err != nil {
				text = err.Error()
			}

			t.Fatalf("expected: %s\ngot %s", prettyFmt(expectedVal), text)
			//t.Fatalf(`expected %#v, got %#v`, expectedVal, v)
		}
	}

	// checkAst := func(input string, expectedVal Node) {
	// 	v, err := (&astParser{}).parse(input)
	// 	if err != nil {
	// 		t.Fatal(input, err)
	// 	}
	// 	nodes := make([]Node, 0, 32)
	// 	expectedNodes := make([]Node, 0, 32)

	// 	Inspect(expectedVal, func(n Node) bool {
	// 		expectedNodes = append(expectedNodes, n)
	// 		return true
	// 	})

	// 	Inspect(v, func(n Node) bool {
	// 		nodes = append(nodes, n)
	// 		return true
	// 	})

	// 	if len(nodes) != len(expectedNodes) {
	// 		t.Fatalf(`expected %d nodes, got %d`, len(expectedNodes), len(nodes))
	// 	}
	// 	for i := len(nodes) - 1; i >= 0; i-- {
	// 		if !reflect.DeepEqual(nodes[i], expectedNodes[i]) {
	// 			t.Fatalf(`expected %s:\ngot %s`, prettyFmt(expectedNodes[i]), prettyFmt(nodes[i]))
	// 		}
	// 	}
	// }

	checkParsedVal(`"tiptop"`,
		&File{
			Root: &Literal{
				Type:  LiteralString,
				Value: "tiptop",
			},
		},
	)

	checkParsedVal(`true`,
		&File{
			Root: &Literal{
				Type:  LiteralTrue,
				Value: "true",
			},
		},
	)

	checkParsedVal(`false`,
		&File{
			Root: &Literal{
				Type:  LiteralFalse,
				Value: "false",
			},
		},
	)

	checkParsedVal(`null`,
		&File{
			Root: &Literal{
				Type:  LiteralNull,
				Value: "null",
			},
		},
	)

	checkParsedVal(`-3.1981e4`,
		&File{
			Root: &Literal{
				Type:  LiteralNumber,
				Value: "-3.1981e4",
			},
		},
	)

	checkParsedVal(`[]`,
		&File{
			Root: &Array{
				[]*Element{},
			},
		},
	)

	checkParsedVal(`[null]`,
		&File{
			Root: &Array{
				[]*Element{
					&Element{
						Value: &Literal{
							Type:  LiteralNull,
							Value: "null",
						},
					},
				},
			},
		},
	)

	checkParsedVal(`{}`,
		&File{
			Root: &Object{
				[]*Field{},
			},
		},
	)

	checkParsedVal(`{"x":null}`,
		&File{
			Root: &Object{
				[]*Field{
					&Field{
						Name: &Literal{
							Type:  LiteralString,
							Value: "x",
						},
						Value: &Literal{
							Type:  LiteralNull,
							Value: "null",
						},
					},
				},
			},
		},
	)

	checkParsedVal(`{
  "x": null
}`,
		&File{
			Root: &Object{
				[]*Field{
					&Field{
						Name: &Literal{
							Type:  LiteralString,
							Value: "x",
						},
						Value: &Literal{
							Type:  LiteralNull,
							Value: "null",
						},
					},
				},
			},
		},
	)

	// checkParsedVal(` null`, nil)
	// checkParsedVal(` null `, nil)
	// checkParsedArray(`[ ]`, []interface{}{})
	// checkParsedArray(`[null,]`, []interface{}{nil})
	// checkParsedArray(` [ null , ] `, []interface{}{nil})
	// checkParsedObject(` { } `, map[string]interface{}{})
	// checkParsedObject("{\n}", map[string]interface{}{})
	// checkParsedObject(`{"x":null,}`, map[string]interface{}{"x": nil})
	// checkParsedObject(` { "x" : null } `, map[string]interface{}{"x": nil})
	// checkParsedObject(` { "x" : null , } `, map[string]interface{}{"x": nil})
}
