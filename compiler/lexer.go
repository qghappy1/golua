package compiler

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// token kind
const (
	TOKEN_EOF         = iota           // end-of-file
	TOKEN_VARARG                       // ...
	TOKEN_SEP_SEMI                     // ;
	TOKEN_SEP_COMMA                    // ,
	TOKEN_SEP_DOT                      // .
	TOKEN_SEP_COLON                    // :
	TOKEN_SEP_LABEL                    // ::
	TOKEN_SEP_LPAREN                   // (
	TOKEN_SEP_RPAREN                   // )
	TOKEN_SEP_LBRACK                   // [
	TOKEN_SEP_RBRACK                   // ]
	TOKEN_SEP_LCURLY                   // {
	TOKEN_SEP_RCURLY                   // }
	TOKEN_OP_ASSIGN                    // =
	TOKEN_OP_MINUS                     // - (sub or unm)
	TOKEN_OP_WAVE                      // ~ (bnot or bxor)
	TOKEN_OP_ADD                       // +
	TOKEN_OP_MUL                       // *
	TOKEN_OP_DIV                       // /
	TOKEN_OP_IDIV                      // //
	TOKEN_OP_POW                       // ^
	TOKEN_OP_MOD                       // %
	TOKEN_OP_BAND                      // &
	TOKEN_OP_BOR                       // |
	TOKEN_OP_SHR                       // >>
	TOKEN_OP_SHL                       // <<
	TOKEN_OP_CONCAT                    // ..
	TOKEN_OP_LT                        // <
	TOKEN_OP_LE                        // <=
	TOKEN_OP_GT                        // >
	TOKEN_OP_GE                        // >=
	TOKEN_OP_EQ                        // ==
	TOKEN_OP_NE                        // ~=
	TOKEN_OP_LEN                       // #
	TOKEN_OP_AND                       // and
	TOKEN_OP_OR                        // or
	TOKEN_OP_NOT                       // not
	TOKEN_KW_BREAK                     // break
	TOKEN_KW_DO                        // do
	TOKEN_KW_ELSE                      // else
	TOKEN_KW_ELSEIF                    // elseif
	TOKEN_KW_END                       // end
	TOKEN_KW_FALSE                     // false
	TOKEN_KW_FOR                       // for
	TOKEN_KW_FUNCTION                  // function
	TOKEN_KW_GOTO                      // goto
	TOKEN_KW_IF                        // if
	TOKEN_KW_IN                        // in
	TOKEN_KW_LOCAL                     // local
	TOKEN_KW_NIL                       // nil
	TOKEN_KW_REPEAT                    // repeat
	TOKEN_KW_RETURN                    // return
	TOKEN_KW_THEN                      // then
	TOKEN_KW_TRUE                      // true
	TOKEN_KW_UNTIL                     // until
	TOKEN_KW_WHILE                     // while
	TOKEN_IDENTIFIER                   // identifier
	TOKEN_NUMBER                       // number literal
	TOKEN_STRING                       // string literal
	TOKEN_OP_UNM      = TOKEN_OP_MINUS // unary minus
	TOKEN_OP_SUB      = TOKEN_OP_MINUS
	TOKEN_OP_BNOT     = TOKEN_OP_WAVE
	TOKEN_OP_BXOR     = TOKEN_OP_WAVE
)

var keywords = map[string]int{
	"and":      TOKEN_OP_AND,
	"break":    TOKEN_KW_BREAK,
	"do":       TOKEN_KW_DO,
	"else":     TOKEN_KW_ELSE,
	"elseif":   TOKEN_KW_ELSEIF,
	"end":      TOKEN_KW_END,
	"false":    TOKEN_KW_FALSE,
	"for":      TOKEN_KW_FOR,
	"function": TOKEN_KW_FUNCTION,
	"goto":     TOKEN_KW_GOTO,
	"if":       TOKEN_KW_IF,
	"in":       TOKEN_KW_IN,
	"local":    TOKEN_KW_LOCAL,
	"nil":      TOKEN_KW_NIL,
	"not":      TOKEN_OP_NOT,
	"or":       TOKEN_OP_OR,
	"repeat":   TOKEN_KW_REPEAT,
	"return":   TOKEN_KW_RETURN,
	"then":     TOKEN_KW_THEN,
	"true":     TOKEN_KW_TRUE,
	"until":    TOKEN_KW_UNTIL,
	"while":    TOKEN_KW_WHILE,
}

