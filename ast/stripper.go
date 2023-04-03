package ast

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type stripper struct {
	lex  *lexer
	item *item
	buf  *bytes.Buffer
}

func (p *stripper) next() *item {
	p.item = p.lex.yield()
	//fmt.Println("item", p.item)
	return p.item
}

// Parse a JSON string. Objects will be map[string]interface{}, arrays
// []interface{} and numbers will be float64 for now.
func (p *stripper) Strip(input []byte) ([]byte, error) {
	p.lex = lex("parse-lexer", input)
	for {
		i := p.next()
		if i.typ == itemEOF {
			return p.buf.Bytes(), nil
		}
		p.emitter(i.typ, i.val, i.start)
	}
}

func (p *stripper) skipWhitespaceOrComment() {
	for {
		switch p.item.typ {
		case itemWhitespace, itemComment:
			p.next()
		default:
			return
		}
	}
}

func (p *stripper) emitter(t itemType, val []byte, start int) {
	switch t {
	case itemString:
		p.buf.Write(val)
	case itemTrue:
		p.buf.WriteString("true")
	case itemFalse:
		p.buf.WriteString("false")
	case itemNull:
		p.buf.WriteString("null")
	case itemNumber:
		p.buf.Write(val)
	case itemArrayOpen:
		p.buf.WriteString("[")
	case itemObjectOpen:
		p.buf.WriteString("{")
	case itemWhitespace, itemComment:
	case itemArrayClose:
		p.buf.WriteString("]")
	case itemComma:
		p.buf.WriteString(",")
	case itemObjectClose:
		p.buf.WriteString("}")
	case itemColon:
		p.buf.WriteString(":")
		// case itemError:
		// 	return fmt.Errorf("parse err: %#v", p.item.val)
		// default:
		// 	return fmt.Errorf("unknown type: %v", p.item.typ)

	}
}

func (p *stripper) parseElement() error {
	switch p.item.typ {
	case itemString:
		p.buf.Write(p.item.val)
	case itemTrue:
		p.buf.WriteString("true")
	case itemFalse:
		p.buf.WriteString("false")
	case itemNull:
		p.buf.WriteString("null")
	case itemNumber:
		p.buf.Write(p.item.val)
	case itemArrayOpen:
		p.buf.WriteString("[")
		return p.parseArray()
	case itemObjectOpen:
		p.buf.WriteString("{")
		return p.parseObject()
	case itemError:
		return fmt.Errorf("parse err: %#v", p.item.val)
	default:
		return fmt.Errorf("unknown type: %v", p.item.typ)
	}
	return nil
}

func (p *stripper) parseArray() error {
	for {
		p.next()
		switch p.item.typ {
		case itemWhitespace, itemComment:
			continue
		case itemArrayClose:
			p.buf.WriteString("]")
			return nil
		case itemComma:
			p.buf.WriteString(",")
		case itemEOF:
			return fmt.Errorf("unexpected EOF reading array")
		default:
			err := p.parseElement()
			if err != nil {
				return err
			}
		}
	}
}

func (p *stripper) parseObject() error {
	for {
		p.next()
		switch p.item.typ {
		case itemWhitespace:
			continue
		case itemComma:
			p.buf.WriteString(",")
		case itemObjectClose:
			p.buf.WriteString("}")
			return nil
		case itemString:
			key := p.item.val
			p.next()
			p.skipWhitespaceOrComment()

			if p.item.typ != itemColon {
				return fmt.Errorf("expected colon delimiter for key token")
			}

			p.next()
			p.skipWhitespaceOrComment()

			p.buf.Write(key)
			p.buf.WriteString(":")
			err := p.parseElement()
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("invalid key token %v", p.item)
		}
	}
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
