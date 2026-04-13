package main

import (
	"bufio"
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
	END       = "END"    // newline token
	KW_END    = "KW_END" // "end" keyword
	PRINT     = "PRINT"
	IMUT      = "IMUT"
	IDEN      = "IDEN"
	EOF       = "EOF"
	// Boolean / relational operators
	AND = "AND"
	OR  = "OR"
	NOT = "NOT"
	EQ  = "EQ" // ==
	GT  = "GT" // >
	LT  = "LT" // <
	// Control-flow keywords
	IF    = "IF"
	WHILE = "WHILE"
	ELSE  = "ELSE"
	READ  = "READ"
	THEN  = "THEN"
	DO    = "DO"
	// Extra credit keywords
	FOR    = "FOR"
	REPEAT = "REPEAT"
	UNTIL  = "UNTIL"
	COMMA  = "COMMA"
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
		case "and":
			l.Next = Token{Type: AND, Value: word}
		case "or":
			l.Next = Token{Type: OR, Value: word}
		case "not":
			l.Next = Token{Type: NOT, Value: word}
		case "if":
			l.Next = Token{Type: IF, Value: word}
		case "while":
			l.Next = Token{Type: WHILE, Value: word}
		case "else":
			l.Next = Token{Type: ELSE, Value: word}
		case "read":
			l.Next = Token{Type: READ, Value: word}
		case "then":
			l.Next = Token{Type: THEN, Value: word}
		case "do":
			l.Next = Token{Type: DO, Value: word}
		case "end":
			l.Next = Token{Type: KW_END, Value: word}
		case "for":
			l.Next = Token{Type: FOR, Value: word}
		case "repeat":
			l.Next = Token{Type: REPEAT, Value: word}
		case "until":
			l.Next = Token{Type: UNTIL, Value: word}
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
		if l.position+1 < len(l.source) && l.source[l.position+1] == '=' {
			l.Next = Token{Type: EQ, Value: "=="}
			l.position += 2
		} else {
			l.Next = Token{Type: ASSIGN, Value: "="}
			l.position++
		}
	case '>':
		l.Next = Token{Type: GT, Value: ">"}
		l.position++
	case '<':
		l.Next = Token{Type: LT, Value: "<"}
		l.position++
	case '|':
		if l.position+1 < len(l.source) && l.source[l.position+1] == '|' {
			l.Next = Token{Type: OR, Value: "||"}
			l.position += 2
		} else {
			panic(fmt.Sprintf("[Lexer] Invalid Symbol %c", ch))
		}
	case '&':
		if l.position+1 < len(l.source) && l.source[l.position+1] == '&' {
			l.Next = Token{Type: AND, Value: "&&"}
			l.position += 2
		} else {
			panic(fmt.Sprintf("[Lexer] Invalid Symbol %c", ch))
		}
	case ',':
		l.Next = Token{Type: COMMA, Value: ","}
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
	constPattern := regexp.MustCompile(`^\s*const\s+([A-Za-z][A-Za-z0-9_]*)\s+(\d+)\s*$`)
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

// UnOp represents a unary operation: "+", "-", "not"
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
	case "not":
		if val == 0 {
			return 1
		}
		return 0
	}
	panic(fmt.Sprintf("[Semantic] Unknown unary operator: %s", n.value))
}

// BinOp represents a binary operation: "+", "-", "*", "/", "**", "==", ">", "<", "and", "or"
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
	case "==":
		if left == right {
			return 1
		}
		return 0
	case ">":
		if left > right {
			return 1
		}
		return 0
	case "<":
		if left < right {
			return 1
		}
		return 0
	case "and":
		if left != 0 && right != 0 {
			return 1
		}
		return 0
	case "or":
		if left != 0 || right != 0 {
			return 1
		}
		return 0
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

// IfNode: children[0]=condition, children[1]=thenBlock, children[2]=elseBlock (optional)
type IfNode struct {
	children []Node
}

func (n *IfNode) Evaluate(st *SymbolTable) int {
	cond := n.children[0].Evaluate(st)
	if cond != 0 {
		n.children[1].Evaluate(st)
	} else if len(n.children) > 2 {
		n.children[2].Evaluate(st)
	}
	return 0
}