var reNewLine = regexp.MustCompile("\r\n|\n\r|\n|\r")
var reIdentifier = regexp.MustCompile(`^[_\d\w]+`)
var reNumber = regexp.MustCompile(`^0[xX][0-9a-fA-F]*(\.[0-9a-fA-F]*)?([pP][+\-]?[0-9]+)?|^[0-9]*(\.[0-9]*)?([eE][+\-]?[0-9]+)?`)
var reShortStr = regexp.MustCompile(`(?s)(^'(\\\\|\\'|\\\n|\\z\s*|[^'\n])*')|(^"(\\\\|\\"|\\\n|\\z\s*|[^"\n])*")`)
var reOpeningLongBracket = regexp.MustCompile(`^\[=*\[`)

var reDecEscapeSeq = regexp.MustCompile(`^\\[0-9]{1,3}`)
var reHexEscapeSeq = regexp.MustCompile(`^\\x[0-9a-fA-F]{2}`)
var reUnicodeEscapeSeq = regexp.MustCompile(`^\\u\{[0-9a-fA-F]+\}`)

type Lexer struct {
	chunk string 		// 源代码
	chunkName string 	// 源文件名
	line int 			// 当前行号

	nextToken     string
	nextTokenKind int
	nextTokenLine int
}


func NewLexer(chunk, chunkName string) *Lexer {
	return &Lexer{chunk, chunkName, 1, "", 0, 0}
}

func (this *Lexer) Line() int {
	return this.line
}

func (this *Lexer) LookAhead() int {
	if this.nextTokenLine > 0 {
		return this.nextTokenKind
	}
	currentLine := this.line
	line, kind, token := this.NextToken()
	this.line = currentLine
	this.nextTokenLine = line
	this.nextTokenKind = kind
	this.nextToken = token
	return kind
}

func (this *Lexer) NextIdentifier() (line int, token string) {
	return this.NextTokenOfKind(TOKEN_IDENTIFIER)
}

func (this *Lexer) NextTokenOfKind(kind int) (line int, token string) {
	line, _kind, token := this.NextToken()
	if kind != _kind {
		this.error("syntax error near '%s'", token)
	}
	return line, token
}

