package ast

import (
	"reflect"
	"testing"

	"github.com/ianbruene/go-difflib/difflib"
)

func TestAstParse(t *testing.T) {
	checkParsedVal := func(input string, expectedVal interface{}) {
		v, err := (&astParser{}).Parse([]byte(input))
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
		output := string(FmtJsonr(v))
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
				Value: []byte(`"tiptop"`),
			},
		},
	)

	checkParsedVal(`"\"quoted\""
`,
		&File{
			Root: &Literal{
				Type:  LiteralString,
				Value: []byte(`"\"quoted\""`),
			},
		},
	)

	checkParsedVal(`true
`,
		&File{
			Root: &Literal{
				Type:  LiteralTrue,
				Value: []byte("true"),
			},
		},
	)

	checkParsedVal(`false
`,
		&File{
			Root: &Literal{
				Type:  LiteralFalse,
				Value: []byte("false"),
			},
		},
	)

	checkParsedVal(`null
`,
		&File{
			Root: &Literal{
				Type:  LiteralNull,
				Value: []byte("null"),
			},
		},
	)

	checkParsedVal(`-3.1981e4
`,
		&File{
			Root: &Literal{
				Type:  LiteralNumber,
				Value: []byte("-3.1981e4"),
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
					{
						Value: &Literal{
							Type:  LiteralNull,
							Value: []byte("null"),
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
				Fields: []*Field{},
			},
		},
	)

	checkParsedVal(`{
  "x": null,
}
`,
		&File{
			Root: &Object{
				Fields: []*Field{
					{
						Name: &Literal{
							Type:  LiteralString,
							Value: []byte(`"x"`),
						},
						Value: &Literal{
							Type:  LiteralNull,
							Value: []byte("null"),
						},
					},
				},
			},
		},
	)

	checkParsedVal(`{
  "quoted\"x": null,
}
`,
		&File{
			Root: &Object{
				Fields: []*Field{
					{
						Name: &Literal{
							Type:  LiteralString,
							Value: []byte(`"quoted\"x"`),
						},
						Value: &Literal{
							Type:  LiteralNull,
							Value: []byte("null"),
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
				Fields: []*Field{
					{
						Name: &Literal{
							Type:  LiteralString,
							Value: []byte(`"x"`),
						},
						Value: &Object{
							Fields: []*Field{
								{
									Name: &Literal{
										Type:  LiteralString,
										Value: []byte(`"nested"`),
									},
									Value: &Literal{
										Type:  LiteralNull,
										Value: []byte("null"),
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
  "x": null,
  // Doc1
  //
  // Doc2
  "y": null, // Trailer.
}
/* postamble */
`,
		&File{
			Doc: &CommentGroup{
				[]*Comment{
					{
						Text: []byte("// Leading doc comment."),
					},
				},
			},
			Comment: &CommentGroup{
				[]*Comment{
					{
						Text: []byte("/* postamble */"),
					},
				},
			},
			Root: &Object{
				Fields: []*Field{
					{
						Name: &Literal{
							Type:  LiteralString,
							Value: []byte(`"x"`),
						},
						Value: &Literal{
							Type:  LiteralNull,
							Value: []byte("null"),
						},
					},
					{
						Doc: &CommentGroup{
							[]*Comment{
								{
									Text: []byte("// Doc1"),
								},
								{
									Text: []byte("//"),
								},
								{
									Text: []byte("// Doc2"),
								},
							},
						},
						Comment: &CommentGroup{
							[]*Comment{
								{
									Text: []byte("// Trailer."),
								},
							},
						},
						Name: &Literal{
							Type:  LiteralString,
							Value: []byte(`"y"`),
						},
						Value: &Literal{
							Type:  LiteralNull,
							Value: []byte("null"),
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

func TestDumpPathEscaping(t *testing.T) {
	s := `{
		"a/b": [0,1]
	}`

	root, err := ParseString(s)
	if err != nil {
		t.Error(err)
	}
	out := FmtKeyValue(root)
	expected := "/a\\/b/0 = 0\n/a\\/b/1 = 1\n"
	if out != expected {
		t.Errorf("expected %s; got %s", expected, out)
	}
}

func TestDumpExprEscaping(t *testing.T) {
	s := `{
		"a\"b": [0,1]
	}`

	root, err := ParseString(s)
	if err != nil {
		t.Error(err)
	}

	out := FmtKeyValue(root, OptionKeyFormatter(FmtKeyAsExpression))
	expected := `.["a\"b"][0] = 0
.["a\"b"][1] = 1
`
	if out != expected {
		t.Errorf("expected %s; got %s", expected, out)
	}
}
