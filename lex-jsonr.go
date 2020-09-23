package jsonr

import (
	"strings"
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
		if l.next() == eof {
			l.emit(itemEOF) // Useful to make EOF a token.
			return
		}
		l.backup()
		continue
	}
}

func lexWhitespace(l *lexer) {
	if l.acceptRun(" \t\r\n") {
		l.emit(itemWhitespace)
	}
}

func lexWhitespaceOrComment(l *lexer) {
	lexWhitespace(l)
	lexComment(l)
	lexWhitespace(l) // A trailing \n is not part of the comment.
}

func lexElement(l *lexer) {
	lexWhitespaceOrComment(l)
	lexValue(l)
	lexWhitespaceOrComment(l)
}

func lexValue(l *lexer) {
	prefix := l.input[l.pos:]
	switch {
	case hasPrefixByte(prefix, '{'):
		lexObject(l)
	case hasPrefixByte(prefix, '['):
		lexArray(l)
	case hasPrefixByte(prefix, '"'):
		lexString(l)
	case strings.HasPrefix(prefix, "true"):
		l.pos += 4
		l.emit(itemTrue)
	case strings.HasPrefix(prefix, "false"):
		l.pos += 5
		l.emit(itemFalse)
	case strings.HasPrefix(prefix, "null"):
		l.pos += 4
		l.emit(itemNull)
	case strings.ContainsAny(prefix, "-0123456789"):
		lexNumber(l)
	default:
		l.errorf("invalid element: %s", l.input[l.pos:l.pos+10])
	}
}

func lexObject(l *lexer) {
	l.pos += 1
	l.emit(itemObjectOpen)
	lexWhitespaceOrComment(l)
	lexMembers(l)
	lexWhitespaceOrComment(l)
	if l.accept("}") {
		l.emit(itemObjectClose)
	} else {
		l.errorf("unclosed object")
	}
}

func lexMembers(l *lexer) {
	for {
		lexWhitespaceOrComment(l)
		switch {
		case hasPrefixByte(l.input[l.pos:], '}'):
			return
		default:
			lexMember(l)
			if l.accept(",") {
				l.emit(itemComma)
			} else {
				return
			}
		}
	}
}

func lexMember(l *lexer) {
	lexWhitespaceOrComment(l)
	if !l.accept(`"`) {
		l.errorf("object key must be string pos:%d : %s", l.start, l.input[l.start:l.pos])
		return
	}
	lexString(l)
	lexWhitespaceOrComment(l)
	if !l.accept(`:`) {
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
	if l.accept("]") {
		l.emit(itemArrayClose)
	} else {
		l.errorf("unclosed array")
	}
}

func lexElements(l *lexer) {
	for {
		lexWhitespaceOrComment(l)
		switch {
		case hasPrefixByte(l.input[l.pos:], ']'):
			return
		default:
			lexElement(l)
			if l.accept(",") {
				l.emit(itemComma)
			} else {
				return
			}
		}
	}
}

func lexNumber(l *lexer) {
	// optional prefix
	l.accept("-")
	// The spec says leading zeros are verboten, but that seems pointlessly pedantic.
	if !l.acceptRun("0123456789") {
		l.errorf("malformed integer number")
		return
	}
	if l.accept(".") {
		if !l.acceptRun("0123456789") {
			l.errorf("malformed real number")
			return
		}
	}
	if l.accept("eE") {
		l.accept("+-")
		if !l.acceptRun("0123456789") {
			l.errorf("malformed exponent")
			return
		}
	}
	l.emit(itemNumber)
}

func lexString(l *lexer) {
	// swallow leading "
	l.pos += 1
	for {
		switch {
		case hasPrefixByte(l.input[l.pos:], '"'):
			l.pos += 1 // swallow ending "
			l.emit(itemString)
			return
		case l.accept(`\`):
			switch {
			case l.accept(`"\/bfnrt`):
				continue
			case l.accept("u"):
				for i := 0; i < 4; i++ {
					if l.accept("0123456789abcdefABCDEF") {
						l.errorf("invalid unicode escape sequence")
					}
				}
				continue
			default:
				l.errorf("invalid escaped character")
			}
		}
		if l.next() == eof {
			l.errorf("unexpected EOF scanning string")
		}
	}
}

func lexComment(l *lexer) {
	if !hasPrefixByte(l.input[l.pos:], '/') {
		return
	}
	if hasPrefixByte(l.input[l.pos+1:], '/') {
		lexLineComment(l)
		return
	}
	if hasPrefixByte(l.input[l.pos+1:], '*') {
		lexRangeComment(l)
		return
	}
}

func lexLineComment(l *lexer) {
	// swallow //
	l.pos += 2
	for {
		if hasPrefixByte(l.input[l.pos:], '\n') {
			// don't include trailng \n
			l.emit(itemComment)
			return
		}
		if l.next() == eof {
			break
		}
	}
	// Correctly reached EOF.
	if l.pos > l.start {
		l.emit(itemComment)
	}
	l.emit(itemEOF)
}

func lexRangeComment(l *lexer) {
	// swallow /*
	l.pos += 2
	for {
		if strings.HasPrefix(l.input[l.pos:], "*/") {
			// swallow */
			l.pos += 2
			l.emit(itemComment)
			return
		}
		if l.next() == eof {
			l.errorf("unexpected EOF scanning comment")
			return
		}
	}
}
