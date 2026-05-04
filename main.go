package main

import (
	"bufio"
	"fmt"
	"math"
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
	FOR    = "FOR"
	REPEAT = "REPEAT"
	UNTIL  = "UNTIL"
	COMMA  = "COMMA"
	CONCAT    = "CONCAT"    // ..
	FLOAT_LIT = "FLOAT_LIT" // float literal
	// v2.2
	VAR  = "VAR"  // "local"
	BOOL = "BOOL" // "true" / "false"
	STR  = "STR"  // string literal
	TYPE = "TYPE" // "number" / "string" / "boolean" / "float"
	// v2.3
	FUNCTION = "FUNCTION" // "function"
	RETURN   = "RETURN"   // "return"
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

type LexerState struct {
	position int
	next     Token
}

func (l *Lexer) Save() LexerState    { return LexerState{l.position, l.Next} }
func (l *Lexer) Restore(s LexerState) { l.position = s.position; l.Next = s.next }

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
		// Float: digit '.' digit (but not '..' concat)
		if l.position < len(l.source) && l.source[l.position] == '.' &&
			!(l.position+1 < len(l.source) && l.source[l.position+1] == '.') {
			l.position++ // consume '.'
			for l.position < len(l.source) && unicode.IsDigit(rune(l.source[l.position])) {
				l.position++
			}
			l.Next = Token{Type: FLOAT_LIT, Value: l.source[start:l.position]}
		} else {
			l.Next = Token{Type: INT, Value: l.source[start:l.position]}
		}
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
		case "number", "string", "boolean", "float":
			l.Next = Token{Type: TYPE, Value: word}
		case "function":
			l.Next = Token{Type: FUNCTION, Value: word}
		case "return":
			l.Next = Token{Type: RETURN, Value: word}
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
	case '.':
		if l.position+1 < len(l.source) && l.source[l.position+1] == '.' {
			l.Next = Token{Type: CONCAT, Value: ".."}
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
	intVal   int
	floatVal float64
	strVal   string
	vartype  string // "number", "float", "string", "boolean"
	immut    bool
	shift    int  // stack offset for code generation (e.g. 4 means [EBP-4])
	isParam  bool // if true, access via [EBP+shift] instead of [EBP-shift]
}

func mkNumber(v int) variable      { return variable{intVal: v, vartype: "number"} }
func mkFloat(f float64) variable   { return variable{floatVal: f, vartype: "float"} }
func mkString(s string) variable   { return variable{strVal: s, vartype: "string"} }
func mkBool(b int) variable        { return variable{intVal: b, vartype: "boolean"} }

func isNumeric(v variable) bool {
	return v.vartype == "number" || v.vartype == "float"
}

func toFloat(v variable) float64 {
	if v.vartype == "float" {
		return v.floatVal
	}
	return float64(v.intVal)
}

func defaultFor(vartype string) variable {
	switch vartype {
	case "string":
		return mkString("")
	case "boolean":
		return mkBool(0)
	case "float":
		return mkFloat(0.0)
	default:
		return mkNumber(0)
	}
}

func truthy(v variable) bool {
	switch v.vartype {
	case "string":
		return v.strVal != ""
	case "float":
		return v.floatVal != 0
	default:
		return v.intVal != 0
	}
}

func valToString(v variable) string {
	switch v.vartype {
	case "string":
		return v.strVal
	case "boolean":
		if v.intVal != 0 {
			return "true"
		}
		return "false"
	case "float":
		return strconv.FormatFloat(v.floatVal, 'f', -1, 64)
	default:
		return strconv.Itoa(v.intVal)
	}
}

type SymbolTable struct {
	table     map[string]variable
	nextShift int
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
	st.table[name] = variable{intVal: val.intVal, floatVal: val.floatVal, strVal: val.strVal, vartype: existing.vartype, immut: false}
}

func (st *SymbolTable) SetImut(name string, val variable) {
	v := val
	v.immut = true
	st.table[name] = v
}

// CreateVariable declares a variable with a given type and default value.
// It also allocates a 4-byte stack slot for code generation.
func (st *SymbolTable) CreateVariable(name string, vartype string) {
	if _, ok := st.table[name]; ok {
		panic(fmt.Sprintf("[Semantic] Variable '%s' already declared", name))
	}
	st.nextShift += 4
	v := defaultFor(vartype)
	v.shift = st.nextShift
	st.table[name] = v
}

// ── AST ───────────────────────────────────────────────────────────────────────

type Node interface {
	Evaluate(st *SymbolTable) variable
	Generate(st *SymbolTable)
}

// IntVal
type IntVal struct{ value int }

func (n *IntVal) Evaluate(_ *SymbolTable) variable { return mkNumber(n.value) }

// FloatVal
type FloatVal struct{ value float64 }

func (n *FloatVal) Evaluate(_ *SymbolTable) variable { return mkFloat(n.value) }

// CastNode: (TYPE) EXPR
type CastNode struct {
	targetType string
	children   []Node
}

func (n *CastNode) Evaluate(st *SymbolTable) variable {
	val := n.children[0].Evaluate(st)
	switch n.targetType {
	case "number":
		switch val.vartype {
		case "number":
			return mkNumber(val.intVal)
		case "float":
			return mkNumber(int(math.Round(val.floatVal)))
		case "boolean":
			return mkNumber(val.intVal)
		case "string":
			i, err := strconv.Atoi(val.strVal)
			if err != nil {
				panic(fmt.Sprintf("[Semantic] Cannot cast string '%s' to number", val.strVal))
			}
			return mkNumber(i)
		}
	case "float":
		switch val.vartype {
		case "float":
			return mkFloat(val.floatVal)
		case "number":
			return mkFloat(float64(val.intVal))
		case "boolean":
			return mkFloat(float64(val.intVal))
		case "string":
			f, err := strconv.ParseFloat(val.strVal, 64)
			if err != nil {
				panic(fmt.Sprintf("[Semantic] Cannot cast string '%s' to float", val.strVal))
			}
			return mkFloat(f)
		}
	case "string":
		return mkString(valToString(val))
	case "boolean":
		if truthy(val) {
			return mkBool(1)
		}
		return mkBool(0)
	}
	panic(fmt.Sprintf("[Semantic] Unknown cast target type: %s", n.targetType))
}

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
		if !isNumeric(val) {
			panic("[Semantic] Unary '+' requires number or float")
		}
		if val.vartype == "float" {
			return mkFloat(val.floatVal)
		}
		return mkNumber(val.intVal)
	case "-":
		if !isNumeric(val) {
			panic("[Semantic] Unary '-' requires number or float")
		}
		if val.vartype == "float" {
			return mkFloat(-val.floatVal)
		}
		return mkNumber(-val.intVal)
	case "not":
		if val.vartype != "boolean" {
			panic("[Semantic] 'not' requires a boolean operand")
		}
		if val.intVal != 0 {
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

func requireNumeric(l, r variable, op string) {
	if !isNumeric(l) || !isNumeric(r) {
		panic(fmt.Sprintf("[Semantic] Operator '%s' requires number or float operands", op))
	}
}

func numericResult(l, r variable, iResult int, fResult float64) variable {
	if l.vartype == "float" || r.vartype == "float" {
		return mkFloat(fResult)
	}
	return mkNumber(iResult)
}

func (n *BinOp) Evaluate(st *SymbolTable) variable {
	left := n.children[0].Evaluate(st)
	right := n.children[1].Evaluate(st)
	switch n.value {
	case "+":
		requireNumeric(left, right, "+")
		return numericResult(left, right, left.intVal+right.intVal, toFloat(left)+toFloat(right))
	case "-":
		requireNumeric(left, right, "-")
		return numericResult(left, right, left.intVal-right.intVal, toFloat(left)-toFloat(right))
	case "*":
		requireNumeric(left, right, "*")
		return numericResult(left, right, left.intVal*right.intVal, toFloat(left)*toFloat(right))
	case "/":
		requireNumeric(left, right, "/")
		if left.vartype == "float" || right.vartype == "float" {
			if toFloat(right) == 0 {
				panic("[Semantic] Division by zero")
			}
			return mkFloat(toFloat(left) / toFloat(right))
		}
		if right.intVal == 0 {
			panic("[Semantic] Division by zero")
		}
		return mkNumber(left.intVal / right.intVal)
	case "**":
		requireNumeric(left, right, "**")
		if left.vartype == "float" || right.vartype == "float" {
			base, exp := toFloat(left), toFloat(right)
			result := 1.0
			for i := 0.0; i < exp; i++ {
				result *= base
			}
			return mkFloat(result)
		}
		result := 1
		for i := 0; i < right.intVal; i++ {
			result *= left.intVal
		}
		return mkNumber(result)
	case "..":
		return mkString(valToString(left) + valToString(right))
	case "==":
		if isNumeric(left) && isNumeric(right) {
			if toFloat(left) == toFloat(right) {
				return mkBool(1)
			}
			return mkBool(0)
		}
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
		if left.vartype == "string" && right.vartype == "string" {
			if left.strVal > right.strVal {
				return mkBool(1)
			}
			return mkBool(0)
		}
		requireNumeric(left, right, ">")
		if toFloat(left) > toFloat(right) {
			return mkBool(1)
		}
		return mkBool(0)
	case "<":
		if left.vartype == "string" && right.vartype == "string" {
			if left.strVal < right.strVal {
				return mkBool(1)
			}
			return mkBool(0)
		}
		requireNumeric(left, right, "<")
		if toFloat(left) < toFloat(right) {
			return mkBool(1)
		}
		return mkBool(0)
	case "and":
		if left.vartype != "boolean" || right.vartype != "boolean" {
			panic("[Semantic] 'and' requires boolean operands")
		}
		if left.intVal != 0 && right.intVal != 0 {
			return mkBool(1)
		}
		return mkBool(0)
	case "or":
		if left.vartype != "boolean" || right.vartype != "boolean" {
			panic("[Semantic] 'or' requires boolean operands")
		}
		if left.intVal != 0 || right.intVal != 0 {
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
			fmt.Println("true")
		} else {
			fmt.Println("false")
		}
	case "float":
		fmt.Println(strconv.FormatFloat(val.floatVal, 'f', -1, 64))
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
	cond := n.children[0].Evaluate(st)
	if cond.vartype != "boolean" {
		panic("[Semantic] 'if' condition must be boolean")
	}
	if truthy(cond) {
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
	for {
		cond := n.children[0].Evaluate(st)
		if cond.vartype != "boolean" {
			panic("[Semantic] 'while' condition must be boolean")
		}
		if !truthy(cond) {
			break
		}
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

// ── Functions ─────────────────────────────────────────────────────────────────

// globalFuncTable stores all declared functions by name.
var globalFuncTable = map[string]*FuncDef{}

// returnValue is used as a panic payload to propagate return values.
type returnValue struct{ val variable }

// FuncDef: "function" NAME "(" params ")" "\n" BLOCK "end"
type FuncDef struct {
	name   string
	params []string
	body   Node
}

func (n *FuncDef) Evaluate(st *SymbolTable) variable {
	globalFuncTable[n.name] = n
	return mkNumber(0)
}

// FuncCall: NAME "(" args ")"
type FuncCall struct {
	name     string
	children []Node // arguments
}

func (n *FuncCall) Evaluate(st *SymbolTable) variable {
	def, ok := globalFuncTable[n.name]
	if !ok {
		panic(fmt.Sprintf("[Semantic] Undefined function: %s", n.name))
	}
	if len(n.children) != len(def.params) {
		panic(fmt.Sprintf("[Semantic] Function '%s' expects %d args but got %d",
			n.name, len(def.params), len(n.children)))
	}
	funcST := NewSymbolTable()
	for i, param := range def.params {
		funcST.table[param] = n.children[i].Evaluate(st)
	}
	var result variable
	func() {
		defer func() {
			if r := recover(); r != nil {
				if rv, ok := r.(returnValue); ok {
					result = rv.val
				} else {
					panic(r)
				}
			}
		}()
		def.body.Evaluate(funcST)
	}()
	return result
}

// ReturnNode: "return" EXPR
type ReturnNode struct {
	children []Node // [expr]
}

func (n *ReturnNode) Evaluate(st *SymbolTable) variable {
	val := n.children[0].Evaluate(st)
	panic(returnValue{val})
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
	if l.Next.Type == FLOAT_LIT {
		val, _ := strconv.ParseFloat(l.Next.Value, 64)
		l.selectNext()
		return &FloatVal{value: val}
	}
	if l.Next.Type == IDEN {
		name := l.Next.Value
		l.selectNext()
		if l.Next.Type == OPEN_PAR {
			l.selectNext() // consume '('
			var args []Node
			if l.Next.Type != CLOSE_PAR {
				args = append(args, parseBoolExpr(l))
				for l.Next.Type == COMMA {
					l.selectNext()
					args = append(args, parseBoolExpr(l))
				}
			}
			if l.Next.Type != CLOSE_PAR {
				panic(fmt.Sprintf("[Parser] Expected ')' in call to '%s' but got %s", name, l.Next.Type))
			}
			l.selectNext()
			return &FuncCall{name: name, children: args}
		}
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
	// Cast: (TYPE) FACTOR — use lookahead via Save/Restore
	if l.Next.Type == OPEN_PAR {
		saved := l.Save()
		l.selectNext() // consume '('
		if l.Next.Type == TYPE {
			castType := l.Next.Value
			l.selectNext() // consume TYPE
			if l.Next.Type == CLOSE_PAR {
				l.selectNext() // consume ')'
				return &CastNode{targetType: castType, children: []Node{parseFactor(l)}}
			}
		}
		l.Restore(saved) // not a cast, fall through to parsePower/parseAtom
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

// parseConcatExpr parses: EXPRESSION { ".." EXPRESSION }  (right-associative)
func parseConcatExpr(l *Lexer) Node {
	left := parseExpression(l)
	if l.Next.Type == CONCAT {
		l.selectNext()
		right := parseConcatExpr(l) // right-associative
		return &BinOp{value: "..", children: []Node{left, right}}
	}
	return left
}

func parseRelExpr(l *Lexer) Node {
	result := parseConcatExpr(l)
	if l.Next.Type == EQ || l.Next.Type == GT || l.Next.Type == LT {
		op := l.Next.Value
		l.selectNext()
		result = &BinOp{value: op, children: []Node{result, parseConcatExpr(l)}}
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
		if l.Next.Type == END {
			l.selectNext()
		}
		block := parseBlock(l)
		if l.Next.Type != KW_END {
			panic(fmt.Sprintf("[Parser] Expected 'end' to close 'do' but got %s", l.Next.Type))
		}
		l.selectNext()
		if l.Next.Type == END {
			l.selectNext()
		}
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

	if l.Next.Type == FUNCTION {
		l.selectNext()
		if l.Next.Type != IDEN {
			panic(fmt.Sprintf("[Parser] Expected function name but got %s", l.Next.Type))
		}
		name := l.Next.Value
		l.selectNext()
		if l.Next.Type != OPEN_PAR {
			panic(fmt.Sprintf("[Parser] Expected '(' after function name but got %s", l.Next.Type))
		}
		l.selectNext()
		var params []string
		if l.Next.Type != CLOSE_PAR {
			if l.Next.Type != IDEN {
				panic(fmt.Sprintf("[Parser] Expected parameter name but got %s", l.Next.Type))
			}
			params = append(params, l.Next.Value)
			l.selectNext()
			for l.Next.Type == COMMA {
				l.selectNext()
				if l.Next.Type != IDEN {
					panic(fmt.Sprintf("[Parser] Expected parameter name but got %s", l.Next.Type))
				}
				params = append(params, l.Next.Value)
				l.selectNext()
			}
		}
		if l.Next.Type != CLOSE_PAR {
			panic(fmt.Sprintf("[Parser] Expected ')' in function definition but got %s", l.Next.Type))
		}
		l.selectNext()
		if l.Next.Type != END {
			panic(fmt.Sprintf("[Parser] Expected newline after function signature but got %s", l.Next.Type))
		}
		l.selectNext()
		body := parseBlock(l)
		if l.Next.Type != KW_END {
			panic(fmt.Sprintf("[Parser] Expected 'end' to close function but got %s", l.Next.Type))
		}
		l.selectNext()
		if l.Next.Type == END {
			l.selectNext()
		}
		return &FuncDef{name: name, params: params, body: body}
	}

	if l.Next.Type == RETURN {
		l.selectNext()
		expr := parseBoolExpr(l)
		if l.Next.Type == END {
			l.selectNext()
		}
		return &ReturnNode{children: []Node{expr}}
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

// ── Code Generator ────────────────────────────────────────────────────────────

var codeInstructions []string
var funcInstructions []string
var currentBuffer = &codeInstructions

func codeAppend(line string) {
	*currentBuffer = append(*currentBuffer, line)
}

func codeDump(filename string) {
	f, err := os.Create(filename)
	if err != nil {
		panic(fmt.Sprintf("[Main] Cannot create asm file: %v", err))
	}
	defer f.Close()
	fmt.Fprintf(f, "section .data\n")
	fmt.Fprintf(f, "  format_out: db \"%%d\", 10, 0\n")
	fmt.Fprintf(f, "  format_in: db \"%%d\", 0\n")
	fmt.Fprintf(f, "  scan_int: dd 0\n\n")
	fmt.Fprintf(f, "section .text\n")
	fmt.Fprintf(f, "  extern printf\n")
	fmt.Fprintf(f, "  extern scanf\n")
	fmt.Fprintf(f, "  global _start\n\n")
	fmt.Fprintf(f, "_start:\n")
	fmt.Fprintf(f, "  push ebp\n")
	fmt.Fprintf(f, "  mov ebp, esp\n\n")
	for _, line := range codeInstructions {
		fmt.Fprintf(f, "%s\n", line)
	}
	fmt.Fprintf(f, "\n  mov esp, ebp\n")
	fmt.Fprintf(f, "  pop ebp\n")
	fmt.Fprintf(f, "  mov eax, 1\n")
	fmt.Fprintf(f, "  xor ebx, ebx\n")
	fmt.Fprintf(f, "  int 0x80\n")
	if len(funcInstructions) > 0 {
		fmt.Fprintf(f, "\n")
		for _, line := range funcInstructions {
			fmt.Fprintf(f, "%s\n", line)
		}
	}
}

var nodeIDCounter int

func newNodeID() int {
	nodeIDCounter++
	return nodeIDCounter
}

// IntVal
func (n *IntVal) Generate(_ *SymbolTable) {
	codeAppend(fmt.Sprintf("  mov eax, %d", n.value))
}

// FloatVal
func (n *FloatVal) Generate(_ *SymbolTable) {
	panic("[CodeGen] float not supported in assembly generation")
}

// BoolVal
func (n *BoolVal) Generate(_ *SymbolTable) {
	codeAppend(fmt.Sprintf("  mov eax, %d", n.value))
}

// StringVal
func (n *StringVal) Generate(_ *SymbolTable) {
	panic("[CodeGen] string not supported in assembly generation")
}

// CastNode
func (n *CastNode) Generate(_ *SymbolTable) {
	panic("[CodeGen] cast not supported in assembly generation")
}

// Identifier
func (n *Identifier) Generate(st *SymbolTable) {
	v := st.Get(n.value)
	if v.isParam {
		codeAppend(fmt.Sprintf("  mov eax, [ebp+%d]", v.shift))
	} else {
		codeAppend(fmt.Sprintf("  mov eax, [ebp-%d]", v.shift))
	}
}

// UnOp — result always in EAX
func (n *UnOp) Generate(st *SymbolTable) {
	n.children[0].Generate(st)
	switch n.value {
	case "+":
		// no-op
	case "-":
		codeAppend("  neg eax")
	case "not":
		codeAppend("  xor eax, 1")
	}
}

// BinOp — convention: generate right → push; generate left → EAX; pop ECX; op EAX, ECX
func (n *BinOp) Generate(st *SymbolTable) {
	n.children[1].Generate(st) // right → EAX
	codeAppend("  push eax")
	n.children[0].Generate(st) // left → EAX
	codeAppend("  pop ecx")    // ECX = right
	switch n.value {
	case "+":
		codeAppend("  add eax, ecx")
	case "-":
		codeAppend("  sub eax, ecx")
	case "*":
		codeAppend("  imul ecx")
	case "/":
		codeAppend("  cdq")
		codeAppend("  idiv ecx")
	case "==":
		codeAppend("  cmp eax, ecx")
		codeAppend("  mov eax, 0")
		codeAppend("  mov ecx, 1")
		codeAppend("  cmove eax, ecx")
	case ">":
		codeAppend("  cmp eax, ecx")
		codeAppend("  mov eax, 0")
		codeAppend("  mov ecx, 1")
		codeAppend("  cmovg eax, ecx")
	case "<":
		codeAppend("  cmp eax, ecx")
		codeAppend("  mov eax, 0")
		codeAppend("  mov ecx, 1")
		codeAppend("  cmovl eax, ecx")
	case "and":
		codeAppend("  and eax, ecx")
	case "or":
		codeAppend("  or eax, ecx")
	case "**", "..":
		panic(fmt.Sprintf("[CodeGen] operator '%s' not supported in assembly generation", n.value))
	default:
		panic(fmt.Sprintf("[CodeGen] unknown operator: %s", n.value))
	}
}

// Assignment
func (n *Assignment) Generate(st *SymbolTable) {
	name := n.children[0].(*Identifier).value
	n.children[1].Generate(st)
	v := st.Get(name)
	if v.isParam {
		codeAppend(fmt.Sprintf("  mov [ebp+%d], eax", v.shift))
	} else {
		codeAppend(fmt.Sprintf("  mov [ebp-%d], eax", v.shift))
	}
}

// ImutAssignment — allocates a stack slot on first encounter
func (n *ImutAssignment) Generate(st *SymbolTable) {
	name := n.children[0].(*Identifier).value
	if _, ok := st.table[name]; !ok {
		st.nextShift += 4
		v := mkNumber(0)
		v.shift = st.nextShift
		v.immut = true
		st.table[name] = v
		codeAppend(fmt.Sprintf("  sub esp, 4 ; imut %s [EBP-%d]", name, st.nextShift))
	}
	n.children[1].Generate(st)
	v := st.Get(name)
	codeAppend(fmt.Sprintf("  mov [ebp-%d], eax", v.shift))
}

// VarDec
func (n *VarDec) Generate(st *SymbolTable) {
	name := n.children[0].(*Identifier).value
	st.CreateVariable(name, n.vartype)
	v := st.table[name]
	codeAppend(fmt.Sprintf("  sub esp, 4 ; var %s [EBP-%d]", name, v.shift))
	if len(n.children) > 1 {
		n.children[1].Generate(st)
		codeAppend(fmt.Sprintf("  mov [ebp-%d], eax", v.shift))
	}
}

// Print
func (n *Print) Generate(st *SymbolTable) {
	n.children[0].Generate(st)
	codeAppend("  push eax")
	codeAppend("  push format_out")
	codeAppend("  call printf")
	codeAppend("  add esp, 8")
}

// Block
func (n *Block) Generate(st *SymbolTable) {
	for _, child := range n.children {
		child.Generate(st)
	}
}

// NoOp
func (n *NoOp) Generate(_ *SymbolTable) {}

// IfNode
func (n *IfNode) Generate(st *SymbolTable) {
	id := newNodeID()
	n.children[0].Generate(st) // condition → EAX
	if len(n.children) == 2 {
		codeAppend("  cmp eax, 0")
		codeAppend(fmt.Sprintf("  je exit_%d", id))
		n.children[1].Generate(st)
		codeAppend(fmt.Sprintf("exit_%d:", id))
	} else {
		codeAppend("  cmp eax, 0")
		codeAppend(fmt.Sprintf("  je else_%d", id))
		n.children[1].Generate(st)
		codeAppend(fmt.Sprintf("  jmp exit_%d", id))
		codeAppend(fmt.Sprintf("else_%d:", id))
		n.children[2].Generate(st)
		codeAppend(fmt.Sprintf("exit_%d:", id))
	}
}

// IfExpr (inline if)
func (n *IfExpr) Generate(st *SymbolTable) {
	id := newNodeID()
	n.children[0].Generate(st) // condition → EAX
	codeAppend("  cmp eax, 0")
	codeAppend(fmt.Sprintf("  je else_%d", id))
	n.children[1].Generate(st) // then → EAX
	codeAppend(fmt.Sprintf("  jmp exit_%d", id))
	codeAppend(fmt.Sprintf("else_%d:", id))
	n.children[2].Generate(st) // else → EAX
	codeAppend(fmt.Sprintf("exit_%d:", id))
}

// WhileNode
func (n *WhileNode) Generate(st *SymbolTable) {
	id := newNodeID()
	codeAppend(fmt.Sprintf("loop_%d:", id))
	n.children[0].Generate(st) // condition → EAX
	codeAppend("  cmp eax, 0")
	codeAppend(fmt.Sprintf("  je exit_%d", id))
	n.children[1].Generate(st) // body
	codeAppend(fmt.Sprintf("  jmp loop_%d", id))
	codeAppend(fmt.Sprintf("exit_%d:", id))
}

// ForNode — children: [start, limit, body] or [start, limit, step, body]
func (n *ForNode) Generate(st *SymbolTable) {
	if _, ok := st.table[n.varName]; !ok {
		st.nextShift += 4
		v := mkNumber(0)
		v.shift = st.nextShift
		st.table[n.varName] = v
		codeAppend(fmt.Sprintf("  sub esp, 4 ; for %s [EBP-%d]", n.varName, st.nextShift))
	}
	varShift := st.table[n.varName].shift

	n.children[0].Generate(st) // start → EAX
	codeAppend(fmt.Sprintf("  mov [ebp-%d], eax", varShift))

	id := newNodeID()
	body := n.children[2]
	var stepNode Node
	if len(n.children) == 4 {
		stepNode = n.children[2]
		body = n.children[3]
	}

	codeAppend(fmt.Sprintf("loop_%d:", id))
	n.children[1].Generate(st) // limit → EAX
	codeAppend("  push eax")
	codeAppend(fmt.Sprintf("  mov eax, [ebp-%d]", varShift))
	codeAppend("  pop ecx")
	codeAppend("  cmp eax, ecx")
	codeAppend("  mov eax, 0")
	codeAppend("  mov ecx, 1")
	codeAppend("  cmovle eax, ecx")
	codeAppend("  cmp eax, 0")
	codeAppend(fmt.Sprintf("  je exit_%d", id))

	body.Generate(st)

	if stepNode != nil {
		stepNode.Generate(st) // step → EAX
		codeAppend("  push eax")
		codeAppend(fmt.Sprintf("  mov eax, [ebp-%d]", varShift))
		codeAppend("  pop ecx")
		codeAppend("  add eax, ecx")
	} else {
		codeAppend(fmt.Sprintf("  mov eax, [ebp-%d]", varShift))
		codeAppend("  add eax, 1")
	}
	codeAppend(fmt.Sprintf("  mov [ebp-%d], eax", varShift))
	codeAppend(fmt.Sprintf("  jmp loop_%d", id))
	codeAppend(fmt.Sprintf("exit_%d:", id))
}

// RepeatNode — repeat body until condition is true
func (n *RepeatNode) Generate(st *SymbolTable) {
	id := newNodeID()
	codeAppend(fmt.Sprintf("loop_%d:", id))
	n.children[0].Generate(st) // body
	n.children[1].Generate(st) // condition → EAX
	codeAppend("  cmp eax, 0")
	codeAppend(fmt.Sprintf("  je loop_%d", id))
}

// ReadVal
func (n *ReadVal) Generate(_ *SymbolTable) {
	codeAppend("  push scan_int")
	codeAppend("  push format_in")
	codeAppend("  call scanf")
	codeAppend("  add esp, 8")
	codeAppend("  mov eax, [scan_int]")
}

// FuncDef — emits subroutine into funcInstructions buffer
func (n *FuncDef) Generate(_ *SymbolTable) {
	globalFuncTable[n.name] = n

	old := currentBuffer
	currentBuffer = &funcInstructions

	// Build a fresh symbol table with params at [EBP+8], [EBP+12], ...
	funcST := NewSymbolTable()
	for i, param := range n.params {
		v := mkNumber(0)
		v.shift = 8 + i*4
		v.isParam = true
		funcST.table[param] = v
	}

	codeAppend(fmt.Sprintf("func_%s:", n.name))
	codeAppend("  push ebp")
	codeAppend("  mov ebp, esp")
	n.body.Generate(funcST)
	// default fall-through return
	codeAppend("  mov esp, ebp")
	codeAppend("  pop ebp")
	codeAppend("  ret")

	currentBuffer = old
}

// FuncCall — pushes args right-to-left, calls subroutine; result in EAX
func (n *FuncCall) Generate(st *SymbolTable) {
	for i := len(n.children) - 1; i >= 0; i-- {
		n.children[i].Generate(st)
		codeAppend("  push eax")
	}
	codeAppend(fmt.Sprintf("  call func_%s", n.name))
	if len(n.children) > 0 {
		codeAppend(fmt.Sprintf("  add esp, %d", len(n.children)*4))
	}
}

// ReturnNode — stores value in EAX and tears down the frame
func (n *ReturnNode) Generate(st *SymbolTable) {
	n.children[0].Generate(st)
	codeAppend("  mov esp, ebp")
	codeAppend("  pop ebp")
	codeAppend("  ret")
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
	ast := run(source)
	ast.Generate(NewSymbolTable())
	asmFile := strings.TrimSuffix(os.Args[1], ".lua") + ".asm"
	codeDump(asmFile)
}
