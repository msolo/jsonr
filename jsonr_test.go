package jsonr

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/msolo/jsonr/ast"
)

type testDoc struct {
	x string
}

var (
	vaguelyRealistic = []byte(`/*
Preamble with fanfare.
*/

{
  // Line comment.
  "x": "a string", // Trailing comment.
  "y": 1.0,
  "z": null,
  // "a": "value temporarily removed for debugging or idle curiosity",
  "array": [],
  "dict": {},  // Now we can have a trailing comma here!
}

// Postamble.
`)

	benchChunk = []byte(`{
  "x": "a string",
  "y": 1.0,
  "z": null,
  "array": [],
  "dict": {}
}
`)
)

func TestJson(t *testing.T) {
	td := &testDoc{}
	if err := Unmarshal([]byte(`{"x":"a string"}`), td); err != nil {
		t.Error(err)
	}
	if err := Unmarshal([]byte(`{/* comment */"x":"a string"}`), td); err != nil {
		t.Error(err)
	}
	if err := Unmarshal([]byte(vaguelyRealistic), td); err != nil {
		t.Error(err)
	}

	type testDoc2 struct {
		X     string         // `json:"x"`
		Y     float32        // `json:"y"`
		Z     int            // `json:"z"` NOTE: null is cast to int(0) which is very un-Go-like
		Array []string       // `json:"array"`
		Dict  map[string]int // `json:"dict"`

	}

	td2 := &testDoc2{}
	dec := NewDecoder(bytes.NewBuffer(vaguelyRealistic))
	dec.DisallowUnknownFields()
	if err := dec.Decode(td2); err != nil {
		t.Error(err)
	}
}
func BenchmarkJSONUnmarshalEmptyStruct(b *testing.B) {
	in := benchChunk
	// I think this causes simple parsing without assinging/allocating any values.
	// It's not necessarily representative - but hey, benchmark!
	out := &struct{}{}
	for i := 0; i < b.N; i++ {
		err := json.Unmarshal(in, out)
		if err != nil {
			b.Errorf("benchmark err: %s", err)
		}
	}
}

// FIXME(msolo) Benchmark confirms that stripping comments is 2.5x
// more expensive than simple JSON parsing. I would accept 2x given
// the naive approach of scanning twice. The original channel approach
// was 9x more expensive. The current overhead is likely due to UTF8
// parsing and readable code style rather than just having a simple
// merged loop over bytes. Of course, it is still absolutely
// "fast-enough" for almost any application I had planned, but best to
// know the costs.
func BenchmarkJSONRUnmarshalEmptyStruct(b *testing.B) {
	in := benchChunk
	out := &struct{}{}
	for i := 0; i < b.N; i++ {
		err := Unmarshal(in, out)
		if err != nil {
			b.Errorf("benchmark err: %s", err)
		}
	}
}

func BenchmarkJSONUnmarshalMap(b *testing.B) {
	in := benchChunk
	out := make(map[string]interface{})
	for i := 0; i < b.N; i++ {
		err := json.Unmarshal(in, &out)
		if err != nil {
			b.Errorf("benchmark err: %s", err)
		}
	}
}

func BenchmarkJSONRUnmarshalMap(b *testing.B) {
	in := benchChunk
	out := make(map[string]interface{})
	for i := 0; i < b.N; i++ {
		err := Unmarshal(in, &out)
		if err != nil {
			b.Errorf("benchmark err: %s", err)
		}
	}
}

func BenchmarkJSONRAstUnmarshalMap(b *testing.B) {
	in := benchChunk
	for i := 0; i < b.N; i++ {
		_, err := ast.JsonUnmarshal(in)
		if err != nil {
			b.Errorf("benchmark err: %s", err)
		}
	}
}

func BenchmarkStrip(b *testing.B) {
	in := benchChunk
	for i := 0; i < b.N; i++ {
		_, err := ast.Strip(in)
		if err != nil {
			b.Errorf("benchmark err: %s", err)
		}
	}
}

func BenchmarkStripReader(b *testing.B) {
	in := benchChunk
	br := bytes.NewReader(in)
	for i := 0; i < b.N; i++ {
		br.Reset(in)
		_, err := ast.StripReader(br)
		if err != nil {
			b.Errorf("benchmark err: %s", err)
		}
	}
}

// func BenchmarkAst2FastStripReader(b *testing.B) {
// 	in := benchChunk
// 	br := bytes.NewReader(in)
// 	for i := 0; i < b.N; i++ {
// 		br.Reset(in)
// 		_, err := ast2.FastStripReader(br)
// 		if err != nil {
// 			b.Errorf("benchmark err: %s", err)
// 		}
// 	}
// }

func TestStripRealistic(t *testing.T) {
	in := vaguelyRealistic
	out, err := ast.Strip(in)
	if err != nil {
		t.Errorf("err: %s", err)
	}
	t.Logf("stripped: %s", string(out))
}

func BenchmarkJSONRRealistic(b *testing.B) {
	in := vaguelyRealistic
	out := make(map[string]interface{})
	for i := 0; i < b.N; i++ {
		err := Unmarshal(in, &out)
		if err != nil {
			b.Errorf("benchmark err: %s", err)
		}
	}
}
