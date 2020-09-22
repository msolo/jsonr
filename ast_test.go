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
			t.Fatalf("parse failed: %#v, err: %T %s \n", input, err, err)
		}
		if !reflect.DeepEqual(v, expectedVal) {
			prettyExpected := prettyFmt(expectedVal)
			prettyGot := prettyFmt(v)
			diff := difflib.UnifiedDiff{
				A:        difflib.SplitLines(prettyExpected),
				B:        difflib.SplitLines(prettyGot),
				FromFile: "expected",
				ToFile:   "got",
				Context:  3,
			}
			diffText, err := difflib.GetUnifiedDiffString(diff)
			if err != nil {
				diffText = err.Error()
			}

			t.Fatalf("expected: %s\ngot: %s\ndiff:\n%s", prettyExpected, prettyGot, diffText)
		}

		// Can we rewrite the same code we had as input?
		output := jsonFmt(v)
		if input != output {
			diff := difflib.UnifiedDiff{
				A:        difflib.SplitLines(input),
				B:        difflib.SplitLines(output),
				FromFile: "expected source",
				ToFile:   "generated source",
				Context:  3,
			}
			diffText, err := difflib.GetUnifiedDiffString(diff)
			if err != nil {
				diffText = err.Error()
			}

			t.Fatalf("expected formatted source: %s\ngot: %s\ndiff:\n%s", input, output, diffText)
		}
	}

	checkParsedVal(`"tiptop"
`,
		&File{
			Root: &Literal{
				Type:  LiteralString,
				Value: `"tiptop"`,
			},
		},
	)

	checkParsedVal(`true
`,
		&File{
			Root: &Literal{
				Type:  LiteralTrue,
				Value: "true",
			},
		},
	)

	checkParsedVal(`false
`,
		&File{
			Root: &Literal{
				Type:  LiteralFalse,
				Value: "false",
			},
		},
	)

	checkParsedVal(`null
`,
		&File{
			Root: &Literal{
				Type:  LiteralNull,
				Value: "null",
			},
		},
	)

	checkParsedVal(`-3.1981e4
`,
		&File{
			Root: &Literal{
				Type:  LiteralNumber,
				Value: "-3.1981e4",
			},
		},
	)

	checkParsedVal(`[]
`,
		&File{
			Root: &Array{
				[]*Element{},
			},
		},
	)

	checkParsedVal(`[
  null,
]
`,
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

	checkParsedVal(`{}
`,
		&File{
			Root: &Object{
				[]*Field{},
			},
		},
	)

	checkParsedVal(`{
  "x": null,
}
`,
		&File{
			Root: &Object{
				[]*Field{
					&Field{
						Name: &Literal{
							Type:  LiteralString,
							Value: `"x"`,
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
  "x": {
    "nested": null,
  },
}
`,
		&File{
			Root: &Object{
				[]*Field{
					&Field{
						Name: &Literal{
							Type:  LiteralString,
							Value: `"x"`,
						},
						Value: &Object{
							[]*Field{
								&Field{
									Name: &Literal{
										Type:  LiteralString,
										Value: `"nested"`,
									},
									Value: &Literal{
										Type:  LiteralNull,
										Value: "null",
									},
								},
							},
						},
					},
				},
			},
		},
	)

	checkParsedVal(`// Leading doc comment.
{
  // Doc1
  // Doc2
  "x": null, // Trailer.
}/* postamble */
`,
		&File{
			Doc: &CommentGroup{
				[]*Comment{
					&Comment{
						Text: "// Leading doc comment.",
					},
				},
			},
			Comment: &CommentGroup{
				[]*Comment{
					&Comment{
						Text: "/* postamble */",
					},
				},
			},
			Root: &Object{
				[]*Field{
					&Field{
						Doc: &CommentGroup{
							[]*Comment{
								&Comment{
									Text: "// Doc1",
								},
								&Comment{
									Text: "// Doc2",
								},
							},
						},
						Comment: &CommentGroup{
							[]*Comment{
								&Comment{
									Text: "// Trailer.",
								},
							},
						},
						Name: &Literal{
							Type:  LiteralString,
							Value: `"x"`,
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