func (this *Lexer) NextToken() (line, kind int, token string) {
	if this.nextTokenLine > 0 {
		line = this.nextTokenLine
		kind = this.nextTokenKind
		token = this.nextToken
		this.line = this.nextTokenLine
		this.nextTokenLine = 0
		return
	}

	this.skipWhiteSpaces()
	if len(this.chunk) == 0 {
		return this.line, TOKEN_EOF, "EOF"
	}

	switch this.chunk[0] {
	case ';':
		this.next(1)
		return this.line, TOKEN_SEP_SEMI, ";"
	case ',':
		this.next(1)
		return this.line, TOKEN_SEP_COMMA, ","
	case '(':
		this.next(1)
		return this.line, TOKEN_SEP_LPAREN, "("
	case ')':
		this.next(1)
		return this.line, TOKEN_SEP_RPAREN, ")"
	case ']':
		this.next(1)
		return this.line, TOKEN_SEP_RBRACK, "]"
	case '{':
		this.next(1)
		return this.line, TOKEN_SEP_LCURLY, "{"
	case '}':
		this.next(1)
		return this.line, TOKEN_SEP_RCURLY, "}"
	case '+':
		this.next(1)
		return this.line, TOKEN_OP_ADD, "+"
	case '-':
		this.next(1)
		return this.line, TOKEN_OP_MINUS, "-"
	case '*':
		this.next(1)
		return this.line, TOKEN_OP_MUL, "*"
	case '^':
		this.next(1)
		return this.line, TOKEN_OP_POW, "^"
	case '%':
		this.next(1)
		return this.line, TOKEN_OP_MOD, "%"
	case '&':
		this.next(1)
		return this.line, TOKEN_OP_BAND, "&"
	case '|':
		this.next(1)
		return this.line, TOKEN_OP_BOR, "|"
	case '#':
		this.next(1)
		return this.line, TOKEN_OP_LEN, "#"
	case ':':
		if this.test("::") {
			this.next(2)
			return this.line, TOKEN_SEP_LABEL, "::"
		} else {
			this.next(1)
			return this.line, TOKEN_SEP_COLON, ":"
		}
	case '/':
		if this.test("//") {
			this.next(2)
			return this.line, TOKEN_OP_IDIV, "//"
		} else {
			this.next(1)
			return this.line, TOKEN_OP_DIV, "/"
		}
	case '~':
		if this.test("~=") {
			this.next(2)
			return this.line, TOKEN_OP_NE, "~="
		} else {
			this.next(1)
			return this.line, TOKEN_OP_WAVE, "~"
		}
	case '=':
		if this.test("==") {
			this.next(2)
			return this.line, TOKEN_OP_EQ, "=="
		} else {
			this.next(1)
			return this.line, TOKEN_OP_ASSIGN, "="
		}
	case '<':
		if this.test("<<") {
			this.next(2)
			return this.line, TOKEN_OP_SHL, "<<"
		} else if this.test("<=") {
			this.next(2)
			return this.line, TOKEN_OP_LE, "<="
		} else {
			this.next(1)
			return this.line, TOKEN_OP_LT, "<"
		}
	case '>':
		if this.test(">>") {
			this.next(2)
			return this.line, TOKEN_OP_SHR, ">>"
		} else if this.test(">=") {
			this.next(2)
			return this.line, TOKEN_OP_GE, ">="
		} else {
			this.next(1)
			return this.line, TOKEN_OP_GT, ">"
		}
	case '.':
		if this.test("...") {
			this.next(3)
			return this.line, TOKEN_VARARG, "..."
		} else if this.test("..") {
			this.next(2)
			return this.line, TOKEN_OP_CONCAT, ".."
		} else if len(this.chunk) == 1 || !isDigit(this.chunk[1]) {
			this.next(1)
			return this.line, TOKEN_SEP_DOT, "."
		}
	case '[':
		if this.test("[[") || this.test("[=") {
			return this.line, TOKEN_STRING, this.scanLongString()
		} else {
			this.next(1)
			return this.line, TOKEN_SEP_LBRACK, "["
		}
	case '\'', '"':
		return this.line, TOKEN_STRING, this.scanShortString()
	}

	c := this.chunk[0]
	if c == '.' || isDigit(c) {
		token := this.scanNumber()
		return this.line, TOKEN_NUMBER, token
	}
	if c == '_' || isLetter(c) {
		token := this.scanIdentifier()
		if kind, found := keywords[token]; found {
			return this.line, kind, token // keyword
		} else {
			return this.line, TOKEN_IDENTIFIER, token
		}
	}

	this.error("unexpected symbol near %q", c)
	return
}

func (this *Lexer) next(n int) {
	this.chunk = this.chunk[n:]
}

func (this *Lexer) test(s string) bool {
	return strings.HasPrefix(this.chunk, s)
}

func (this *Lexer) error(f string, a ...interface{}) {
	err := fmt.Sprintf(f, a...)
	err = fmt.Sprintf("%s:%d: %s", this.chunkName, this.line, err)
	panic(err)
}

func (this *Lexer) skipWhiteSpaces() {
	for len(this.chunk) > 0 {
		if this.test("--") {
			this.skipComment()
		} else if this.test("\r\n") || this.test("\n\r") {
			this.next(2)
			this.line += 1
		} else if isNewLine(this.chunk[0]) {
			this.next(1)
			this.line += 1
		} else if isWhiteSpace(this.chunk[0]) {
			this.next(1)
		} else {
			break
		}
	}
}

func (this *Lexer) skipComment() {
	this.next(2) // skip --

	// long comment ?
	if this.test("[") {
		if reOpeningLongBracket.FindString(this.chunk) != "" {
			this.scanLongString()
			return
		}
	}

	// short comment
	for len(this.chunk) > 0 && !isNewLine(this.chunk[0]) {
		this.next(1)
	}
}

func (this *Lexer) scanIdentifier() string {
	return this.scan(reIdentifier)
}

