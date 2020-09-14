package jsonr

import "testing"

func toSlice(t *testing.T, items chan item) []item {
	tl := make([]item, 0, 16)
	for i := range items {
		tl = append(tl, i)
	}
	t.Logf("tokens: %v", tl)
	return tl
}

func lexToSlice(t *testing.T, s string) []item {
	l := lex("test-lex", s)
	items := make([]item, 0, 16)
	for {
		i := l.yield()
		items = append(items, i)
		if i.typ == itemEOF || i.typ == itemError {
			// if i.typ == itemError {
			// 	t.Log(i)
			// }
			return items
		}
	}
}

func checkItem(t *testing.T, i item, val string) {
	if i.val != val {
		t.Fatalf("expected %#v: got %#v", val, i.val)
	}
}

func checkTokenVals(t *testing.T, items []item, val ...string) {
	// +1 for implicit EOF
	if len(items) != len(val)+1 {
		t.Fatalf("expected %d tokens: got %d", len(val)+1, len(items))
	}
	for i, v := range val {
		if items[i].val != v {
			t.Fatalf("expected %#v: got %#v at token %d", v, items[i].val, i)
		}
	}
}

func TestElement(t *testing.T) {
	tl := lexToSlice(t, `"tiptop"`)
	checkItem(t, tl[0], `"tiptop"`)

	tl = lexToSlice(t, `"tiptop1""tiptop2"`)
	checkItem(t, tl[0], `"tiptop1"`)
	checkItem(t, tl[1], `"tiptop2"`)

	tl = lexToSlice(t, `"tiptop1" "tiptop2"`)
	checkItem(t, tl[0], `"tiptop1"`)
	checkItem(t, tl[2], `"tiptop2"`)

	tl = lexToSlice(t, `null`)
	checkItem(t, tl[0], `null`)

	tl = lexToSlice(t, `true`)
	checkItem(t, tl[0], `true`)

	tl = lexToSlice(t, `false`)
	checkItem(t, tl[0], `false`)

	// FIXME(msolo) This might not be valid, it's not totally clear. jq
	// can't parse it, that's for sure.
	tl = lexToSlice(t, `nullnull`)
	checkItem(t, tl[0], `null`)
	checkItem(t, tl[1], `null`)
}

func TestElementNull(t *testing.T) {
	tl := lexToSlice(t, `null`)
	checkItem(t, tl[0], `null`)
}

func TestElementNumber(t *testing.T) {
	tl := lexToSlice(t, `0`)
	checkItem(t, tl[0], `0`)
	tl = lexToSlice(t, `1.1`)
	checkItem(t, tl[0], `1.1`)
	tl = lexToSlice(t, `-1.1`)
	checkItem(t, tl[0], `-1.1`)
	tl = lexToSlice(t, `1.1e01`)
	checkItem(t, tl[0], `1.1e01`)
	tl = lexToSlice(t, `+1.1e01`)
	checkItem(t, tl[len(tl)-1], `malformed integer number`)
}

func TestEmptyArray(t *testing.T) {
	tl := lexToSlice(t, `[]`)
	checkItem(t, tl[0], `[`)
	checkItem(t, tl[1], `]`)

	tl = lexToSlice(t, `[ ]`)
	checkItem(t, tl[0], `[`)
	checkItem(t, tl[2], `]`)
}

func TestArray(t *testing.T) {
	tl := lexToSlice(t, `[null]`)
	checkItem(t, tl[0], `[`)
	checkItem(t, tl[2], `]`)

	tl = lexToSlice(t, `["1", "2"]`)
	checkTokenVals(t, tl, `[`, `"1"`, `,`, ` `, `"2"`, `]`)
}

func TestEmptyObject(t *testing.T) {
	tl := lexToSlice(t, `{}`)
	checkTokenVals(t, tl, `{`, `}`)

	tl = lexToSlice(t, `{ }`)
	checkItem(t, tl[0], `{`)
	checkItem(t, tl[2], `}`)
}

func TestObject(t *testing.T) {
	tl := lexToSlice(t, `{"a":null}`)
	checkTokenVals(t, tl, `{`, `"a"`, `:`, `null`, `}`)

	tl = lexToSlice(t, `{"a":null,"b":null}`)
	checkTokenVals(t, tl, `{`, `"a"`, `:`, `null`, `,`, `"b"`, `:`, `null`, `}`)
}

// func TestLineComment(t *testing.T) {
// 	tl := lexToSlice(t, `{//}\n`)
// 	checkItem(t, tl[0], `{`)
// }

// func TestRangeComment(t *testing.T) {
// 	tl := lexToSlice(t, `{/**/}`)
// 	checkItem(t, tl[1], `/**/`)
// }

// func TestRangeCommentInString(t *testing.T) {
// 	tl := lexToSlice(t, `{"/**/"}`)
// 	checkItem(t, tl[1], `"/**/"`)
// }

// func TestNestedQuoteInString(t *testing.T) {
// 	tl := lexToSlice(t, `{"\""}`)
// 	checkItem(t, tl[1], `"\""`)
// }

// func TestNoCommentTerminator(t *testing.T) {
// 	tl := lexToSlice(t, `{/*}`)
// 	if tl[len(tl)-1].typ != itemError {
// 		t.Error("expected a parsing error - no comment terminator")
// 	}
// }
