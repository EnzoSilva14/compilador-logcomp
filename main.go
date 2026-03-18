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
	POW       = "POW"
	FACT      = "FACT"
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
		l.position++
	case '-':
		l.Next = Token{Type: MINUS, Value: "-"}
		l.position++
	case '*':
		if l.position+1 < len(l.source) && l.source[l.position+1] == '*' {
			l.Next = Token{Type: POW, Value: "**"}
			l.position += 2
		} else {
			l.Next = Token{Type: MULT, Value: "*"}
			l.position++
		}
	case '/':
		l.Next = Token{Type: DIV, Value: "/"}
		l.position++
	case '(':
		l.Next = Token{Type: OPEN_PAR, Value: "("}
		l.position++
	case ')':
		l.Next = Token{Type: CLOSE_PAR, Value: ")"}
		l.position++
	case '!':
		l.Next = Token{Type: FACT, Value: "!"}
		l.position++
	default:
		panic(fmt.Sprintf("[Lexer] Invalid Symbol %c", ch))
	}
}

// ── AST ──────────────────────────────────────────────────────────────────────

// Node is the base interface for all AST nodes
type Node interface {
	Evaluate() int
}

// IntVal is a leaf node representing an integer literal
type IntVal struct {
	value    int
	children []Node // always empty
}

func (n *IntVal) Evaluate() int {
	return n.value
}

// UnOp represents a unary operation: "+", "-", or "!" (postfix factorial)
type UnOp struct {
	value    string // operator symbol
	children []Node // exactly 1 child (the operand)
}

func (n *UnOp) Evaluate() int {
	val := n.children[0].Evaluate()
	switch n.value {
	case "+":
		return val
	case "-":
		return -val
	case "!":
		if val < 0 {
			panic(fmt.Sprintf("[Semantic] Factorial of negative number: %d", val))
		}
		result := 1
		for i := 2; i <= val; i++ {
			result *= i
		}
		return result
	}
	panic(fmt.Sprintf("[Semantic] Unknown unary operator: %s", n.value))
}

// BinOp represents a binary operation: "+", "-", "*", "/", "**"
type BinOp struct {
	value    string // operator symbol
	children []Node // exactly 2 children: [left, right]
}

func (n *BinOp) Evaluate() int {
	left := n.children[0].Evaluate()
	right := n.children[1].Evaluate()
	switch n.value {
	case "+":
		return left + right
	case "-":
		return left - right
	case "*":
		return left * right
	case "/":
		return left / right
	case "**":
		result := 1
		for i := 0; i < right; i++ {
			result *= left
		}
		return result
	}
	panic(fmt.Sprintf("[Semantic] Unknown binary operator: %s", n.value))
}

// ── Parser ───────────────────────────────────────────────────────────────────

// parseAtom parses: "(" EXPRESSION ")" | NUMBER
func parseAtom(l *Lexer) Node {
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
		return &IntVal{value: val}
	}
	panic(fmt.Sprintf("[Parser] Unexpected token %s", l.Next.Type))
}

// parsePower parses: ATOM { "!" } [ "**" FACTOR ]  (** is right-associative)
// Postfix "!" binds tighter than "**", so -3! means -(3!)
func parsePower(l *Lexer) Node {
	base := parseAtom(l)
	for l.Next.Type == FACT {
		l.selectNext()
		base = &UnOp{value: "!", children: []Node{base}}
	}
	if l.Next.Type == POW {
		l.selectNext()
		exp := parseFactor(l)
		return &BinOp{value: "**", children: []Node{base, exp}}
	}
	return base
}

// parseFactor parses: ("+" | "-") FACTOR | POWER
func parseFactor(l *Lexer) Node {
	if l.Next.Type == PLUS {
		l.selectNext()
		return &UnOp{value: "+", children: []Node{parseFactor(l)}}
	}
	if l.Next.Type == MINUS {
		l.selectNext()
		return &UnOp{value: "-", children: []Node{parseFactor(l)}}
	}
	return parsePower(l)
}

// parseTerm parses: FACTOR { ("*" | "/") FACTOR }
func parseTerm(l *Lexer) Node {
	result := parseFactor(l)
	for l.Next.Type == MULT || l.Next.Type == DIV {
		op := l.Next.Value
		l.selectNext()
		result = &BinOp{value: op, children: []Node{result, parseFactor(l)}}
	}
	return result
}

// parseExpression parses: TERM { ("+" | "-") TERM }
func parseExpression(l *Lexer) Node {
	result := parseTerm(l)
	for l.Next.Type == PLUS || l.Next.Type == MINUS {
		op := l.Next.Value
		l.selectNext()
		result = &BinOp{value: op, children: []Node{result, parseTerm(l)}}
	}
	return result
}

// run creates a Lexer, parses the full expression, and returns the AST root
func run(source string) Node {
	l := NewLexer(source)
	root := parseExpression(l)
	if l.Next.Type != EOF {
		panic(fmt.Sprintf("[Parser] Unexpected token %s", l.Next.Type))
	}
	return root
}

func main() {
	if len(os.Args) < 2 {
		panic("[Main] Nenhum argumento fornecido. Uso: go run main.go 'expressao'")
	}
	fmt.Println(run(os.Args[1]).Evaluate())
}
