package jsonr

import "strings"

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

func lexStream(l *lexer) stateFn {
	for {
		lexElement(l)
		if l.next() == eof {
			l.emit(itemEOF) // Useful to make EOF a token.
			return nil      // Stop the run loop.
		}
		l.backup()
		continue
	}
}

func lexElement(l *lexer) {
	lexWhitespace(l)
	lexValue(l)
	lexWhitespace(l)
}

func lexWhitespace(l *lexer) {
	if l.acceptRun(" \t\r\n") {
		l.emit(itemWhitespace)
	}
}

func lexValue(l *lexer) {
	switch {
	case hasPrefixByte(l.input[l.pos:], '{'):
		lexObject(l)
	case hasPrefixByte(l.input[l.pos:], '['):
		lexArray(l)
	case hasPrefixByte(l.input[l.pos:], '"'):
		lexString(l)
	case strings.HasPrefix(l.input[l.pos:], "true"):
		l.pos += 4
		l.emit(itemTrue)
	case strings.HasPrefix(l.input[l.pos:], "false"):
		l.pos += 5
		l.emit(itemFalse)
	case strings.HasPrefix(l.input[l.pos:], "null"):
		l.pos += 4
		l.emit(itemNull)
	case strings.ContainsAny(l.input[l.pos:], "-0123456789"):
		lexNumber(l)
	default:
		l.errorf("invalid element")
	}
}

func lexObject(l *lexer) {
	l.pos += 1
	l.emit(itemObjectOpen)
	lexWhitespace(l)
	lexMembers(l)
	lexWhitespace(l)
	if l.accept("}") {
		l.emit(itemObjectClose)
	} else {
		l.errorf("unclosed object")
	}
}

func lexMembers(l *lexer) {
	for {
		lexWhitespace(l)
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
	if !l.accept(`"`) {
		l.errorf("object key must be string")
		return
	}
	lexString(l)
	lexWhitespace(l)
	if !l.accept(`:`) {
		l.errorf("object member has no : delimiter")
		return
	}
	l.emit(itemColon)
	lexWhitespace(l)

	lexValue(l)
	lexWhitespace(l)
}

func lexArray(l *lexer) {
	l.pos += 1
	l.emit(itemArrayOpen)
	lexWhitespace(l)
	lexElements(l)
	lexWhitespace(l)
	if l.accept("]") {
		l.emit(itemArrayClose)
	} else {
		l.errorf("unclosed array")
	}
}

func lexElements(l *lexer) {
	for {
		lexWhitespace(l)
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
			case l.accept("u"):
				for i := 0; i < 4; i++ {
					if l.accept("0123456789abcdefABCDEF") {
						l.errorf("invalid unicode escape sequence")
					}
				}
			default:
				l.errorf("invalid escaped character")
			}
		}
		if l.next() == eof {
			l.errorf("unexected EOF scanning string")
		}
	}
}

func lexComment(l *lexer) stateFn {
	if strings.HasPrefix(l.input[l.pos:], "//") {
		return lexLineComment
	}
	if strings.HasPrefix(l.input[l.pos:], "/*") {
		return lexRangeComment
	}
	return lexStream
}

func lexLineComment(l *lexer) stateFn {
	// swallow //
	l.pos += 2
	for {
		if hasPrefixByte(l.input[l.pos:], '\n') {
			// don't include trailng \n
			l.emit(itemComment)
			return lexStream
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
	return nil // Stop the run loop.
}

func lexRangeComment(l *lexer) stateFn {
	// swallow /*
	l.pos += 2
	for {
		if strings.HasPrefix(l.input[l.pos:], "*/") {
			// swallow */
			l.pos += 2
			l.emit(itemComment)
			return lexStream
		}
		if l.next() == eof {
			return l.errorf("unexpected EOF scanning comment")
		}
	}
}
