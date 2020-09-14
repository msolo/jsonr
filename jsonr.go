// jsonr provides an analogous API to the standard json package, but
// allows json data to contain comments // ...  or /* ... */ as well
// as superfluous yet convenient commas. These functions strip
// comments and allow JSON parsing to proceed as expected using the
// standard json package.
package jsonr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// See json.Unmarsal.
func Unmarshal(data []byte, v interface{}) error {
	return fmt.Errorf("FIXME")
}

type jsonrReader struct {
	r   io.Reader
	buf *bytes.Buffer
}

func (jr *jsonrReader) Read(b []byte) (n int, err error) {
	return 0, fmt.Errorf("FIXME")

}

// FIXME(msolo) This strips a whole buffer at a time rather than reading incrementally from the underlying reader. No one should confuse JSONR for something high performance, but we needed waste too many resources.
//
// See json.NewDecoder.
func NewDecoder(r io.Reader) *json.Decoder {
	jr := &jsonrReader{r: r}
	return json.NewDecoder(jr)
}
