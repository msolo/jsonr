package ast

import (
	"bytes"
	"fmt"
	"strings"
)

type expFormatter struct {
	keyPath []string
}

func fmtKeyPath(keyPath []string) string {
	for i, x := range keyPath {
		keyPath[i] = strings.ReplaceAll(x, "/", `\/`)
	}
	return "/" + strings.Join(keyPath, "/")
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
		b.WriteString(fmtKeyPath(f.keyPath))
		b.WriteString(" = ")
		if tn.Value[0] == '"' && tn.Value[len(tn.Value)-1] == '"' {
			b.WriteString(fmt.Sprintf("%q", tn.Value[1:len(tn.Value)-1]))
		} else {
			b.WriteString(tn.Value)
		}
		ensureNewline()
	case *Array:
		if len(tn.Elements) != 0 {
			f.keyPath = append(f.keyPath, "")
			for i, e := range tn.Elements {
				f.keyPath[len(f.keyPath)-1] = fmt.Sprintf("%d", i)
				b.WriteString(f.fmtNode(e.Value))
				ensureNewline()
			}
			f.keyPath = f.keyPath[:len(f.keyPath)-1]
		}
	case *Object:
		if len(tn.Fields) != 0 {
			f.keyPath = append(f.keyPath, "")
			for _, fl := range tn.Fields {
				v := fl.Name.(*Literal).Value
				f.keyPath[len(f.keyPath)-1] = v[1 : len(v)-1]
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

// Format an AST using key notation.
func JsonLineFmt(node Node) string {
	return (&expFormatter{}).fmtNode(node)
}
