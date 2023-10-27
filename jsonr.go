// jsonr provides an analogous API to the standard json package, but
// allows json data to contain comments // ...  or /* ... */ as well
// as superfluous yet convenient commas. These functions strip
// comments and allow JSON parsing to proceed as expected using the
// standard json package.
package jsonr

import (
	"encoding/json"
	"io"

	"github.com/msolo/jsonr/ast"
)

// See json.Unmarshal.
func Unmarshal(data []byte, v interface{}) error {
	js, err := ast.Strip(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(js, v)
}

// FIXME(msolo) This strips a whole buffer at a time rather than
// reading incrementally from the underlying reader. No one should
// confuse JSONR for something high performance, but we need not waste
// too many resources.
//
// See json.NewDecoder.
func NewDecoder(r io.Reader) *json.Decoder {
	return ast.NewDecoder(r)
}

// Return a JSON-compatible string from a JSONR source string.
// This removes comments, normalizes trailing commas and generally
// pretty-prints.
func Convert2Json(in []byte) ([]byte, error) {
	tree, err := ast.Parse(in)
	if err != nil {
		return nil, err
	}
	return ast.FmtJson(tree), nil
}
