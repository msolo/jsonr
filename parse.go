package jsonr

import (
	"fmt"
	"strconv"
)

type parser struct {
	lex       *lexer
	item      item
	peekItems []item
}

func (p *parser) next() item {
	if len(p.peekItems) > 0 {
		p.item = p.peekItems[0]
		p.peekItems = p.peekItems[1:]
		return p.item
	}
	p.item = p.lex.yield()
	//fmt.Println("item", p.item)
	return p.item
}

func (p *parser) peek() item {
	i := p.lex.yield()
	p.peekItems = append(p.peekItems, i)
	return i
}

func (p *parser) parse(input string) (interface{}, error) {
	p.lex = lex("parse-lexer", input)
	i := p.next()
	if i.typ == itemWhitespace {
		i = p.next()
	}
	return p.parseElement()
}

func (p *parser) parseElement() (interface{}, error) {
	switch p.item.typ {
	case itemString:
		return p.item.val[1 : len(p.item.val)-1], nil
	case itemTrue:
		return true, nil
	case itemFalse:
		return false, nil
	case itemNull:
		return nil, nil
	case itemNumber:
		// FIXME(msolo) Doesn't have to be a float if I'm reading the spec
		// correctly, but perhaps this a concession to informal
		// compatibility?
		x, err := strconv.ParseFloat(p.item.val, 64)
		return x, err
	case itemArrayOpen:
		return p.parseArray()
	case itemObjectOpen:
		return p.parseObject()
	case itemError:
		return nil, fmt.Errorf(p.item.val)
	default:
		return nil, fmt.Errorf("unknown type: %v", p.item.typ)
	}
	return nil, nil
}

func (p *parser) parseArray() (interface{}, error) {
	x := make([]interface{}, 0, 16)
	for {
		i := p.next()
	onLast:
		switch i.typ {
		case itemWhitespace:
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
			i = p.next()
			if i.typ == itemWhitespace {
				i = p.next()
			}
			if i.typ != itemComma {
				goto onLast
			}
		}
	}
}

func (p *parser) parseObject() (interface{}, error) {
	x := make(map[string]interface{}, 16)
	for {
		i := p.next()
	onLast:
		switch {
		case i.typ == itemWhitespace:
			continue
		case i.typ == itemObjectClose:
			return x, nil
		case i.typ == itemString:
			key := i.val[1 : len(i.val)-1]
			i = p.next()
			if i.typ == itemWhitespace {
				i = p.next()
			}
			if i.typ != itemColon {
				return nil, fmt.Errorf("expected colon delimiter for key token")
			}
			i = p.next()
			if i.typ == itemWhitespace {
				i = p.next()
			}
			val, err := p.parseElement()
			if err != nil {
				return nil, err
			}
			x[key] = val
			i = p.next()
			if i.typ == itemWhitespace {
				i = p.next()
			}
			if i.typ != itemComma {
				goto onLast
			}
		default:
			return nil, fmt.Errorf("invalid key token %v", i)
		}
	}
}
