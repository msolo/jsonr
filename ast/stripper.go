package ast

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
)

type stripper struct {
	lex      *lexer
	buf      *bytes.Buffer
	prevItem *item
}

// Parse a JSON string. Objects will be map[string]interface{}, arrays
// []interface{} and numbers will be float64 for now.
func (p *stripper) Strip(input []byte) ([]byte, error) {
	p.lex = lex("parse-lexer", input)
	p.lex.emitter = p.emitter
	for {
		i := p.lex.yield()
		if i.typ == itemEOF {
			return p.buf.Bytes(), nil
		}
	}
}

func (p *stripper) emitter(t itemType, val []byte, start int) {
	if p.prevItem == nil {
		p.prevItem = &item{t, val, start}
		return
	}

	if t == itemWhitespace || t == itemComment {
		return
	}

	switch p.prevItem.typ {
	case itemString:
		p.buf.Write(p.prevItem.val)
	case itemTrue:
		p.buf.WriteString("true")
	case itemFalse:
		p.buf.WriteString("false")
	case itemNull:
		p.buf.WriteString("null")
	case itemNumber:
		p.buf.Write(p.prevItem.val)
	case itemArrayOpen:
		p.buf.WriteString("[")
	case itemObjectOpen:
		p.buf.WriteString("{")
	case itemWhitespace, itemComment:
	case itemArrayClose:
		p.buf.WriteString("]")
	case itemComma:
		if t != itemArrayClose && t != itemObjectClose {
			p.buf.WriteString(",")
		}
	case itemObjectClose:
		p.buf.WriteString("}")
	case itemColon:
		p.buf.WriteString(":")
	case itemEOF:
		return
	case itemError:
		panic(fmt.Errorf("parse err: %#v", p.prevItem.val))
	default:
		panic(fmt.Errorf("programmer unknown type: %v", p.prevItem.typ))

	}

	p.prevItem.typ = t
	p.prevItem.val = val
	p.prevItem.start = start
}

func FastStrip(in []byte) ([]byte, error) {
	return (&stripper{buf: bytes.NewBuffer(make([]byte, 0, 256))}).Strip(in)
}

func JsonUnmarshalFast(in []byte, x interface{}) error {
	b, err := (&stripper{buf: bytes.NewBuffer(make([]byte, 0, 256))}).Strip(in)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, x)
}

type stripReader struct {
	r   io.Reader
	buf *bytes.Buffer
}

func (jr *stripReader) Read(b []byte) (n int, err error) {
	if jr.buf == nil {
		in, err := ioutil.ReadAll(jr.r)
		if err != nil {
			return 0, err
		}
		stripped, err := FastStrip(in)
		if err != nil {
			return 0, err
		}
		jr.buf = bytes.NewBuffer(stripped)
	}
	return jr.buf.Read(b)
}

func NewDecoder(r io.Reader) *json.Decoder {
	jr := &stripReader{r: r}
	return json.NewDecoder(jr)
}
