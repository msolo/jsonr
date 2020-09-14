// jsonr provides an analogous API to the standard json package, but
// allows json data to contain comments // ...  or /* ... */ as well
// as superfluous yet convenient commas. These functions strip
// comments and allow JSON parsing to proceed as expected using the
// standard json package.
package jsonr

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
)

// See json.Unmarsal.
func Unmarshal(data []byte, v interface{}) error {
	buf, err := StripComments(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(buf, v)
}

type jsonrReader struct {
	r   io.Reader
	buf *bytes.Buffer
}

func (jr *jsonrReader) Read(b []byte) (n int, err error) {
	if jr.buf == nil {
		in, err := ioutil.ReadAll(jr.r)
		if err != nil {
			return 0, err
		}
		stripped, err := StripComments(in)
		if err != nil {
			return 0, err
		}
		jr.buf = bytes.NewBuffer(stripped)
	}
	return jr.buf.Read(b)
}

// FIXME(msolo) This strips a whole buffer at a time rather than reading incrementally from the underlying reader. No one should confuse JSONR for something high performance, but we needed waste too many resources.
//
// See json.NewDecoder.
func NewDecoder(r io.Reader) *json.Decoder {
	jr := &jsonrReader{r: r}
	return json.NewDecoder(jr)
}
