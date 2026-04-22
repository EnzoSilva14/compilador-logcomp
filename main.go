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

// ── Token types ───────────────────────────────────────────────────────────────

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
	END       = "END"    // newline
	KW_END    = "KW_END" // "end" keyword
	PRINT     = "PRINT"
	IMUT      = "IMUT"
	IDEN      = "IDEN"
	EOF       = "EOF"
	AND       = "AND"
	OR        = "OR"
	NOT       = "NOT"
	EQ        = "EQ" // ==
	GT        = "GT" // >
	LT        = "LT" // <
	IF        = "IF"
	WHILE     = "WHILE"
	ELSE      = "ELSE"
	READ      = "READ"
	THEN      = "THEN"
	DO        = "DO"
	FOR       = "FOR"
	REPEAT    = "REPEAT"
	UNTIL     = "UNTIL"
	COMMA     = "COMMA"
	// v2.2
	VAR  = "VAR"  // "local"
	BOOL = "BOOL" // "true" / "false"
	STR  = "STR"  // string literal
	TYPE = "TYPE" // "number" / "string" / "boolean"
)

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
	l := &Lexer{source: source}
	l.selectNext()
	return l
}

func (l *Lexer) selectNext() {
	for l.position < len(l.source) &&
		(l.source[l.position] == ' ' || l.source[l.position] == '\t' || l.source[l.position] == '\r') {
		l.position++
	}
	if l.position >= len(l.source) {
		l.Next = Token{Type: EOF}
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

	if unicode.IsLetter(ch) || ch == '_' {
		start := l.position
		for l.position < len(l.source) &&
			(unicode.IsLetter(rune(l.source[l.position])) ||
				unicode.IsDigit(rune(l.source[l.position])) ||
				l.source[l.position] == '_') {
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
		case "local":
			l.Next = Token{Type: VAR, Value: word}
		case "true", "false":
			l.Next = Token{Type: BOOL, Value: word}
		case "number", "string", "boolean":
			l.Next = Token{Type: TYPE, Value: word}
		default:
			l.Next = Token{Type: IDEN, Value: word}
		}
		return
	}

	switch ch {
	case '"':
		l.position++
		start := l.position
		for l.position < len(l.source) && l.source[l.position] != '"' {
			l.position++
		}
		if l.position >= len(l.source) {
			panic("[Lexer] Unterminated string literal")
		}
		str := l.source[start:l.position]
		l.position++ // consume closing "
		l.Next = Token{Type: STR, Value: str}
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

type PrePro struct{}

func (p PrePro) Filter(code string) string {
	lines := strings.Split(code, "\n")
	for i, line := range lines {
		if idx := strings.Index(line, "--"); idx >= 0 {
			lines[i] = line[:idx]
		}
	}
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
	code = strings.Join(remaining, "\n")
	for name, val := range constMap {
		re := regexp.MustCompile(`\b` + name + `\b`)
		code = re.ReplaceAllString(code, val)
	}
	return code
}

// ── Variable & SymbolTable ────────────────────────────────────────────────────

type variable struct {
	intVal  int
	strVal  string
	vartype string // "number", "string", "boolean"
	immut   bool
}

func mkNumber(v int) variable    { return variable{intVal: v, vartype: "number"} }
func mkString(s string) variable { return variable{strVal: s, vartype: "string"} }
func mkBool(b int) variable      { return variable{intVal: b, vartype: "boolean"} }

func defaultFor(vartype string) variable {
	switch vartype {
	case "string":
		return mkString("")
	case "boolean":
		return mkBool(0)
	default:
		return mkNumber(0)
	}
}

func truthy(v variable) bool {
	if v.vartype == "string" {
		return v.strVal != ""
	}
	return v.intVal != 0
}

type SymbolTable struct {
	table map[string]variable
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{table: make(map[string]variable)}
}

func (st *SymbolTable) Get(name string) variable {
	v, ok := st.table[name]
	if !ok {
		panic(fmt.Sprintf("[Semantic] Undefined variable: %s", name))
	}
	return v
}

// Set assigns val to name. If name not yet declared, auto-declares (backward compat).
// If declared, enforces type match.
func (st *SymbolTable) Set(name string, val variable) {
	existing, ok := st.table[name]
	if !ok {
		// auto-declare (old-style assignment)
		st.table[name] = val
		return
	}
	if existing.immut {
		panic(fmt.Sprintf("[Semantic] cannot change the value of %s", name))
	}
	if existing.vartype != val.vartype {
		panic(fmt.Sprintf("[Semantic] Type mismatch: cannot assign %s to %s variable '%s'",
			val.vartype, existing.vartype, name))
	}
	st.table[name] = variable{intVal: val.intVal, strVal: val.strVal, vartype: existing.vartype, immut: false}
}

func (st *SymbolTable) SetImut(name string, val variable) {
	v := val
	v.immut = true
	st.table[name] = v
}

// CreateVariable declares a variable with a given type and default value.
func (st *SymbolTable) CreateVariable(name string, vartype string) {
	if _, ok := st.table[name]; ok {
		panic(fmt.Sprintf("[Semantic] Variable '%s' already declared", name))
	}
	st.table[name] = defaultFor(vartype)
}

// ── AST ───────────────────────────────────────────────────────────────────────

type Node interface {
	Evaluate(st *SymbolTable) variable
}

// IntVal
type IntVal struct{ value int }

func (n *IntVal) Evaluate(_ *SymbolTable) variable { return mkNumber(n.value) }

// BoolVal
type BoolVal struct{ value int } // 1=true, 0=false

func (n *BoolVal) Evaluate(_ *SymbolTable) variable { return mkBool(n.value) }

// StringVal
type StringVal struct{ value string }

func (n *StringVal) Evaluate(_ *SymbolTable) variable { return mkString(n.value) }

// Identifier
type Identifier struct{ value string }

func (n *Identifier) Evaluate(st *SymbolTable) variable { return st.Get(n.value) }

// UnOp
type UnOp struct {
	value    string
	children []Node
}

func (n *UnOp) Evaluate(st *SymbolTable) variable {
	val := n.children[0].Evaluate(st)
	switch n.value {
	case "+":
		if val.vartype != "number" {
			panic("[Semantic] Unary '+' requires number")
		}
		return mkNumber(val.intVal)
	case "-":
		if val.vartype != "number" {
			panic("[Semantic] Unary '-' requires number")
		}
		return mkNumber(-val.intVal)
	case "not":
		if truthy(val) {
			return mkBool(0)
		}
		return mkBool(1)
	}
	panic(fmt.Sprintf("[Semantic] Unknown unary operator: %s", n.value))
}

// BinOp
type BinOp struct {
	value    string
	children []Node
}

func requireNumbers(l, r variable, op string) {
	if l.vartype != "number" || r.vartype != "number" {
		panic(fmt.Sprintf("[Semantic] Operator '%s' requires number operands", op))
	}
}

func (n *BinOp) Evaluate(st *SymbolTable) variable {
	left := n.children[0].Evaluate(st)
	right := n.children[1].Evaluate(st)
	switch n.value {
	case "+":
		if left.vartype == "number" && right.vartype == "number" {
			return mkNumber(left.intVal + right.intVal)
		}
		if left.vartype == "string" && right.vartype == "string" {
			return mkString(left.strVal + right.strVal)
		}
		panic("[Semantic] Operator '+' requires both operands to be number or both string")
	case "-":
		requireNumbers(left, right, "-")
		return mkNumber(left.intVal - right.intVal)
	case "*":
		requireNumbers(left, right, "*")
		return mkNumber(left.intVal * right.intVal)
	case "/":
		requireNumbers(left, right, "/")
		if right.intVal == 0 {
			panic("[Semantic] Division by zero")
		}
		return mkNumber(left.intVal / right.intVal)
	case "**":
		requireNumbers(left, right, "**")
		result := 1
		for i := 0; i < right.intVal; i++ {
			result *= left.intVal
		}
		return mkNumber(result)
	case "==":
		if left.vartype != right.vartype {
			panic(fmt.Sprintf("[Semantic] Type mismatch in '==': %s vs %s", left.vartype, right.vartype))
		}
		if left.vartype == "string" {
			if left.strVal == right.strVal {
				return mkBool(1)
			}
			return mkBool(0)
		}
		if left.intVal == right.intVal {
			return mkBool(1)
		}
		return mkBool(0)
	case ">":
		requireNumbers(left, right, ">")
		if left.intVal > right.intVal {
			return mkBool(1)
		}
		return mkBool(0)
	case "<":
		requireNumbers(left, right, "<")
		if left.intVal < right.intVal {
			return mkBool(1)
		}
		return mkBool(0)
	case "and":
		if truthy(left) && truthy(right) {
			return mkBool(1)
		}
		return mkBool(0)
	case "or":
		if truthy(left) || truthy(right) {
			return mkBool(1)
		}
		return mkBool(0)
	}
	panic(fmt.Sprintf("[Semantic] Unknown binary operator: %s", n.value))
}

// Assignment
type Assignment struct {
	children []Node // [Identifier, Expr]
}

func (n *Assignment) Evaluate(st *SymbolTable) variable {
	name := n.children[0].(*Identifier).value
	val := n.children[1].Evaluate(st)
	st.Set(name, val)
	return mkNumber(0)
}

// ImutAssignment
type ImutAssignment struct {
	children []Node
}

func (n *ImutAssignment) Evaluate(st *SymbolTable) variable {
	name := n.children[0].(*Identifier).value
	val := n.children[1].Evaluate(st)
	st.SetImut(name, val)
	return mkNumber(0)
}

// VarDec: "local" IDEN TYPE ["=" EXPR]
type VarDec struct {
	vartype  string
	children []Node // [Identifier] or [Identifier, Expr]
}

func (n *VarDec) Evaluate(st *SymbolTable) variable {
	name := n.children[0].(*Identifier).value
	st.CreateVariable(name, n.vartype)
	if len(n.children) > 1 {
		val := n.children[1].Evaluate(st)
		st.Set(name, val)
	}
	return mkNumber(0)
}

// Print
type Print struct {
	children []Node
}

func (n *Print) Evaluate(st *SymbolTable) variable {
	val := n.children[0].Evaluate(st)
	switch val.vartype {
	case "string":
		fmt.Println(val.strVal)
	case "boolean":
		if val.intVal != 0 {
			fmt.Println("True")
		} else {
			fmt.Println("False")
		}
	default:
		fmt.Println(val.intVal)
	}
	return mkNumber(0)
}

// Block
type Block struct {
	children []Node
}

func (n *Block) Evaluate(st *SymbolTable) variable {
	for _, child := range n.children {
		child.Evaluate(st)
	}
	return mkNumber(0)
}

// NoOp
type NoOp struct{}

func (n *NoOp) Evaluate(_ *SymbolTable) variable { return mkNumber(0) }

// IfNode
type IfNode struct {
	children []Node // [cond, thenBlock] or [cond, thenBlock, elseBlock]
}

func (n *IfNode) Evaluate(st *SymbolTable) variable {
	if truthy(n.children[0].Evaluate(st)) {
		n.children[1].Evaluate(st)
	} else if len(n.children) > 2 {
		n.children[2].Evaluate(st)
	}
	return mkNumber(0)
}

// IfExpr: inline "if cond then expr else expr end"
type IfExpr struct {
	children []Node // [cond, thenExpr, elseExpr]
}

func (n *IfExpr) Evaluate(st *SymbolTable) variable {
	if truthy(n.children[0].Evaluate(st)) {
		return n.children[1].Evaluate(st)
	}
	return n.children[2].Evaluate(st)
}

// WhileNode
type WhileNode struct {
	children []Node // [cond, body]
}

func (n *WhileNode) Evaluate(st *SymbolTable) variable {
	for truthy(n.children[0].Evaluate(st)) {
		n.children[1].Evaluate(st)
	}
	return mkNumber(0)
}

// ForNode
type ForNode struct {
	varName  string
	children []Node // [start, limit, body] or [start, limit, step, body]
}

func (n *ForNode) Evaluate(st *SymbolTable) variable {
	start := n.children[0].Evaluate(st).intVal
	limit := n.children[1].Evaluate(st).intVal
	step := 1
	body := n.children[2]
	if len(n.children) == 4 {
		step = n.children[2].Evaluate(st).intVal
		body = n.children[3]
	}
	if step == 0 {
		panic("[Semantic] 'for' step cannot be zero")
	}
	i := start
	for ; (step > 0 && i <= limit) || (step < 0 && i >= limit); i += step {
		st.Set(n.varName, mkNumber(i))
		body.Evaluate(st)
	}
	st.Set(n.varName, mkNumber(i))
	return mkNumber(0)
}

// RepeatNode
type RepeatNode struct {
	children []Node // [body, condition]
}

func (n *RepeatNode) Evaluate(st *SymbolTable) variable {
	for {
		n.children[0].Evaluate(st)
		if truthy(n.children[1].Evaluate(st)) {
			break
		}
	}
	return mkNumber(0)
}

// ReadVal
type ReadVal struct{}

var stdinReader = bufio.NewReader(os.Stdin)

func (n *ReadVal) Evaluate(_ *SymbolTable) variable {
	var val int
	fmt.Fscan(stdinReader, &val)
	return mkNumber(val)
}

// ── Parser ────────────────────────────────────────────────────────────────────

// Forward declarations handled by Go's package-level functions.

func parseAtom(l *Lexer) Node {
	// Inline if expression
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
	if l.Next.Type == BOOL {
		v := 0
		if l.Next.Value == "true" {
			v = 1
		}
		l.selectNext()
		return &BoolVal{value: v}
	}
	if l.Next.Type == STR {
		s := l.Next.Value
		l.selectNext()
		return &StringVal{value: s}
	}
	panic(fmt.Sprintf("[Parser] Unexpected token %s in atom", l.Next.Type))
}

func parsePower(l *Lexer) Node {
	base := parseAtom(l)
	if l.Next.Type == POW {
		l.selectNext()
		return &BinOp{value: "**", children: []Node{base, parseFactor(l)}}
	}
	return base
}

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

func parseTerm(l *Lexer) Node {
	result := parseFactor(l)
	for l.Next.Type == MULT || l.Next.Type == DIV {
		op := l.Next.Value
		l.selectNext()
		result = &BinOp{value: op, children: []Node{result, parseFactor(l)}}
	}
	return result
}

func parseExpression(l *Lexer) Node {
	result := parseTerm(l)
	for l.Next.Type == PLUS || l.Next.Type == MINUS {
		op := l.Next.Value
		l.selectNext()
		result = &BinOp{value: op, children: []Node{result, parseTerm(l)}}
	}
	return result
}

func parseRelExpr(l *Lexer) Node {
	result := parseExpression(l)
	if l.Next.Type == EQ || l.Next.Type == GT || l.Next.Type == LT {
		op := l.Next.Value
		l.selectNext()
		result = &BinOp{value: op, children: []Node{result, parseExpression(l)}}
	}
	return result
}

func parseNotExpr(l *Lexer) Node {
	if l.Next.Type == NOT {
		l.selectNext()
		return &UnOp{value: "not", children: []Node{parseNotExpr(l)}}
	}
	return parseRelExpr(l)
}

func parseBoolTerm(l *Lexer) Node {
	result := parseNotExpr(l)
	for l.Next.Type == AND {
		l.selectNext()
		result = &BinOp{value: "and", children: []Node{result, parseNotExpr(l)}}
	}
	return result
}

func parseBoolExpr(l *Lexer) Node {
	result := parseBoolTerm(l)
	for l.Next.Type == OR {
		l.selectNext()
		result = &BinOp{value: "or", children: []Node{result, parseBoolTerm(l)}}
	}
	return result
}

func parseBlock(l *Lexer) Node {
	var children []Node
	for l.Next.Type != EOF && l.Next.Type != KW_END &&
		l.Next.Type != ELSE && l.Next.Type != UNTIL {
		children = append(children, parseStatement(l))
	}
	return &Block{children: children}
}

func parseStatement(l *Lexer) Node {
	// local declaration
	if l.Next.Type == VAR {
		l.selectNext()
		if l.Next.Type != IDEN {
			panic(fmt.Sprintf("[Parser] Expected identifier after 'local' but got %s", l.Next.Type))
		}
		name := l.Next.Value
		l.selectNext()
		if l.Next.Type != TYPE {
			panic(fmt.Sprintf("[Parser] Expected type after identifier in 'local' but got %s", l.Next.Type))
		}
		vartype := l.Next.Value
		l.selectNext()
		children := []Node{&Identifier{value: name}}
		if l.Next.Type == ASSIGN {
			l.selectNext()
			children = append(children, parseBoolExpr(l))
		}
		if l.Next.Type != END {
			panic(fmt.Sprintf("[Parser] Expected newline but got %s", l.Next.Type))
		}
		l.selectNext()
		return &VarDec{vartype: vartype, children: children}
	}

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
			children = append(children, parseBlock(l))
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
		forChildren := []Node{start, limit}
		if l.Next.Type == COMMA {
			l.selectNext()
			forChildren = append(forChildren, parseExpression(l))
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

func parseProgram(l *Lexer) Node {
	block := parseBlock(l)
	if l.Next.Type != EOF {
		panic(fmt.Sprintf("[Parser] Unexpected token %s (expected EOF)", l.Next.Type))
	}
	return block
}

func run(source string) Node {
	return parseProgram(NewLexer(source))
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
	run(source).Evaluate(NewSymbolTable())
}
