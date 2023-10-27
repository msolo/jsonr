package ast

import (
	"bytes"
	"testing"
)

func lexToSlice(t *testing.T, s string) []*item {
	l := lex("test-lex", []byte(s))
	items := make([]*item, 0, 16)
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

func checkItem(t *testing.T, i *item, val string) {
	if !bytes.Equal(i.val, []byte(val)) {
		t.Fatalf("expected %s: got %s", val, i.val)
	}
}

func checkItemHasPrefix(t *testing.T, i *item, val string) {
	if !bytes.HasPrefix(i.val, []byte(val)) {
		t.Fatalf("expected %s: got %s", val, i.val)
	}
}

func checkTokenVals(t *testing.T, items []*item, val ...string) {
	// +1 for implicit EOF
	if len(items) != len(val)+1 {
		t.Fatalf("expected %d tokens: got %d", len(val)+1, len(items))
	}
	for i, v := range val {
		if !bytes.Equal(items[i].val, []byte(v)) {
			t.Fatalf("expected %#v: got %s at token %d", v, items[i].val, i)
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

	tl = lexToSlice(t, `nul`)
	checkItem(t, tl[0], `failed parsing null`)

	tl = lexToSlice(t, `treu`)
	checkItem(t, tl[0], `failed parsing true`)

	tl = lexToSlice(t, `fals`)
	checkItem(t, tl[0], `failed parsing false`)

	// FIXME(msolo) This might not be valid, it's not totally clear. jq
	// can't parse it, that's for sure.
	tl = lexToSlice(t, `nullnull`)
	checkItem(t, tl[0], `null`)
	checkItem(t, tl[1], `null`)
}

func TestElementString(t *testing.T) {
	tl := lexToSlice(t, `"1\t2"`)
	checkItem(t, tl[0], `"1\t2"`)
	tl = lexToSlice(t, `"1\x2"`)
	checkItem(t, tl[0], `invalid escaped character`)
	tl = lexToSlice(t, `"1\u000"`)
	checkItem(t, tl[0], `invalid unicode escape sequence`)
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
	tl = lexToSlice(t, `1.1E01`)
	checkItem(t, tl[0], `1.1E01`)
	tl = lexToSlice(t, `1.1e-1`)
	checkItem(t, tl[0], `1.1e-1`)
}

func TestElementInvalidNumber(t *testing.T) {
	tl := lexToSlice(t, `+1.1e01`)
	checkItemHasPrefix(t, tl[len(tl)-1], `malformed number`)
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

func TestLineComment(t *testing.T) {
	tl := lexToSlice(t, `{
//}
}
`)
	checkItem(t, tl[0], `{`)
}

func TestFieldComment(t *testing.T) {
	tl := lexToSlice(t, `{
  "x": null,
  // 1
  // 2
  // 3
  "y": null,
}
`)
	checkItem(t, tl[8], `// 1`)
	checkItem(t, tl[10], `// 2`)
	checkItem(t, tl[12], `// 3`)
}

func TestRangeComment(t *testing.T) {
	tl := lexToSlice(t, `{/**/}`)
	checkItem(t, tl[1], `/**/`)
}

func TestMultilineRangeComment(t *testing.T) {
	tl := lexToSlice(t, `{/*
*/}`)
	checkItem(t, tl[1], `/*
*/`)
}

func TestRangeCommentInString(t *testing.T) {
	tl := lexToSlice(t, `"/**/"`)
	checkItem(t, tl[0], `"/**/"`)
}

func TestNestedQuoteInString(t *testing.T) {
	tl := lexToSlice(t, `"\""`)
	checkItem(t, tl[0], `"\""`)
}

func TestNoCommentTerminator(t *testing.T) {
	tl := lexToSlice(t, `{/*}`)
	if tl[len(tl)-1].typ != itemError {
		t.Error("expected a parsing error - no comment terminator")
	}
}

func TestCommentedObject(t *testing.T) {
	tl := lexToSlice(t, `// c1
{
  // c2
  "a": null, // c3
} /* c4 */
`)
	tlm := make([]*item, 0, len(tl))
	// Let's just check the "meaningful" tokens for now.
	for _, i := range tl {
		if i.typ != itemWhitespace {
			tlm = append(tlm, i)
		}
	}
	checkTokenVals(t, tlm, `// c1`, `{`, `// c2`, `"a"`, `:`, `null`, `,`, `// c3`, `}`, `/* c4 */`)
}
