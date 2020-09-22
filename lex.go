package jsonr

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

//go:generate stringer -type=itemType
type itemType int

type item struct {
	typ   itemType
	val   string
	start int
}

func (i item) String() string {
	switch {
	case i.typ == itemEOF:
		return "EOF"
	case i.typ == itemError:
		return fmt.Sprintf("ERR: start:%d %s", i.start, i.val)
	case len(i.val) > 10:
		return fmt.Sprintf("%.10q...", i.val)
	}
	return fmt.Sprintf("%q", i.val)
}

// Simple non-concurrent fifo
type fifo struct {
	items []item
}

func (f *fifo) Get() item {
	i := f.items[0]
	copy(f.items, f.items[1:])
	f.items = f.items[:len(f.items)-1]
	return i
}

func (f *fifo) Put(i item) {
	f.items = append(f.items, i)
}

func (f *fifo) Len() int {
	return len(f.items)
}

type stateFn func(l *lexer) stateFn

type lexer struct {
	name  string // used only for error reports.
	input string // the string being scanned.
	start int    // start position of this item.
	pos   int    // current position in the input.
	width int    // width of last rune read from input.
	state stateFn
	items *fifo
}

func lex(name, input string) *lexer {
	l := &lexer{
		name:  name,
		input: input,
		state: lexStream,
		items: &fifo{make([]item, 0, 16)},
	}
	return l
}

// Run for a bit until an item has been produced.
// Returns itemEOF when there is no more input to be consumed.
func (l *lexer) yield() (i item) {
	defer func() {
		if x := recover(); x != nil {
			l.state = nil
			i = x.(item)
		}
	}()
	if l.items.Len() > 0 {
		return l.items.Get()
	}
	for l.state != nil {
		l.state = l.state(l)
		if l.items.Len() > 0 {
			return l.items.Get()
		}
	}
	return item{typ: itemEOF}
}

// emit passes an item back to the client.
func (l *lexer) emit(t itemType) {
	v := l.input[l.start:l.pos]
	//	fmt.Printf("emit %s %#v %d:%d\n", t, v, l.start, l.pos)
	i := item{t, v, l.start}
	l.items.Put(i)
	l.start = l.pos
}

func (l *lexer) ignore() {
	l.start = l.pos
}

const eof = -1

// next returns the next rune in the input.
func (l *lexer) next() (r rune) {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return r
}

// nextByte returns the next byte in the input.
func (l *lexer) nextByte() (b byte, ok bool) {
	if l.pos >= len(l.input) {
		l.width = 0
		return '\000', false
	}
	b, l.width = l.input[l.pos], 1
	l.pos += l.width
	return b, ok
}

// backup steps back one rune.
// Can be called only once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width
}

// peek returns but does not consume
// the next rune in the input.
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// accept consumes the next rune
// if it's from the valid set.
func (l *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

// acceptRun consumes a run of runes from the valid set.
func (l *lexer) acceptRun(valid string) (accepted bool) {
	for strings.IndexRune(valid, l.next()) >= 0 {
		accepted = true
	}
	l.backup()
	return accepted
}

// error returns an error token and terminates the scan
// by passing back a nil pointer that will be the next
// state, terminating l.run.
func (l *lexer) errorf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if msg == "" {
		panic("empty error")
	}
	i := item{
		itemError,
		msg,
		l.start,
	}
	l.items.Put(i)
	panic(i)
}

func hasPrefixByte(s string, b byte) bool {
	if len(s) == 0 {
		return false
	}
	return s[0] == b
}
