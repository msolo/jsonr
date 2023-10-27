package ast

import (
	"fmt"
	"strconv"
)

type parser struct {
	lex       *lexer
	item      *item
	peekItems []*item
}

func (p *parser) next() *item {
	if len(p.peekItems) > 0 {
		p.item = p.peekItems[0]
		p.peekItems = p.peekItems[1:]
		return p.item
	}
	p.item = p.lex.yield()
	//fmt.Println("item", p.item)
	return p.item
}

func (p *parser) peek() *item {
	i := p.lex.yield()
	p.peekItems = append(p.peekItems, i)
	return i
}

// Parse a JSON string. Objects will be map[string]interface{}, arrays
// []interface{} and numbers will be float64 for now.
func (p *parser) parse(input []byte) (interface{}, error) {
	p.lex = lex("parse-lexer", input)
	p.next()
	p.skipWhitespaceOrComment()
	return p.parseElement()
}

func (p *parser) skipWhitespaceOrComment() {
	for {
		switch p.item.typ {
		case itemWhitespace, itemComment:
			p.next()
		default:
			return
		}
	}
}

func (p *parser) parseElement() (interface{}, error) {
	switch p.item.typ {
	case itemString:
		return string(p.item.val[1 : len(p.item.val)-1]), nil
	case itemTrue:
		return true, nil
	case itemFalse:
		return false, nil
	case itemNull:
		return nil, nil
	case itemNumber:
		// FIXME(msolo) Doesn't have to be a float if I'm reading the spec
		// correctly, but perhaps this a concession to informal
		// compatibility with the scourge that is Javascript?
		// FIXME(msolo) string copy
		x, err := strconv.ParseFloat(string(p.item.val), 64)
		return x, err
	case itemArrayOpen:
		return p.parseArray()
	case itemObjectOpen:
		return p.parseObject()
	case itemError:
		return nil, fmt.Errorf("parse err: %#v", p.item.val)
	default:
		return nil, fmt.Errorf("unknown type: %v", p.item.typ)
	}
}

func (p *parser) parseArray() (interface{}, error) {
	x := make([]interface{}, 0, 16)
	for {
		p.next()
	onLast:
		switch p.item.typ {
		case itemWhitespace, itemComment:
			continue
		case itemArrayClose:
			return x, nil
		case itemEOF:
			return nil, fmt.Errorf("unexpected EOF reading array")
		default:
			y, err := p.parseElement()
			if err != nil {
				return nil, err
			}
			x = append(x, y)
			p.next()
			p.skipWhitespaceOrComment()
			if p.item.typ != itemComma {
				goto onLast
			}
		}
	}
}

func (p *parser) parseObject() (interface{}, error) {
	x := make(map[string]interface{}, 16)
	for {
		p.next()
	onLast:
		switch p.item.typ {
		case itemWhitespace:
			continue
		case itemObjectClose:
			return x, nil
		case itemString:
			key := p.item.val
			key = key[1 : len(key)-1]
			p.next()
			p.skipWhitespaceOrComment()

			if p.item.typ != itemColon {
				return nil, fmt.Errorf("expected colon delimiter for key token")
			}

			p.next()
			p.skipWhitespaceOrComment()

			val, err := p.parseElement()
			if err != nil {
				return nil, err
			}
			x[string(key)] = val

			p.next()
			p.skipWhitespaceOrComment()

			if p.item.typ != itemComma {
				goto onLast
			}
		default:
			return nil, fmt.Errorf("invalid key token %v", p.item)
		}
	}
}

// Unmarshal a JSON string. Objects will be map[string]interface{}, arrays
// []interface{} and numbers will be float64 for now.
// jsonr.Unmarshal is faster and more useful, this is mostly for comparison.
func JsonUnmarshal(in []byte) (interface{}, error) {
	return (&parser{}).parse(in)
}
