package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
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
	OPEN_PAR  = "OPEN_PAR"
	CLOSE_PAR = "CLOSE_PAR"
	ASSIGN    = "ASSIGN"
	END       = "END"
	PRINT     = "PRINT"
	IMUT      = "IMUT"
	IDEN      = "IDEN"
	EOF       = "EOF"
)

// Token holds the type and value of a lexical token
type Token struct {
	Type  string
	Value string
}

// ── Lexer ─────────────────────────────────────────────────────────────────────

type Lexer struct {
	source   string
	position int
	Next     Token
}

func NewLexer(source string) *Lexer {
	l := &Lexer{source: source, position: 0}
	l.selectNext()
	return l
}

func (l *Lexer) selectNext() {
	// Skip spaces and tabs but NOT newlines
	for l.position < len(l.source) && (l.source[l.position] == ' ' || l.source[l.position] == '\t' || l.source[l.position] == '\r') {
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

	if unicode.IsLetter(ch) {
		start := l.position
		for l.position < len(l.source) && (unicode.IsLetter(rune(l.source[l.position])) || unicode.IsDigit(rune(l.source[l.position])) || l.source[l.position] == '_') {
			l.position++
		}
		word := l.source[start:l.position]
		switch word {
		case "print":
			l.Next = Token{Type: PRINT, Value: word}
		case "imut":
			l.Next = Token{Type: IMUT, Value: word}
		default:
			l.Next = Token{Type: IDEN, Value: word}
		}
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
	case '=':
		l.Next = Token{Type: ASSIGN, Value: "="}
		l.position++
	case '\n':
		l.Next = Token{Type: END, Value: "\n"}
		l.position++
	default:
		panic(fmt.Sprintf("[Lexer] Invalid Symbol %c", ch))
	}
}

// ── PrePro ────────────────────────────────────────────────────────────────────

// PrePro is a preprocessor that removes inline comments before lexing
type PrePro struct{}

// Filter removes "--" comments, resolves "const" declarations via text substitution
func (p PrePro) Filter(code string) string {
	// Step 1: strip comments
	lines := strings.Split(code, "\n")
	for i, line := range lines {
		if idx := strings.Index(line, "--"); idx >= 0 {
			lines[i] = line[:idx]
		}
	}

	// Step 2: extract "const NAME = NUMBER" declarations and remove those lines
	constPattern := regexp.MustCompile(`^\s*const\s+([A-Za-z][A-Za-z0-9_]*)\s*=\s*(\d+)\s*$`)
	constMap := make(map[string]string)
	var remaining []string
	for _, line := range lines {
		if m := constPattern.FindStringSubmatch(line); m != nil {
			constMap[m[1]] = m[2]
		} else {
			remaining = append(remaining, line)
		}
	}

	// Step 3: substitute constant names with their values (whole-word replacement)
	code = strings.Join(remaining, "\n")
	for name, val := range constMap {
		re := regexp.MustCompile(`\b` + name + `\b`)
		code = re.ReplaceAllString(code, val)
	}

	return code
}

// ── Symbol Table ──────────────────────────────────────────────────────────────

// variable holds a value and an immutability flag
type variable struct {
	value int
	immut bool
}

// SymbolTable stores variable bindings for the program
type SymbolTable struct {
	table map[string]variable
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{table: make(map[string]variable)}
}

func (st *SymbolTable) Get(name string) int {
	v, ok := st.table[name]
	if !ok {
		panic(fmt.Sprintf("[Semantic] Undefined variable: %s", name))
	}
	return v.value
}

func (st *SymbolTable) Set(name string, val int) {
	if v, ok := st.table[name]; ok && v.immut {
		panic(fmt.Sprintf("[Semantic] cannot change the value of %s", name))
	}
	st.table[name] = variable{value: val, immut: false}
}

func (st *SymbolTable) SetImut(name string, val int) {
	st.table[name] = variable{value: val, immut: true}
}

// ── AST ───────────────────────────────────────────────────────────────────────

// Node is the base interface for all AST nodes
type Node interface {
	Evaluate(st *SymbolTable) int
}

// IntVal is a leaf node representing an integer literal
type IntVal struct {
	value    int
	children []Node
}

func (n *IntVal) Evaluate(st *SymbolTable) int { return n.value }

// UnOp represents a unary operation: "+" or "-"
type UnOp struct {
	value    string
	children []Node // 1 child
}

func (n *UnOp) Evaluate(st *SymbolTable) int {
	val := n.children[0].Evaluate(st)
	switch n.value {
	case "+":
		return val
	case "-":
		return -val
	}
	panic(fmt.Sprintf("[Semantic] Unknown unary operator: %s", n.value))
}

// BinOp represents a binary operation: "+", "-", "*", "/", "**"
type BinOp struct {
	value    string
	children []Node // 2 children: [left, right]
}

func (n *BinOp) Evaluate(st *SymbolTable) int {
	left := n.children[0].Evaluate(st)
	right := n.children[1].Evaluate(st)
	switch n.value {
	case "+":
		return left + right
	case "-":
		return left - right
	case "*":
		return left * right
	case "/":
		if right == 0 {
			panic("[Semantic] Division by zero")
		}
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

// Identifier is a leaf node representing a variable reference
type Identifier struct {
	value    string // variable name
	children []Node
}

func (n *Identifier) Evaluate(st *SymbolTable) int { return st.Get(n.value) }

// Assignment stores a value into the SymbolTable
type Assignment struct {
	value    string
	children []Node // [Identifier, Expression]
}

func (n *Assignment) Evaluate(st *SymbolTable) int {
	name := n.children[0].(*Identifier).value
	val := n.children[1].Evaluate(st)
	st.Set(name, val)
	return 0
}

// ImutAssignment stores a value into the SymbolTable as immutable
type ImutAssignment struct {
	value    string
	children []Node // [Identifier, Expression]
}

func (n *ImutAssignment) Evaluate(st *SymbolTable) int {
	name := n.children[0].(*Identifier).value
	val := n.children[1].Evaluate(st)
	st.SetImut(name, val)
	return 0
}

// Print evaluates its child and prints the result
type Print struct {
	value    string
	children []Node // 1 child
}

func (n *Print) Evaluate(st *SymbolTable) int {
	fmt.Println(n.children[0].Evaluate(st))
	return 0
}

// Block holds a sequence of statements and evaluates them all
type Block struct {
	value    string
	children []Node
}

func (n *Block) Evaluate(st *SymbolTable) int {
	for _, child := range n.children {
		child.Evaluate(st)
	}
	return 0
}

// NoOp is a dummy node for empty lines
type NoOp struct {
	value    string
	children []Node
}

func (n *NoOp) Evaluate(st *SymbolTable) int { return 0 }

// ── Parser ────────────────────────────────────────────────────────────────────

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

// parsePower parses: ATOM [ "**" FACTOR ]  (right-associative)
func parsePower(l *Lexer) Node {
	base := parseAtom(l)
	if l.Next.Type == POW {
		l.selectNext()
		exp := parseFactor(l)
		return &BinOp{value: "**", children: []Node{base, exp}}
	}
	return base
}

// parseFactor parses: ("+" | "-") FACTOR | IDENTIFIER | POWER
func parseFactor(l *Lexer) Node {
	if l.Next.Type == PLUS {
		l.selectNext()
		return &UnOp{value: "+", children: []Node{parseFactor(l)}}
	}
	if l.Next.Type == MINUS {
		l.selectNext()
		return &UnOp{value: "-", children: []Node{parseFactor(l)}}
	}
	if l.Next.Type == IDEN {
		name := l.Next.Value
		l.selectNext()
		return &Identifier{value: name}
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

// parseStatement parses: (IDEN "=" EXPRESSION | "imut" IDEN "=" EXPRESSION | "print" "(" EXPRESSION ")" | ε) "\n"
func parseStatement(l *Lexer) Node {
	if l.Next.Type == IMUT {
		l.selectNext()
		if l.Next.Type != IDEN {
			panic(fmt.Sprintf("[Parser] Expected identifier after 'imut' but got %s", l.Next.Type))
		}
		name := l.Next.Value
		l.selectNext()
		if l.Next.Type != ASSIGN {
			panic(fmt.Sprintf("[Parser] Expected '=' but got %s", l.Next.Type))
		}
		l.selectNext()
		expr := parseExpression(l)
		if l.Next.Type != END {
			panic(fmt.Sprintf("[Parser] Expected newline but got %s", l.Next.Type))
		}
		l.selectNext()
		return &ImutAssignment{children: []Node{&Identifier{value: name}, expr}}
	}
	if l.Next.Type == IDEN {
		name := l.Next.Value
		l.selectNext()
		if l.Next.Type != ASSIGN {
			panic(fmt.Sprintf("[Parser] Expected '=' but got %s", l.Next.Type))
		}
		l.selectNext()
		expr := parseExpression(l)
		if l.Next.Type != END {
			panic(fmt.Sprintf("[Parser] Expected newline but got %s", l.Next.Type))
		}
		l.selectNext()
		return &Assignment{children: []Node{&Identifier{value: name}, expr}}
	}
	if l.Next.Type == PRINT {
		l.selectNext()
		if l.Next.Type != OPEN_PAR {
			panic(fmt.Sprintf("[Parser] Expected '(' but got %s", l.Next.Type))
		}
		l.selectNext()
		expr := parseExpression(l)
		if l.Next.Type != CLOSE_PAR {
			panic(fmt.Sprintf("[Parser] Expected ')' but got %s", l.Next.Type))
		}
		l.selectNext()
		if l.Next.Type != END {
			panic(fmt.Sprintf("[Parser] Expected newline but got %s", l.Next.Type))
		}
		l.selectNext()
		return &Print{children: []Node{expr}}
	}
	if l.Next.Type == END {
		l.selectNext()
		return &NoOp{}
	}
	panic(fmt.Sprintf("[Parser] Unexpected token %s", l.Next.Type))
}

// parseProgram parses: { STATEMENT }
func parseProgram(l *Lexer) Node {
	var children []Node
	for l.Next.Type != EOF {
		children = append(children, parseStatement(l))
	}
	return &Block{children: children}
}

// run creates a Lexer and returns the AST root
func run(source string) Node {
	l := NewLexer(source)
	return parseProgram(l)
}

func main() {
	if len(os.Args) < 2 {
		panic("[Main] Nenhum argumento fornecido. Uso: go run main.go arquivo.lua")
	}
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(fmt.Sprintf("[Main] Erro ao ler arquivo: %v", err))
	}
	source := string(data) + "\n"
	source = PrePro{}.Filter(source)
	st := NewSymbolTable()
	run(source).Evaluate(st)
}