// WhileNode: children[0]=condition, children[1]=body
type WhileNode struct {
	children []Node
}

func (n *WhileNode) Evaluate(st *SymbolTable) int {
	for n.children[0].Evaluate(st) != 0 {
		n.children[1].Evaluate(st)
	}
	return 0
}

// ForNode: numeric for loop
// children: [start, limit, body] or [start, limit, step, body]
type ForNode struct {
	varName  string
	children []Node
}

func (n *ForNode) Evaluate(st *SymbolTable) int {
	start := n.children[0].Evaluate(st)
	limit := n.children[1].Evaluate(st)
	step := 1
	body := n.children[2]
	if len(n.children) == 4 {
		step = n.children[2].Evaluate(st)
		body = n.children[3]
	}
	if step == 0 {
		panic("[Semantic] 'for' step cannot be zero")
	}
	i := start
	for ; (step > 0 && i <= limit) || (step < 0 && i >= limit); i += step {
		st.Set(n.varName, i)
		body.Evaluate(st)
	}
	// leave loop variable at its final (out-of-range) value, like Lua
	st.Set(n.varName, i)
	return 0
}

// IfExpr is a ternary-like inline expression: if COND then EXPR else EXPR end
type IfExpr struct {
	children []Node // [condition, thenExpr, elseExpr]
}

func (n *IfExpr) Evaluate(st *SymbolTable) int {
	if n.children[0].Evaluate(st) != 0 {
		return n.children[1].Evaluate(st)
	}
	return n.children[2].Evaluate(st)
}

// RepeatNode: repeat ... until condition
// children[0]=body, children[1]=condition
type RepeatNode struct {
	children []Node
}

func (n *RepeatNode) Evaluate(st *SymbolTable) int {
	for {
		n.children[0].Evaluate(st)
		if n.children[1].Evaluate(st) != 0 {
			break
		}
	}
	return 0
}

// ReadVal reads an integer from stdin and returns it as an expression value
type ReadVal struct{}

var stdinReader = bufio.NewReader(os.Stdin)

func (n *ReadVal) Evaluate(st *SymbolTable) int {
	var val int
	fmt.Fscan(stdinReader, &val)
	return val
}

// ── Parser ────────────────────────────────────────────────────────────────────