func (this *Lexer) scanNumber() string {
	return this.scan(reNumber)
}

func (this *Lexer) scan(re *regexp.Regexp) string {
	if token := re.FindString(this.chunk); token != "" {
		this.next(len(token))
		return token
	}
	panic("unreachable!")
}

func (this *Lexer) scanLongString() string {
	openingLongBracket := reOpeningLongBracket.FindString(this.chunk)
	if openingLongBracket == "" {
		this.error("invalid long string delimiter near '%s'",
			this.chunk[0:2])
	}

	closingLongBracket := strings.Replace(openingLongBracket, "[", "]", -1)
	closingLongBracketIdx := strings.Index(this.chunk, closingLongBracket)
	if closingLongBracketIdx < 0 {
		this.error("unfinished long string or comment")
	}

	str := this.chunk[len(openingLongBracket):closingLongBracketIdx]
	this.next(closingLongBracketIdx + len(closingLongBracket))

	str = reNewLine.ReplaceAllString(str, "\n")
	this.line += strings.Count(str, "\n")
	if len(str) > 0 && str[0] == '\n' {
		str = str[1:]
	}

	return str
}

func (this *Lexer) scanShortString() string {
	if str := reShortStr.FindString(this.chunk); str != "" {
		this.next(len(str))
		str = str[1 : len(str)-1]
		if strings.Index(str, `\`) >= 0 {
			this.line += len(reNewLine.FindAllString(str, -1))
			str = this.escape(str)
		}
		return str
	}
	this.error("unfinished string")
	return ""
}

func (this *Lexer) escape(str string) string {
	var buf bytes.Buffer

	for len(str) > 0 {
		if str[0] != '\\' {
			buf.WriteByte(str[0])
			str = str[1:]
			continue
		}

		if len(str) == 1 {
			this.error("unfinished string")
		}

		switch str[1] {
		case 'a':
			buf.WriteByte('\a')
			str = str[2:]
			continue
		case 'b':
			buf.WriteByte('\b')
			str = str[2:]
			continue
		case 'f':
			buf.WriteByte('\f')
			str = str[2:]
			continue
		case 'n', '\n':
			buf.WriteByte('\n')
			str = str[2:]
			continue
		case 'r':
			buf.WriteByte('\r')
			str = str[2:]
			continue
		case 't':
			buf.WriteByte('\t')
			str = str[2:]
			continue
		case 'v':
			buf.WriteByte('\v')
			str = str[2:]
			continue
		case '"':
			buf.WriteByte('"')
			str = str[2:]
			continue
		case '\'':
			buf.WriteByte('\'')
			str = str[2:]
			continue
		case '\\':
			buf.WriteByte('\\')
			str = str[2:]
			continue
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9': // \ddd
			if found := reDecEscapeSeq.FindString(str); found != "" {
				d, _ := strconv.ParseInt(found[1:], 10, 32)
				if d <= 0xFF {
					buf.WriteByte(byte(d))
					str = str[len(found):]
					continue
				}
				this.error("decimal escape too large near '%s'", found)
			}
		case 'x': // \xXX
			if found := reHexEscapeSeq.FindString(str); found != "" {
				d, _ := strconv.ParseInt(found[2:], 16, 32)
				buf.WriteByte(byte(d))
				str = str[len(found):]
				continue
			}
		case 'u': // \u{XXX}
			if found := reUnicodeEscapeSeq.FindString(str); found != "" {
				d, err := strconv.ParseInt(found[3:len(found)-1], 16, 32)
				if err == nil && d <= 0x10FFFF {
					buf.WriteRune(rune(d))
					str = str[len(found):]
					continue
				}
				this.error("UTF-8 value too large near '%s'", found)
			}
		case 'z':
			str = str[2:]
			for len(str) > 0 && isWhiteSpace(str[0]) { // todo
				str = str[1:]
			}
			continue
		}
		this.error("invalid escape sequence near '\\%c'", str[1])
	}

	return buf.String()
}

func isWhiteSpace(c byte) bool {
	switch c {
	case '\t', '\n', '\v', '\f', '\r', ' ':
		return true
	}
	return false
}

func isNewLine(c byte) bool {
	return c == '\r' || c == '\n'
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isLetter(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}