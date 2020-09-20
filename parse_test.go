package jsonr

import "testing"

// func toSlice(t *testing.T, items chan item) []item {
// 	tl := make([]item, 0, 16)
// 	for i := range items {
// 		tl = append(tl, i)
// 	}
// 	t.Logf("tokens: %v", tl)
// 	return tl
// }

// func lexToSlice(t *testing.T, s string) []item {
// 	l := lex("test-lex", s)
// 	items := make([]item, 0, 16)
// 	for {
// 		i := l.yield()
// 		items = append(items, i)
// 		if i.typ == itemEOF || i.typ == itemError {
// 			// if i.typ == itemError {
// 			// 	t.Log(i)
// 			// }
// 			return items
// 		}
// 	}
// }

// func checkTokenVals(t *testing.T, items []item, val ...string) {
// 	// +1 for implicit EOF
// 	if len(items) != len(val)+1 {
// 		t.Fatalf("expected %d tokens: got %d", len(val)+1, len(items))
// 	}
// 	for i, v := range val {
// 		if items[i].val != v {
// 			t.Fatalf("expected %#v: got %#v at token %d", v, items[i].val, i)
// 		}
// 	}
// }

func TestParse(t *testing.T) {
	checkParsedVal := func(input string, expectedVal interface{}) {
		v, err := (&parser{}).parse(input)
		if err != nil {
			t.Fatal(input, err)
		}
		if v != expectedVal {
			t.Fatalf(`expected %#v, got %#v`, expectedVal, v)
		}
	}

	checkParsedArray := func(input string, expectedVals []interface{}) {
		v, err := (&parser{}).parse(input)
		if err != nil {
			t.Fatal(input, err)
		}
		vals := v.([]interface{})
		for i, x := range vals {
			expectedVal := expectedVals[i]
			if x != expectedVal {
				t.Fatalf(`expected %#v, got %#v`, expectedVal, v)
			}
		}
		if len(vals) != len(expectedVals) {
			t.Fatalf(`expected %d vals, got %d`, len(expectedVals), len(vals))
		}
	}

	checkParsedObject := func(input string, expectedVals map[string]interface{}) {
		v, err := (&parser{}).parse(input)
		if err != nil {
			t.Fatalf(`input %#v len:%d err: %s`, input, len(input), err)
		}
		for k, val := range v.(map[string]interface{}) {
			if val != expectedVals[k] {
				t.Fatalf(`expected %#v, got %#v`, expectedVals[k], val)
			}
		}
	}

	checkParsedVal(`"tiptop"`, "tiptop")
	checkParsedVal(`false`, false)
	checkParsedVal(`true`, true)
	checkParsedVal(`null`, nil)
	checkParsedVal(` null`, nil)
	checkParsedVal(` null `, nil)
	checkParsedVal(`-3.1981e4`, -3.1981e4)
	checkParsedArray(`[]`, []interface{}{})
	checkParsedArray(`[ ]`, []interface{}{})
	checkParsedArray(`[null]`, []interface{}{nil})
	checkParsedArray(`[null,]`, []interface{}{nil})
	checkParsedArray(` [ null , ] `, []interface{}{nil})
	checkParsedObject(`{}`, map[string]interface{}{})
	checkParsedObject(` { } `, map[string]interface{}{})
	checkParsedObject("{\n}", map[string]interface{}{})
	checkParsedObject(`{"x":null}`, map[string]interface{}{"x": nil})
	checkParsedObject(`{"x":null,}`, map[string]interface{}{"x": nil})
	checkParsedObject(` { "x" : null } `, map[string]interface{}{"x": nil})
	checkParsedObject(` { "x" : null , } `, map[string]interface{}{"x": nil})

	//FIXME(msolo) check back-to-back docs
	//FIXME(msolo) check whitespace

}
