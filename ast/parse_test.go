package ast

import "testing"

func TestParse(t *testing.T) {
	checkParsedVal := func(input string, expectedVal interface{}) {
		v, err := (&parser{}).parse([]byte(input))
		if err != nil {
			t.Fatal(input, err)
		}
		if v != expectedVal {
			t.Fatalf(`expected %T %#v, got %T, %v, %s`, expectedVal, expectedVal, v, v, v)
		}
	}

	checkParsedArray := func(input string, expectedVals []interface{}) {
		v, err := (&parser{}).parse([]byte(input))
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
		v, err := (&parser{}).parse([]byte(input))
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
	checkParsedObject(`/*comment*/ {"x":null}`, map[string]interface{}{"x": nil})
	checkParsedObject(`{"x":null,}`, map[string]interface{}{"x": nil})
	checkParsedObject(` { "x" : null } `, map[string]interface{}{"x": nil})
	checkParsedObject(` { "x" : null , } `, map[string]interface{}{"x": nil})

	//FIXME(msolo) check back-to-back docs

}
