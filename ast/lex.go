package ast

import (
	"container/list"
	"fmt"
)

//go:generate stringer -type=itemType
type itemType int

type item struct {
	typ   itemType
	val   []byte
	start int
}

func (i item) String() string {
	switch {
	case i.typ == itemEOF:
		return "EOF"
	case i.typ == itemError:
		return fmt.Sprintf("ERR: position:%d %s", i.start, i.val)
	case len(i.val) > 10:
		return fmt.Sprintf("%.10q...", i.val)
	}
	return fmt.Sprintf("%q", i.val)
}

type fifo struct {
	deque *list.List
}

func (f *fifo) Get() *item {
	e := f.deque.Front()
	f.deque.Remove(e)
	return e.Value.(*item)
}

func (f *fifo) Put(i *item) {
	f.deque.PushBack(i)
}

func (f *fifo) Len() int {
	return f.deque.Len()
}

type lexer struct {
	name     string // used only for error reports.
	input    []byte // the data being scanned.
	start    int    // start position of this item.
	pos      int    // current position in the input.
	lastRead int    // size of last read from input.

	// FIXME(msolo) This fifo is essentially an unnecessary buffer. It is used
	// when an emitter function is not set. Rewriting the parser as an emitter
	// callback would remove the need for this.
	items   *fifo
	emitter func(t itemType, val []byte, start int)
}

func lex(name string, input []byte) *lexer {
	l := &lexer{
		name:  name,
		input: input,
		items: &fifo{list.New()},
	}
	return l
}

// Run for a bit until an item has been produced.
// Returns itemEOF when there is no more input to be consumed.
func (l *lexer) yield() (i *item) {
	defer func() {
		if x := recover(); x != nil {
			_i, ok := x.(*item)
			if ok {
				i = _i
			} else {
				panic(x)
			}

		}
	}()
	if l.items.Len() > 0 {
		return l.items.Get()
	}
	lexStream(l)
	if l.items.Len() > 0 {
		return l.items.Get()
	}
	return &item{typ: itemEOF}
}

// emit passes an item back to the client.
func (l *lexer) emit(t itemType) {
	v := l.input[l.start:l.pos]
	//	fmt.Printf("emit %s %#v %d:%d\n", t, v, l.start, l.pos)
	if l.emitter != nil {
		l.emitter(t, v, l.start)
	} else {
		i := &item{t, v, l.start}
		l.items.Put(i)
	}
	l.start = l.pos
}

func (l *lexer) ignore() {
	l.start = l.pos
}

// next returns the next rune in the input.
func (l *lexer) next() (b byte, eof bool) {
	if l.pos >= len(l.input) {
		l.lastRead = 0
		return b, true
	}
	l.lastRead = 1
	b = l.input[l.pos]
	l.pos += l.lastRead
	return b, false
}

// backup steps back one rune.
// Can be called only once per call of next.
func (l *lexer) backup() {
	l.pos -= l.lastRead
}

// peek returns but does not consume
// the next rune in the input.
func (l *lexer) peek() byte {
	r, _ := l.next()
	l.backup()
	return r
}

// accept consumes the next byte
// if it's from the valid set.
func (l *lexer) accept(valid []byte) bool {
	n, _ := l.next()
	for _, b := range valid {
		if n == b {
			return true
		}
	}
	l.backup()
	return false
}

// accept consumes the next byte
// if it's from the valid set.
func (l *lexer) acceptByte(valid byte) bool {
	n, _ := l.next()
	if n == valid {
		return true
	}
	l.backup()
	return false
}

// acceptRun consumes a run of runes from the valid set.
func (l *lexer) acceptRun(valid []byte) (accepted bool) {
	for l.accept(valid) {
		accepted = true
	}
	return accepted
}

func (l *lexer) acceptInt() (accepted bool) {
	for {
		n, _ := l.next()
		if '0' <= n && n <= '9' {
			accepted = true
		} else {
			l.backup()
			return accepted
		}
	}
}
func (l *lexer) acceptWhitespace() (accepted bool) {
	for {
		n, _ := l.next()
		switch n {
		case ' ', '\t', '\n', '\r':
			accepted = true
		default:
			l.backup()
			return accepted
		}
	}
}
func (l *lexer) acceptHex() (accepted bool) {
	for {
		n, _ := l.next()
		switch {
		case '0' <= n && n <= '9':
			accepted = true
		case 'a' <= n && n <= 'f':
			accepted = true
		case 'A' <= n && n <= 'F':
			accepted = true
		default:
			l.backup()
			return accepted
		}
	}
}

// error panics an error token and terminates the scan
func (l *lexer) errorf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if msg == "" {
		panic("unhelpful programmer error: empty error string")
	}
	i := &item{
		itemError,
		[]byte(msg),
		l.pos, // The "start" of the error is typically the current position, not where the token itself started.
	}
	l.items.Put(i)
	panic(i)
}
