package jsonr

import (
	"encoding/json"
	"fmt"
)

type File struct {
	Doc     *CommentGroup
	Root    Node // we only have one root element.
	Comment *CommentGroup
}

type Literal struct {
	Type  LiteralType
	Value string
}

type Object struct {
	Fields []*Field
}

type Field struct {
	Doc     *CommentGroup
	Name    Node
	Value   Node
	Comment *CommentGroup
}

type Element struct {
	Doc     *CommentGroup
	Type    ElementType
	Value   Node
	Comment *CommentGroup
}

type Array struct {
	Elements []*Element
}

type ElementType int

type LiteralType int

const (
	LiteralNull LiteralType = iota
	LiteralTrue
	LiteralFalse
	LiteralString
	LiteralNumber
)

type CommentGroup struct {
	List []*Comment // len(List) > 0
}

type Comment struct {
	Text string
}

type Node interface{}

type Visitor interface {
	Visit(node Node) (w Visitor)
}

func Walk(v Visitor, node Node) {
	//fmt.Printf("walk %#v\n", node)
	if w := v.Visit(node); w == nil {
		return
	}

	switch n := node.(type) {
	case *File:
		Walk(v, n.Root)
	case *Literal:
	case *Object:
		for _, f := range n.Fields {
			Walk(v, f)
		}
	case *Array:
		for _, e := range n.Elements {
			Walk(v, e)
		}
	case *Field:
		Walk(v, n.Name)
		Walk(v, n.Value)
	case *Element:
		Walk(v, n.Value)
	case *Comment:
	case *CommentGroup:
		for _, c := range n.List {
			Walk(v, c)
		}
	default:
		fmt.Printf("Error %#v\n", n)
	}
}

type inspector func(Node) bool

func (f inspector) Visit(node Node) Visitor {
	if f(node) {
		return f
	}
	return nil
}

func Inspect(node Node, f func(Node) bool) {
	Walk(inspector(f), node)
}

func prettyFmt(data interface{}) string {
	var p []byte
	p, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(p)
}

// func Print(node Node) {
// 	Inspect(node, func(n Node) bool {
// 		switch tn := n.(type) {
// 		case *File:
// 		case *Literal:
// 		}
// 		return true
// 	})
// }

type astParser struct {
	lex       *lexer
	item      item
	peekItems []item
}

func (p *astParser) next() item {
	if len(p.peekItems) > 0 {
		p.item = p.peekItems[0]
		p.peekItems = p.peekItems[1:]
		return p.item
	}
	p.item = p.lex.yield()
	//fmt.Println("item", p.item)
	return p.item
}

func (p *astParser) peek() item {
	i := p.lex.yield()
	p.peekItems = append(p.peekItems, i)
	return i
}

func (p *astParser) parse(input string) (Node, error) {
	p.lex = lex("ast-parse-lexer", input)
	i := p.next()
	if i.typ == itemWhitespace {
		i = p.next()
	}
	elt, err := p.parseElement()
	if err != nil {
		return nil, err
	}
	return &File{Root: elt}, nil
}

func (p *astParser) parseElement() (Node, error) {
	switch p.item.typ {
	case itemString:
		return &Literal{Type: LiteralString,
			Value: p.item.val[1 : len(p.item.val)-1]}, nil
	case itemTrue:
		return &Literal{Type: LiteralTrue, Value: "true"}, nil
	case itemFalse:
		return &Literal{Type: LiteralFalse, Value: "false"}, nil
	case itemNull:
		return &Literal{Type: LiteralNull, Value: "null"}, nil
	case itemNumber:
		return &Literal{Type: LiteralNumber, Value: p.item.val}, nil
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

func (p *astParser) parseArray() (Node, error) {
	x := &Array{Elements: make([]*Element, 0, 16)}
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

			x.Elements = append(x.Elements, &Element{Value: y})
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

func (p *astParser) parseObject() (Node, error) {
	x := &Object{Fields: make([]*Field, 0, 16)}
	for {
		i := p.next()
	onLast:
		switch {
		case i.typ == itemWhitespace:
			continue
		case i.typ == itemObjectClose:
			return x, nil
		case i.typ == itemString:
			key, err := p.parseElement()
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
			x.Fields = append(x.Fields, &Field{Name: key, Value: val})
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
