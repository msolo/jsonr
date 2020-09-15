package jsonr

import (
	"fmt"
	"strconv"
)

type parser struct {
	lex  *lexer
	item item
}

func (p *parser) next() item {
	p.item = p.lex.yield()
	//fmt.Println("item", p.item)
	return p.item
}

func (p *parser) parse(input string) (interface{}, error) {
	p.lex = lex("parse-lexer", input)
	p.next()
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
		if i.typ == itemObjectClose {
			return x, nil
		}
		if i.typ != itemString {
			return nil, fmt.Errorf("invalid key token %v", i)
		}
		key := i.val[1 : len(i.val)-1]
		i = p.next()
		if i.typ != itemColon {
			return nil, fmt.Errorf("expected colon delimter for key token")
		}
		i = p.next()
		val, err := p.parseElement()
		if err != nil {
			return nil, err
		}
		x[key] = val
		i = p.next()
		if i.typ != itemComma {
			goto onLast
		}
	}
}
