package jsonr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
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
	return p.item
}

func (p *astParser) peek() item {
	i := p.lex.yield()
	p.peekItems = append(p.peekItems, i)
	return i
}

// Parse the input string into an AST.  This is only useful when you
// are planning to programmatically manipulate the tree.
func (p *astParser) Parse(input string) (Node, error) {
	p.lex = lex("ast-parse-lexer", input)
	p.next()
	doc := p.parseCommentGroup()
	elt, err := p.parseElement()
	if err != nil {
		return nil, err
	}
	p.next()
	comment := p.parseCommentGroup()
	return &File{Doc: doc, Root: elt, Comment: comment}, nil
}

func (p *astParser) parseCommentGroup() *CommentGroup {
	cl := make([]*Comment, 0, 4)
	for {
		if p.item.typ == itemWhitespace {
			p.next()
			continue
		}
		if p.item.typ == itemComment {
			cl = append(cl, &Comment{p.item.val})
			p.next()
			continue
		}
		if len(cl) > 0 {
			return &CommentGroup{cl}
		}
		return nil
	}
}

func (p *astParser) parseElement() (Node, error) {
	switch p.item.typ {
	case itemString:
		return &Literal{Type: LiteralString,
			Value: p.item.val}, nil
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
		return nil, fmt.Errorf("itemError: %#v", p.item.val)
	default:
		return nil, fmt.Errorf("unknown type: %v", p.item.typ)
	}
	return nil, nil
}

func (p *astParser) parseArray() (Node, error) {
	x := &Array{Elements: make([]*Element, 0, 16)}
	for {
		p.next()
		doc := p.parseCommentGroup()
	onLast:
		switch p.item.typ {
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

			e := &Element{Doc: doc, Value: y}
			x.Elements = append(x.Elements, e)
			p.next()
			if p.item.typ == itemWhitespace {
				p.next()
			}
			//			fmt.Println("check post white", i)
			if p.item.typ != itemComma {
				goto onLast
			}
			// Handle trailing comment after the comma.
			// FIXME(msolo) Having [ val /* comment */, ] seems visually confusing but legal.

			p.next()
			e.Comment = p.parseCommentGroup()
			goto onLast
		}
	}
}

func (p *astParser) parseObject() (Node, error) {
	x := &Object{Fields: make([]*Field, 0, 16)}
	for {
		p.next()
		doc := p.parseCommentGroup()
	onLast:
		switch {
		case p.item.typ == itemWhitespace:
			continue
		case p.item.typ == itemObjectClose:
			return x, nil
		case p.item.typ == itemString:
			key, err := p.parseElement()
			p.next()
			if p.item.typ == itemWhitespace {
				p.next()
			}
			if p.item.typ != itemColon {
				return nil, fmt.Errorf("expected colon delimiter for key token")
			}
			p.next()
			if p.item.typ == itemWhitespace {
				p.next()
			}
			val, err := p.parseElement()
			if err != nil {
				return nil, err
			}
			f := &Field{Doc: doc, Name: key, Value: val}
			x.Fields = append(x.Fields, f)
			p.next()
			if p.item.typ == itemWhitespace {
				p.next()
			}
			if p.item.typ != itemComma {
				goto onLast
			}
			// Handle trailing comment after the comma.
			// FIXME(msolo) Having val /* comment */, ] seems visually confusing but legal.
			p.next()
			f.Comment = p.parseCommentGroup()
			goto onLast
		default:
			return nil, fmt.Errorf("invalid key token %v", p.item)
		}
	}
}

func prettyFmt(data interface{}) string {
	var p []byte
	// Oh, the irony.
	p, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(p)
}

type formatter struct {
	indentLevel    int
	skipNextIndent bool
}

func (f *formatter) fmtNode(n Node) string {
	// FIXME(msolo) WTF, (nil,nil) strikes again? a typed nil is coerced
	// to Node and we no longer get equivalence?
	if n == nil {
		return ""
	}
	b := &bytes.Buffer{}

	indent := func() {
		if f.skipNextIndent {
			f.skipNextIndent = false
			return
		}
		b.WriteString(strings.Repeat("  ", f.indentLevel))
	}

	switch tn := n.(type) {
	case *File:
		b.WriteString(f.fmtNode(tn.Doc))
		b.WriteString(f.fmtNode(tn.Root))
		b.WriteString(f.fmtNode(tn.Comment))
		b.WriteString("\n")
	case *Literal:
		indent()
		b.WriteString(tn.Value)
	case *Array:
		indent()
		b.WriteString("[")
		if len(tn.Elements) != 0 {
			f.indentLevel++
			b.WriteString("\n")
			for _, e := range tn.Elements {
				b.WriteString(f.fmtNode(e.Doc))
				b.WriteString(f.fmtNode(e.Value))
				b.WriteString(",")
				if e.Comment != nil {
					b.WriteString(" ")
					f.skipNextIndent = true
					b.WriteString(f.fmtNode(e.Comment))
				}
				if !bytes.HasSuffix(b.Bytes(), []byte("\n")) {
					b.WriteString("\n")
				}
			}
			f.indentLevel--
			indent()
		}
		b.WriteString("]")
	case *Object:
		indent()
		b.WriteString("{")
		if len(tn.Fields) != 0 {
			f.indentLevel++
			b.WriteString("\n")
			for _, fl := range tn.Fields {
				b.WriteString(f.fmtNode(fl.Doc))
				b.WriteString(f.fmtNode(fl.Name))
				b.WriteString(": ")
				f.skipNextIndent = true
				b.WriteString(f.fmtNode(fl.Value))
				b.WriteString(",")
				if fl.Comment != nil {
					b.WriteString(" ")
					f.skipNextIndent = true
					b.WriteString(f.fmtNode(fl.Comment))
				}
				if !bytes.HasSuffix(b.Bytes(), []byte("\n")) {
					b.WriteString("\n")
				}
			}
			f.indentLevel--
			indent()
		}
		b.WriteString("}")
	case *CommentGroup:
		// FIXME(msolo) Surely this isn't necessary? See above.
		if tn == nil {
			return ""
		}
		for _, c := range tn.List {
			indent()
			b.WriteString(c.Text)
			if strings.HasPrefix(c.Text, "//") {
				b.WriteString("\n")
			}
		}
	}
	return b.String()
}

// Format an AST according to some heauristics. Thanks gofmt.
func JsonFmt(node Node) string {
	return (&formatter{}).fmtNode(node)
}
