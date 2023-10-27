package ast

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
)

type File struct {
	Doc     *CommentGroup
	Root    Node // we only have one root element.
	Comment *CommentGroup
}

type Literal struct {
	Type  LiteralType
	Value []byte
}

type Object struct {
	Doc     *CommentGroup
	Fields  []*Field
	Comment *CommentGroup
}

type Field struct {
	Doc     *CommentGroup
	Name    Node
	Value   Node
	Comment *CommentGroup
}

type Element struct {
	Doc     *CommentGroup
	Value   Node
	Comment *CommentGroup
}

type Array struct {
	Elements []*Element
}

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
	Text []byte
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

// Parse a string in JSONR syntax into an AST and return the root node.
func Parse(in []byte) (Node, error) {
	return (&astParser{}).Parse(in)
}

func ParseString(in string) (Node, error) {
	return (&astParser{}).Parse([]byte(in))
}

type astParser struct {
	lex       *lexer
	item      *item
	peekItems []*item
}

func (p *astParser) next() *item {
	if len(p.peekItems) > 0 {
		p.item = p.peekItems[0]
		p.peekItems = p.peekItems[1:]
		return p.item
	}
	p.item = p.lex.yield()
	return p.item
}

// Parse the input string into an AST.  This is only useful when you
// are planning to programmatically manipulate the tree.
func (p *astParser) Parse(input []byte) (Node, error) {
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
	var cl []*Comment
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

var (
	_true  = &Literal{Type: LiteralTrue, Value: []byte("true")}
	_false = &Literal{Type: LiteralFalse, Value: []byte("false")}
	_null  = &Literal{Type: LiteralNull, Value: []byte("null")}
)

func (p *astParser) parseElement() (Node, error) {
	switch p.item.typ {
	case itemString:
		for i, r := range p.item.val {
			if r < '\u001f' {
				return nil, fmt.Errorf("invalid literal character %q at position %d: control characters from \\u0000 - \\u001f must be escaped", r, p.item.start+i)
			}
		}
		return &Literal{Type: LiteralString,
			Value: p.item.val}, nil
	case itemTrue:
		return _true, nil
	case itemFalse:
		return _false, nil
	case itemNull:
		return _null, nil
	case itemNumber:
		return &Literal{Type: LiteralNumber, Value: p.item.val}, nil
	case itemArrayOpen:
		return p.parseArray()
	case itemObjectOpen:
		return p.parseObject()
	case itemError:
		return nil, fmt.Errorf("%v", p.item)
	default:
		return nil, fmt.Errorf("unknown type: %v", p.item.typ)
	}
}

func (p *astParser) parseArray() (Node, error) {
	x := &Array{Elements: make([]*Element, 0, 16)}
	p.next()
	for {
		doc := p.parseCommentGroup()
		switch p.item.typ {
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

			if p.item.typ == itemComma {
				p.next()
			}

			// FIXME(msolo) This is probably a hack. It means a comment
			// after a newline is properly interpreted as not belonging to
			// the current element. We don't support arbitrary whitespace
			// while formatting. That's another kettle of fish at this
			// point.
			if p.item.typ == itemWhitespace && bytes.IndexByte(p.item.val, '\n') >= 0 {
				p.next()
				continue
			}

			// Handle trailing comment regardless of trailing comma.
			// FIXME(msolo) Having [ val /* comment */, ] seems visually confusing but legal.
			e.Comment = p.parseCommentGroup()

		}
	}
}

func (p *astParser) parseObject() (Node, error) {
	x := &Object{Fields: make([]*Field, 0, 16)}
	p.next() // skip {
	for {
		doc := p.parseCommentGroup()
		switch {
		case p.item.typ == itemObjectClose:
			return x, nil
		case p.item.typ == itemString:
			key, err := p.parseElement()
			if err != nil {
				return nil, err
			}

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

			if p.item.typ == itemComma {
				p.next()
			}

			// FIXME(msolo) This is probably a hack. It means a comment
			// after a newline is properly interpreted as not belonging to
			// the current element. We don't support arbitrary whitespace
			// while formatting. That's another kettle of fish at this
			// point.
			if p.item.typ == itemWhitespace && bytes.IndexByte(p.item.val, '\n') >= 0 {
				p.next()
				continue
			}
			// Handle trailing comment regardless of trailing comma.
			// FIXME(msolo) Having val /* comment */, } seems visually
			// confusing but legal.
			f.Comment = p.parseCommentGroup()
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
	indentLevel        int
	skipNextIndent     bool
	skipComments       bool
	elideTrailingComma bool
	sortKeys           bool
	buf                *bytes.Buffer
}

var valueDelimiter = []byte(": ")
var indentDelimiter = []byte("  ")

func (f *formatter) indent() []byte {
	if f.skipNextIndent {
		f.skipNextIndent = false
		return nil
	}
	return bytes.Repeat(indentDelimiter, f.indentLevel)
}

func (f *formatter) fmtNode(n Node) []byte {
	if f.buf == nil {
		f.buf = bytes.NewBuffer(make([]byte, 0, 64))
	}
	b := f.buf

	ensureNewline := func() {
		if buf := b.Bytes(); len(buf) > 0 && buf[len(buf)-1] != '\n' {
			b.WriteByte('\n')
		}
	}

	switch tn := n.(type) {
	case *File:
		f.fmtNode(tn.Doc)
		ensureNewline()
		f.fmtNode(tn.Root)
		ensureNewline()
		f.fmtNode(tn.Comment)
		ensureNewline()
	case *Literal:
		b.Write(f.indent())
		b.Write(tn.Value)
	case *Array:
		b.Write(f.indent())
		b.WriteByte('[')
		if len(tn.Elements) != 0 {
			f.indentLevel++
			b.WriteByte('\n')
			for i, e := range tn.Elements {
				f.fmtNode(e.Doc)
				ensureNewline()
				f.fmtNode(e.Value)
				if f.elideTrailingComma {
					if i != len(tn.Elements)-1 {
						b.WriteByte(',')
					}
				} else {
					b.WriteByte(',')
				}

				if e.Comment != nil {
					b.WriteByte(' ')
					f.skipNextIndent = true
					f.fmtNode(e.Comment)
				}
				ensureNewline()
			}
			f.indentLevel--
			b.Write(f.indent())
		}
		b.WriteByte(']')
	case *Object:
		b.Write(f.indent())
		b.WriteByte('{')
		if len(tn.Fields) != 0 {
			if f.sortKeys {
				sort.Sort(byKey(tn.Fields))
			}

			f.indentLevel++
			b.WriteByte('\n')
			for i, fl := range tn.Fields {
				f.fmtNode(fl.Doc)
				ensureNewline()
				f.fmtNode(fl.Name)
				b.Write(valueDelimiter)
				f.skipNextIndent = true
				f.fmtNode(fl.Value)
				if f.elideTrailingComma {
					if i != len(tn.Fields)-1 {
						b.WriteByte(',')
					}
				} else {
					b.WriteByte(',')
				}
				if fl.Comment != nil {
					b.WriteByte(' ')
					f.skipNextIndent = true
					f.fmtNode(fl.Comment)
				}
				ensureNewline()
			}
			f.indentLevel--
			b.Write(f.indent())
		}
		b.WriteByte('}')
	case *CommentGroup:
		if f.skipComments {
			// Whether or not we process this, reset indent.
			f.skipNextIndent = false
			return nil
		}
		// FIXME(msolo) We return CommentGroup as a typed nil. This is a mess.
		if tn == nil {
			return nil
		}
		for _, c := range tn.List {
			b.Write(f.indent())
			b.Write(c.Text)
			if bytes.HasPrefix(c.Text, commentStart) {
				b.WriteByte('\n')
			}
		}
	}
	return nil
}

type Option func(f *formatter)

func OptionSortKeys(f *formatter) {
	f.sortKeys = true
}

// Format an AST according to JSON rules for backward compatibility.
func FmtJson(node Node, options ...Option) []byte {
	fmt := &formatter{
		skipComments:       true,
		elideTrailingComma: true,
	}
	for _, opt := range options {
		opt(fmt)
	}
	fmt.fmtNode(node)
	return fmt.buf.Bytes()
}

// Format an AST according to some aesthetic heuristics. Thanks gofmt.
func FmtJsonr(node Node, options ...Option) []byte {
	fmt := &formatter{}
	for _, o := range options {
		o(fmt)
	}
	fmt.fmtNode(node)
	return fmt.buf.Bytes()
}

type byKey []*Field

func (a byKey) Len() int      { return len(a) }
func (a byKey) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byKey) Less(i, j int) bool {
	return (bytes.Compare(
		a[i].Name.(*Literal).Value,
		a[j].Name.(*Literal).Value) < 0)
}
