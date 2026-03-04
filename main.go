package main

import (
	"fmt"
	"os"
	"strconv"
	"unicode"
)

// Token types
const (
	INT   = "INT"
	PLUS  = "PLUS"
	MINUS = "MINUS"
	XOR   = "XOR"
	EOF   = "EOF"
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

	if ch == '+' {
		l.Next = Token{Type: PLUS, Value: "+"}
		l.position++
		return
	}

	if ch == '-' {
		l.Next = Token{Type: MINUS, Value: "-"}
		l.position++
		return
	}

	if ch == '^' {
		l.Next = Token{Type: XOR, Value: "^"}
		l.position++
		return
	}

	panic(fmt.Sprintf("[Lexer] Invalid Symbol %c", ch))
}

// parseExpression parses and evaluates: INT ( ('+' | '-') INT )*
func parseExpression(l *Lexer) int {
	if l.Next.Type != INT {
		panic(fmt.Sprintf("[Parser] Unexpected token %s", l.Next.Type))
	}

	result, _ := strconv.Atoi(l.Next.Value)
	l.selectNext()

	for l.Next.Type == PLUS || l.Next.Type == MINUS || l.Next.Type == XOR {
		op := l.Next.Type
		l.selectNext()

		if l.Next.Type != INT {
			panic(fmt.Sprintf("[Parser] Unexpected token %s", l.Next.Type))
		}

		val, _ := strconv.Atoi(l.Next.Value)
		l.selectNext()

		if op == PLUS {
			result += val
		} else if op == MINUS {
			result -= val
		} else {
			result ^= val
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