// parseAtom parses: "(" BOOLEXPRESSION ")" | "if" BOOLEXPR "then" EXPR "else" EXPR "end" | NUMBER | IDENTIFIER
func parseAtom(l *Lexer) Node {
	if l.Next.Type == IF {
		l.selectNext()
		cond := parseBoolExpr(l)
		if l.Next.Type != THEN {
			panic(fmt.Sprintf("[Parser] Expected 'then' in inline if but got %s", l.Next.Type))
		}
		l.selectNext()
		thenExpr := parseBoolExpr(l)
		if l.Next.Type != ELSE {
			panic(fmt.Sprintf("[Parser] Expected 'else' in inline if but got %s", l.Next.Type))
		}
		l.selectNext()
		elseExpr := parseBoolExpr(l)
		if l.Next.Type != KW_END {
			panic(fmt.Sprintf("[Parser] Expected 'end' to close inline if but got %s", l.Next.Type))
		}
		l.selectNext()
		return &IfExpr{children: []Node{cond, thenExpr, elseExpr}}
	}
	if l.Next.Type == OPEN_PAR {
		l.selectNext()
		result := parseBoolExpr(l)
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
	if l.Next.Type == IDEN {
		name := l.Next.Value
		l.selectNext()
		return &Identifier{value: name}
	}
	panic(fmt.Sprintf("[Parser] Unexpected token %s in atom", l.Next.Type))
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

// parseFactor parses: ("+" | "-") FACTOR | READ "()" | POWER
func parseFactor(l *Lexer) Node {
	if l.Next.Type == PLUS {
		l.selectNext()
		return &UnOp{value: "+", children: []Node{parseFactor(l)}}
	}
	if l.Next.Type == MINUS {
		l.selectNext()
		return &UnOp{value: "-", children: []Node{parseFactor(l)}}
	}
	if l.Next.Type == READ {
		l.selectNext()
		if l.Next.Type != OPEN_PAR {
			panic(fmt.Sprintf("[Parser] Expected '(' after 'read' but got %s", l.Next.Type))
		}
		l.selectNext()
		if l.Next.Type != CLOSE_PAR {
			panic(fmt.Sprintf("[Parser] Expected ')' in 'read()' but got %s", l.Next.Type))
		}
		l.selectNext()
		return &ReadVal{}
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

// parseRelExpr parses: EXPRESSION [ ("==" | ">" | "<") EXPRESSION ]
func parseRelExpr(l *Lexer) Node {
	result := parseExpression(l)
	if l.Next.Type == EQ || l.Next.Type == GT || l.Next.Type == LT {
		op := l.Next.Value
		l.selectNext()
		result = &BinOp{value: op, children: []Node{result, parseExpression(l)}}
	}
	return result
}

// parseNotExpr parses: "not" NOTEXPR | RELEXPR
func parseNotExpr(l *Lexer) Node {
	if l.Next.Type == NOT {
		l.selectNext()
		return &UnOp{value: "not", children: []Node{parseNotExpr(l)}}
	}
	return parseRelExpr(l)
}

// parseBoolTerm parses: NOTEXPR { "and" NOTEXPR }
func parseBoolTerm(l *Lexer) Node {
	result := parseNotExpr(l)
	for l.Next.Type == AND {
		l.selectNext()
		result = &BinOp{value: "and", children: []Node{result, parseNotExpr(l)}}
	}
	return result
}

// parseBoolExpr parses: BOOLTERM { "or" BOOLTERM }
func parseBoolExpr(l *Lexer) Node {
	result := parseBoolTerm(l)
	for l.Next.Type == OR {
		l.selectNext()
		result = &BinOp{value: "or", children: []Node{result, parseBoolTerm(l)}}
	}
	return result
}

// parseBlock parses statements until "end", "else", "until", or EOF
func parseBlock(l *Lexer) Node {
	var children []Node
	for l.Next.Type != EOF && l.Next.Type != KW_END && l.Next.Type != ELSE && l.Next.Type != UNTIL {
		children = append(children, parseStatement(l))
	}
	return &Block{children: children}
}

// parseStatement parses a single statement followed by a newline
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
		expr := parseBoolExpr(l)
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
		expr := parseBoolExpr(l)
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
		expr := parseBoolExpr(l)
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

	if l.Next.Type == DO {
		l.selectNext()
		if l.Next.Type != END {
			panic(fmt.Sprintf("[Parser] Expected newline after 'do' but got %s", l.Next.Type))
		}
		l.selectNext()
		block := parseBlock(l)
		if l.Next.Type != KW_END {
			panic(fmt.Sprintf("[Parser] Expected 'end' to close 'do' but got %s", l.Next.Type))
		}
		l.selectNext()
		if l.Next.Type != END {
			panic(fmt.Sprintf("[Parser] Expected newline after 'end' but got %s", l.Next.Type))
		}
		l.selectNext()
		return block
	}

	if l.Next.Type == IF {
		l.selectNext()
		cond := parseBoolExpr(l)
		if l.Next.Type != THEN {
			panic(fmt.Sprintf("[Parser] Expected 'then' but got %s", l.Next.Type))
		}
		l.selectNext()
		if l.Next.Type != END {
			panic(fmt.Sprintf("[Parser] Expected newline after 'then' but got %s", l.Next.Type))
		}
		l.selectNext()
		thenBlock := parseBlock(l)
		children := []Node{cond, thenBlock}
		if l.Next.Type == ELSE {
			l.selectNext()
			if l.Next.Type != END {
				panic(fmt.Sprintf("[Parser] Expected newline after 'else' but got %s", l.Next.Type))
			}
			l.selectNext()
			elseBlock := parseBlock(l)
			children = append(children, elseBlock)
		}
		if l.Next.Type != KW_END {
			panic(fmt.Sprintf("[Parser] Expected 'end' to close 'if' but got %s", l.Next.Type))
		}
		l.selectNext()
		if l.Next.Type != END {
			panic(fmt.Sprintf("[Parser] Expected newline after 'end' but got %s", l.Next.Type))
		}
		l.selectNext()
		return &IfNode{children: children}
	}

	if l.Next.Type == WHILE {
		l.selectNext()
		cond := parseBoolExpr(l)
		if l.Next.Type != DO {
			panic(fmt.Sprintf("[Parser] Expected 'do' but got %s", l.Next.Type))
		}
		l.selectNext()
		if l.Next.Type != END {
			panic(fmt.Sprintf("[Parser] Expected newline after 'do' but got %s", l.Next.Type))
		}
		l.selectNext()
		body := parseBlock(l)
		if l.Next.Type != KW_END {
			panic(fmt.Sprintf("[Parser] Expected 'end' to close 'while' but got %s", l.Next.Type))
		}
		l.selectNext()
		if l.Next.Type != END {
			panic(fmt.Sprintf("[Parser] Expected newline after 'end' but got %s", l.Next.Type))
		}
		l.selectNext()
		return &WhileNode{children: []Node{cond, body}}
	}

	if l.Next.Type == FOR {
		l.selectNext()
		if l.Next.Type != IDEN {
			panic(fmt.Sprintf("[Parser] Expected identifier after 'for' but got %s", l.Next.Type))
		}
		varName := l.Next.Value
		l.selectNext()
		if l.Next.Type != ASSIGN {
			panic(fmt.Sprintf("[Parser] Expected '=' in 'for' but got %s", l.Next.Type))
		}
		l.selectNext()
		start := parseExpression(l)
		if l.Next.Type != COMMA {
			panic(fmt.Sprintf("[Parser] Expected ',' after start in 'for' but got %s", l.Next.Type))
		}
		l.selectNext()
		limit := parseExpression(l)
		var forChildren []Node
		if l.Next.Type == COMMA {
			l.selectNext()
			step := parseExpression(l)
			forChildren = []Node{start, limit, step}
		} else {
			forChildren = []Node{start, limit}
		}
		if l.Next.Type != DO {
			panic(fmt.Sprintf("[Parser] Expected 'do' in 'for' but got %s", l.Next.Type))
		}
		l.selectNext()
		if l.Next.Type != END {
			panic(fmt.Sprintf("[Parser] Expected newline after 'do' but got %s", l.Next.Type))
		}
		l.selectNext()
		body := parseBlock(l)
		if l.Next.Type != KW_END {
			panic(fmt.Sprintf("[Parser] Expected 'end' to close 'for' but got %s", l.Next.Type))
		}
		l.selectNext()
		if l.Next.Type != END {
			panic(fmt.Sprintf("[Parser] Expected newline after 'end' but got %s", l.Next.Type))
		}
		l.selectNext()
		forChildren = append(forChildren, body)
		return &ForNode{varName: varName, children: forChildren}
	}

	if l.Next.Type == REPEAT {
		l.selectNext()
		if l.Next.Type != END {
			panic(fmt.Sprintf("[Parser] Expected newline after 'repeat' but got %s", l.Next.Type))
		}
		l.selectNext()
		body := parseBlock(l)
		if l.Next.Type != UNTIL {
			panic(fmt.Sprintf("[Parser] Expected 'until' to close 'repeat' but got %s", l.Next.Type))
		}
		l.selectNext()
		cond := parseBoolExpr(l)
		if l.Next.Type != END {
			panic(fmt.Sprintf("[Parser] Expected newline after 'until' condition but got %s", l.Next.Type))
		}
		l.selectNext()
		return &RepeatNode{children: []Node{body, cond}}
	}

	if l.Next.Type == END {
		l.selectNext()
		return &NoOp{}
	}

	panic(fmt.Sprintf("[Parser] Unexpected token %s", l.Next.Type))
}

// parseProgram parses: { STATEMENT } EOF
func parseProgram(l *Lexer) Node {
	block := parseBlock(l)
	if l.Next.Type != EOF {
		panic(fmt.Sprintf("[Parser] Unexpected token %s (expected EOF)", l.Next.Type))
	}
	return block
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
