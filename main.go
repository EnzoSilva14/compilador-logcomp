package main

import (
	"fmt"
	"os"
	"strconv"
	"unicode"
)

// Token types
const (
	INT       = "INT"
	PLUS      = "PLUS"
	MINUS     = "MINUS"
	MULT      = "MULT"
	DIV       = "DIV"
	OPEN_PAR  = "OPEN_PAR"
	CLOSE_PAR = "CLOSE_PAR"
	EOF       = "EOF"
)

// Token holds the type and value of a lexical token
type Token struct {
	Type  string
	Value string
}

// Lexer tokenizes the input string on demand
type Lexer struct {
	source   string
	position int
	Next     Token
}

// NewLexer creates a Lexer and reads the first token
func NewLexer(source string) *Lexer {
	l := &Lexer{source: source, position: 0}
	l.selectNext()
	return l
}

// selectNext advances the Lexer to the next token, storing it in Next
func (l *Lexer) selectNext() {
	for l.position < len(l.source) && unicode.IsSpace(rune(l.source[l.position])) {
		l.position++
	}

	if l.position >= len(l.source) {
		l.Next = Token{Type: EOF, Value: ""}
		return
	}

	ch := rune(l.source[l.position])

	if unicode.IsDigit(ch) {
		start := l.position
		for l.position < len(l.source) && unicode.IsDigit(rune(l.source[l.position])) {
			l.position++
		}
		l.Next = Token{Type: INT, Value: l.source[start:l.position]}
		return
	}

	switch ch {
	case '+':
		l.Next = Token{Type: PLUS, Value: "+"}
	case '-':
		l.Next = Token{Type: MINUS, Value: "-"}
	case '*':
		l.Next = Token{Type: MULT, Value: "*"}
	case '/':
		l.Next = Token{Type: DIV, Value: "/"}
	case '(':
		l.Next = Token{Type: OPEN_PAR, Value: "("}
	case ')':
		l.Next = Token{Type: CLOSE_PAR, Value: ")"}
	default:
		panic(fmt.Sprintf("[Lexer] Invalid Symbol %c", ch))
	}
	l.position++
}

// parseFactor parses: ("+" | "-") FACTOR | "(" EXPRESSION ")" | NUMBER
func parseFactor(l *Lexer) int {
	if l.Next.Type == PLUS {
		l.selectNext()
		return +parseFactor(l)
	}
	if l.Next.Type == MINUS {
		l.selectNext()
		return -parseFactor(l)
	}
	if l.Next.Type == OPEN_PAR {
		l.selectNext()
		result := parseExpression(l)
		if l.Next.Type != CLOSE_PAR {
			panic(fmt.Sprintf("[Parser] Expected ')' but got %s", l.Next.Type))
		}
		l.selectNext()
		return result
	}
	if l.Next.Type == INT {
		val, _ := strconv.Atoi(l.Next.Value)
		l.selectNext()
		return val
	}
	panic(fmt.Sprintf("[Parser] Unexpected token %s", l.Next.Type))
}

// parseTerm parses: FACTOR { ("*" | "/") FACTOR }
func parseTerm(l *Lexer) int {
	result := parseFactor(l)
	for l.Next.Type == MULT || l.Next.Type == DIV {
		op := l.Next.Type
		l.selectNext()
		val := parseFactor(l)
		if op == MULT {
			result *= val
		} else {
			result /= val
		}
	}
	return result
}

// parseExpression parses: TERM { ("+" | "-") TERM }
func parseExpression(l *Lexer) int {
	result := parseTerm(l)
	for l.Next.Type == PLUS || l.Next.Type == MINUS {
		op := l.Next.Type
		l.selectNext()
		val := parseTerm(l)
		if op == PLUS {
			result += val
		} else {
			result -= val
		}
	}
	return result
}

// run creates a Lexer, parses the full expression, and verifies EOF
func run(source string) int {
	l := NewLexer(source)
	result := parseExpression(l)
	if l.Next.Type != EOF {
		panic(fmt.Sprintf("[Parser] Unexpected token %s", l.Next.Type))
	}
	return result
}

func main() {
	if len(os.Args) < 2 {
		panic("[Main] Nenhum argumento fornecido. Uso: go run main.go 'expressao'")
	}
	fmt.Println(run(os.Args[1]))
}
