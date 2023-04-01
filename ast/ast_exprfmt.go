package ast

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

type expFormatter struct {
	keyPath    []KeyStep
	fmtKeyPath func(keyPath []KeyStep) string
}

type KeyStep interface {
	String() string
}

type ByName string

func (k ByName) String() string {
	return string(k)
}

type ByIdx int

func (i ByIdx) String() string {
	return strconv.Itoa(int(i))
}

// Format key path a /-delimited hierarchy.
func FmtKeyAsPath(keyPath []KeyStep) string {
	kpc := make([]string, 0, len(keyPath))
	for _, x := range keyPath {
		kpc = append(kpc, strings.ReplaceAll(x.String(), "/", `\/`))
	}
	return "/" + strings.Join(kpc, "/")
}

// Format key path like a normalized jq/Python-esque expression
func FmtKeyAsExpression(keyPath []KeyStep) string {
	kpc := make([]string, 0, len(keyPath))
	for _, kp := range keyPath {
		switch x := kp.(type) {
		case ByIdx:
			kpc = append(kpc, "["+x.String()+"]")
		case ByName:
			kpc = append(kpc, "["+strconv.Quote(x.String())+"]")
		}
	}
	return "." + strings.Join(kpc, "")
}

func (f *expFormatter) fmtNode(n Node) string {
	b := &bytes.Buffer{}
	ensureNewline := func() {
		if buf := b.Bytes(); len(buf) > 0 && buf[len(buf)-1] != '\n' {
			b.WriteString("\n")
		}
	}

	switch tn := n.(type) {
	case *File:
		b.WriteString(f.fmtNode(tn.Root))
		ensureNewline()
	case *Literal:
		b.WriteString(f.fmtKeyPath(f.keyPath))
		b.WriteString(" = ")
		b.Write(tn.Value)
		ensureNewline()
	case *Array:
		if len(tn.Elements) != 0 {
			f.keyPath = append(f.keyPath, nil)
			for i, e := range tn.Elements {
				f.keyPath[len(f.keyPath)-1] = ByIdx(i)
				b.WriteString(f.fmtNode(e.Value))
				ensureNewline()
			}
			f.keyPath = f.keyPath[:len(f.keyPath)-1]
		}
	case *Object:
		if len(tn.Fields) != 0 {
			f.keyPath = append(f.keyPath, nil)
			for _, fl := range tn.Fields {
				v := fl.Name.(*Literal).Value
				s, err := strconv.Unquote(string(v))
				if err != nil {
					panic(fmt.Errorf("unquote err: %s %s", err, v))
				}

				f.keyPath[len(f.keyPath)-1] = ByName(s)
				b.WriteString(f.fmtNode(fl.Value))
				ensureNewline()
			}
			f.keyPath = f.keyPath[:len(f.keyPath)-1]
		}
	case *CommentGroup:
		return ""
	}
	return b.String()
}

type KVOption func(f *expFormatter)

func (KVOption) OptionFmtKeyAsExpression(f *expFormatter) {
	f.fmtKeyPath = FmtKeyAsExpression
}

func OptionKeyFormatter(kf func(keyPath []KeyStep) string) KVOption {
	return func(f *expFormatter) {
		f.fmtKeyPath = FmtKeyAsExpression
	}
}

// Format an AST using key = value notation.
func FmtKeyValue(node Node, options ...KVOption) string {
	f := &expFormatter{fmtKeyPath: FmtKeyAsPath}
	for _, o := range options {
		o(f)
	}
	return f.fmtNode(node)
}
