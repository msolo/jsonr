package ast

import (
	"bytes"
)

const (
	itemError itemType = iota // error occurred; value is text of error

	itemEOF
	itemWhitespace
	itemComment
	itemString
	itemTrue
	itemFalse
	itemNull
	itemArrayOpen
	itemArrayClose
	itemComma
	itemObjectOpen
	itemObjectClose
	itemColon
	itemNumber
)

func lexStream(l *lexer) {
	for {
		lexElement(l)
		if _, eof := l.next(); eof {
			l.emit(itemEOF) // Useful to make EOF a token.
			return
		}
		l.backup()
		continue
	}
}

func lexWhitespace(l *lexer) {
	if l.acceptWhitespace() {
		l.emit(itemWhitespace)
	}
}

func lexWhitespaceOrComment(l *lexer) {
	lexWhitespace(l)
	for lexComment(l) {
		lexWhitespace(l) // A trailing \n is not part of the comment.
	}
	lexWhitespace(l) // A trailing \n is not part of the comment.
}

func lexElement(l *lexer) {
	lexWhitespaceOrComment(l)
	lexValue(l)
	lexWhitespaceOrComment(l)
}

func lexValue(l *lexer) {
	b := l.input[l.pos]
	switch b {
	case '{':
		lexObject(l)
	case '[':
		lexArray(l)
	case '"':
		lexString(l)
	case 't':
		l.pos += 1
		if l.acceptByte('r') && l.acceptByte('u') && l.acceptByte('e') {
			l.emit(itemTrue)
		} else {
			l.errorf("failed parsing true")
		}
	case 'f':
		l.pos += 1
		if l.acceptByte('a') && l.acceptByte('l') && l.acceptByte('s') && l.acceptByte('e') {
			l.emit(itemFalse)
		} else {
			l.errorf("failed parsing false")
		}
	case 'n':
		l.pos += 1
		if l.acceptByte('u') && l.acceptByte('l') && l.acceptByte('l') {
			l.emit(itemNull)
		} else {
			l.errorf("failed parsing null")
		}
	case '-':
		lexNumber(l)
	case '+':
		l.errorf("malformed number: %s", l.input[l.pos:min(len(l.input), l.pos+10)])
	default:
		if '0' <= b && b <= '9' {
			lexNumber(l)
		} else {
			l.errorf("invalid element: %s", l.input[l.pos:min(len(l.input), l.pos+10)])
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func lexObject(l *lexer) {
	l.pos += 1
	l.emit(itemObjectOpen)
	lexWhitespaceOrComment(l)
	lexMembers(l)
	lexWhitespaceOrComment(l)
	if l.acceptByte('}') {
		l.emit(itemObjectClose)
	} else {
		l.errorf("unclosed object")
	}
}

func lexMembers(l *lexer) {
	for {
		lexWhitespaceOrComment(l)
		switch {
		case l.peek() == '}':
			return
		default:
			lexMember(l)
			if l.acceptByte(',') {
				l.emit(itemComma)
			} else {
				return
			}
		}
	}
}

func lexMember(l *lexer) {
	lexWhitespaceOrComment(l)
	if !l.acceptByte('"') {
		l.errorf("object key must be string pos:%d : %s", l.start, l.input[l.start:l.pos])
		return
	}
	lexString(l)
	lexWhitespaceOrComment(l)
	if !l.acceptByte(':') {
		l.errorf("object member has no : delimiter")
		return
	}
	l.emit(itemColon)
	lexWhitespaceOrComment(l)

	lexValue(l)
	lexWhitespaceOrComment(l)
}

func lexArray(l *lexer) {
	l.pos += 1
	l.emit(itemArrayOpen)
	lexWhitespaceOrComment(l)
	lexElements(l)
	lexWhitespaceOrComment(l)
	if l.acceptByte(']') {
		l.emit(itemArrayClose)
	} else {
		l.errorf("unclosed array")
	}
}

func lexElements(l *lexer) {
	for {
		lexWhitespaceOrComment(l)
		switch {
		case l.peek() == ']':
			return
		default:
			lexElement(l)
			if l.acceptByte(',') {
				l.emit(itemComma)
			} else {
				return
			}
		}
	}
}

var validExponent = []byte("eE")
var validSign = []byte("+-")

func lexNumber(l *lexer) {
	// optional prefix
	l.acceptByte('-')
	// The spec says leading zeros are verboten, but that seems pointlessly pedantic.
	if !l.acceptInt() {
		l.errorf("malformed integer number")
		return
	}
	if l.acceptByte('.') {
		if !l.acceptInt() {
			l.errorf("malformed real number")
			return
		}
	}
	if l.accept(validExponent) {
		l.accept(validSign)
		if !l.acceptInt() {
			l.errorf("malformed exponent")
			return
		}
	}
	l.emit(itemNumber)
}

var validEscapes = []byte(`"\/bfnrt`)

func lexString(l *lexer) {
	// swallow leading "
	l.pos += 1

	for {
		switch {
		case l.acceptByte('"'):
			l.emit(itemString)
			return
		case l.acceptByte('\\'):
			switch {
			case l.accept(validEscapes):
				continue
			case l.acceptByte('u'):
				for i := 0; i < 4; i++ {
					if l.acceptHex() {
						l.errorf("invalid unicode escape sequence")
					}
				}
				continue
			default:
				l.errorf("invalid escaped character")
			}
		}
		if _, eof := l.next(); eof {
			l.errorf("unexpected EOF scanning string")
		}
	}
}

var commentStart = []byte("//")
var commentStartRange = []byte("/*")

func lexComment(l *lexer) bool {
	if l.peek() == '/' {
		l.pos += 1
		switch l.peek() {
		case '/':
			lexLineComment(l)
			return true
		case '*':
			lexRangeComment(l)
			return true
		}
		l.backup()
	}

	return false
}

func lexLineComment(l *lexer) {
	// swallow //
	l.pos += 1
	for {
		b, eof := l.next()

		if b == '\n' {
			l.backup()
			// don't include trailng \n
			l.emit(itemComment)
			return
		}
		if eof {
			break
		}
	}
	// Correctly reached EOF.
	if l.pos > l.start {
		l.emit(itemComment)
	}
	l.emit(itemEOF)
}

var endRangeComment = []byte("*/")

func lexRangeComment(l *lexer) {
	// swallow /*
	l.pos += 1
	for {
		if bytes.HasPrefix(l.input[l.pos:], endRangeComment) {
			// swallow */
			l.pos += 2
			l.emit(itemComment)
			return
		}
		if _, eof := l.next(); eof {
			l.errorf("unexpected EOF scanning comment")
			return
		}
	}
}
