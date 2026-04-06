# Documentação do Compilador — compilador-logcomp

Compilador incremental desenvolvido para a disciplina **LOGCOMP** (Insper), escrito em Go.
Versão atual: **v2.0**

---

## Visão Geral do Pipeline

```
Arquivo .lua
     │
     ▼
┌──────────┐
│  PrePro  │  Remove comentários (--)
└──────────┘
     │  string filtrada
     ▼
┌──────────┐
│  Lexer   │  Transforma texto em tokens
└──────────┘
     │  stream de tokens
     ▼
┌──────────┐
│  Parser  │  Constrói a AST a partir dos tokens
└──────────┘
     │  árvore de nós (Node)
     ▼
┌──────────────┐
│  SymbolTable │  Tabela de variáveis usada durante avaliação
└──────────────┘
     │
     ▼
┌──────────────────┐
│  AST.Evaluate()  │  Percorre a árvore e executa o programa
└──────────────────┘
     │
     ▼
  Saída (stdout)
```

---

## 1. PrePro (Pré-processador)

**Struct:** `PrePro`
**Método:** `Filter(code string) string`

### Função
Executa antes do Lexer. Recebe o código-fonte completo como string e remove todos os comentários inline antes que qualquer análise léxica aconteça.

### Como funciona
- Divide o código em linhas pelo caractere `\n`
- Em cada linha, procura pela sequência `--`
- Tudo a partir de `--` até o fim da linha é descartado
- As linhas são reunidas de volta com `\n`, preservando a estrutura original

### Exemplo
```
-- Comentário inicial
x = 3 + 5   -- soma dois valores
y = x * 2
```
Após `Filter`:
```
\n
x = 3 + 5   \n
y = x * 2\n
```

### Por que é uma etapa separada
Comentários não têm significado semântico. Removê-los antes da tokenização simplifica o Lexer, que não precisa conhecer a sintaxe de comentários.

---

## 2. Lexer (Analisador Léxico)

**Struct:** `Lexer`
**Método principal:** `selectNext()`
**Campo público:** `Next Token`

### Função
Transforma a string do código-fonte em uma sequência de **tokens** consumidos um a um pelo Parser. Opera de forma *lazy* (sob demanda): cada chamada a `selectNext()` avança para o próximo token e o armazena em `Next`.

### Tokens reconhecidos

| Constante   | Símbolo       | Descrição                          |
|-------------|---------------|------------------------------------|
| `INT`       | `0`..`9`+     | Literal inteiro                    |
| `IDEN`      | `[a-zA-Z][a-zA-Z0-9_]*` | Identificador de variável |
| `PRINT`     | `print`       | Palavra-chave de impressão         |
| `PLUS`      | `+`           | Adição / unário positivo           |
| `MINUS`     | `-`           | Subtração / unário negativo        |
| `MULT`      | `*`           | Multiplicação                      |
| `DIV`       | `/`           | Divisão                            |
| `POW`       | `**`          | Potenciação (extra credit v1.1)    |
| `OPEN_PAR`  | `(`           | Abre parêntese                     |
| `CLOSE_PAR` | `)`           | Fecha parêntese                    |
| `ASSIGN`    | `=`           | Atribuição                         |
| `END`       | `\n`          | Fim de instrução (newline)         |
| `EOF`       | —             | Fim do arquivo                     |

### Regras importantes
- **Espaços e tabs** são ignorados; **`\n` não é ignorado** — ele vira o token `END` que delimita instruções.
- **Identificadores** começam obrigatoriamente com letra; underscores e dígitos são permitidos nas posições seguintes. Exemplos válidos: `x`, `x1`, `z_final`. Inválidos: `1x`, `_x`.
- **Palavra-chave `print`** é reconhecida pelo Lexer e retorna token `PRINT`, não `IDEN`.
- **`**`** (potência): o Lexer faz lookahead de 1 caractere para distinguir `*` de `**`.
- Qualquer caractere não reconhecido causa `panic("[Lexer] Invalid Symbol ...")`.

---

## 3. Parser (Analisador Sintático)

**Funções:** `parseProgram`, `parseStatement`, `parseExpression`, `parseTerm`, `parseFactor`, `parsePower`, `parseAtom`

### Função
Consome os tokens produzidos pelo Lexer e constrói a **Árvore Sintática Abstrata (AST)**. Cada função corresponde a uma regra da gramática e retorna um nó `Node`.

O Parser usa a técnica de **descida recursiva**: cada função chama as funções de menor precedência de acordo com a gramática EBNF.

### Gramática (EBNF — v2.0)

```ebnf
PROGRAM    = { STATEMENT } ;
STATEMENT  = ( IDENTIFIER, "=", EXPRESSION
             | "print", "(", EXPRESSION, ")"
             | ε ), "\n" ;
EXPRESSION = TERM, { ("+" | "-"), TERM } ;
TERM       = FACTOR, { ("*" | "/"), FACTOR } ;
FACTOR     = ("+" | "-"), FACTOR
           | "(", EXPRESSION, ")"
           | NUMBER
           | IDENTIFIER ;
NUMBER     = DIGIT, { DIGIT } ;
IDENTIFIER = LETTER, { LETTER | DIGIT | "_" } ;
DIGIT      = "0" | "1" | ... | "9" ;
```

### Hierarquia de precedência (maior → menor)

| Nível | Regra             | Operadores              |
|-------|-------------------|-------------------------|
| 1     | `parseAtom`       | `(...)`, literal `NUMBER` |
| 2     | `parsePower`      | `**` (direita-associativo) |
| 3     | `parseFactor`     | unário `+`, unário `-`, `IDENTIFIER` |
| 4     | `parseTerm`       | `*`, `/`                |
| 5     | `parseExpression` | `+`, `-`                |
| 6     | `parseStatement`  | `=`, `print`            |
| 7     | `parseProgram`    | bloco de instruções     |

### Descrição de cada função

#### `parseProgram`
Ponto de entrada da análise. Chama `parseStatement` em loop até EOF e agrupa todos os nós em um único `Block`.

#### `parseStatement`
Identifica o tipo de instrução pelo token atual:
- `IDEN` → atribuição: consome `IDEN = EXPRESSION \n`, retorna `Assignment`
- `PRINT` → impressão: consome `print(EXPRESSION) \n`, retorna `Print`
- `END` → linha vazia: consome `\n`, retorna `NoOp`

#### `parseExpression`
Lida com `+` e `-` entre termos (operadores de menor precedência após `parseStatement`). Produz nós `BinOp("+")` ou `BinOp("-")` encadeados à esquerda.

#### `parseTerm`
Lida com `*` e `/` entre fatores. Mesma lógica de `parseExpression`, encadeamento à esquerda.

#### `parseFactor`
Lida com operadores unários (`+`, `-`) de forma **recursiva à direita**, e com identificadores de variável. Se nenhum desses, delega para `parsePower`.

#### `parsePower`
Lida com `**` de forma **recursiva à direita** (direita-associativo): `2 ** 3 ** 2` = `2 ** (3 ** 2)`. Delega o operando esquerdo para `parseAtom`.

#### `parseAtom`
Regra base: lida com `(EXPRESSION)` (grupos entre parênteses) e literais `NUMBER`. Os parênteses não geram nó próprio na AST — apenas influenciam a ordem de construção da árvore.

---

## 4. AST — Árvore Sintática Abstrata

### Interface base: `Node`

```go
type Node interface {
    Evaluate(st *SymbolTable) int
}
```

Todos os nós implementam `Evaluate(st *SymbolTable) int`. O método recebe a tabela de símbolos e retorna o valor inteiro resultante (nós sem valor de retorno retornam `0`).

### Regra de ouro
> Cada nó cuida apenas de si mesmo. Chama `Evaluate()` nos filhos e aplica sua operação. Nunca olha para nós ancestrais ou além dos seus filhos diretos.

### Nós disponíveis

#### `IntVal` — Literal inteiro
- **Filhos:** nenhum (nó folha)
- **`value`:** o inteiro constante
- **`Evaluate()`:** retorna `value`
- **Exemplo:** `42` → `IntVal{value: 42}`

#### `Identifier` — Referência a variável
- **Filhos:** nenhum (nó folha)
- **`value`:** nome da variável (string)
- **`Evaluate()`:** consulta `st.Get(name)` e retorna o valor atual da variável; erro `[Semantic]` se não definida
- **Exemplo:** `x` → `Identifier{value: "x"}`

#### `UnOp` — Operação unária
- **Filhos:** 1 (o operando)
- **`value`:** operador (`"+"` ou `"-"`)
- **`Evaluate()`:** avalia o filho e aplica o operador
- **Exemplo:** `-3` → `UnOp{"-", [IntVal{3}]}`

#### `BinOp` — Operação binária
- **Filhos:** 2 (`children[0]` = esquerdo, `children[1]` = direito)
- **`value`:** operador (`"+"`, `"-"`, `"*"`, `"/"`, `"**"`)
- **`Evaluate()`:** avalia ambos os filhos e aplica o operador; lança `[Semantic] Division by zero` se divisor = 0
- **Exemplo:** `3 + 5` → `BinOp{"+", [IntVal{3}, IntVal{5}]}`

#### `Assignment` — Atribuição de variável
- **Filhos:** 2 (`children[0]` = `Identifier` com o nome, `children[1]` = expressão)
- **`Evaluate()`:** avalia `children[1]` e armazena o resultado em `st.Set(name, val)`; retorna 0
- **Exemplo:** `x = 3 + 5` → `Assignment{[Identifier{"x"}, BinOp{"+", ...}]}`

#### `Print` — Impressão
- **Filhos:** 1 (a expressão a imprimir)
- **`Evaluate()`:** avalia o filho e imprime o resultado via `fmt.Println`; retorna 0
- **Exemplo:** `print(x)` → `Print{[Identifier{"x"}]}`

#### `Block` — Bloco de instruções
- **Filhos:** N (uma instrução por filho)
- **`Evaluate()`:** chama `Evaluate()` em cada filho em ordem; retorna 0
- **Papel:** raiz da AST; agrupa todo o programa

#### `NoOp` — Instrução vazia
- **Filhos:** nenhum
- **`Evaluate()`:** não faz nada, retorna 0
- **Uso:** representa linhas em branco ou comentários que viraram linhas vazias após o PrePro

### Exemplo de AST

Código:
```lua
x = 3 + 5
print(x)
```

Árvore resultante:
```
Block
├── Assignment
│   ├── Identifier("x")
│   └── BinOp("+")
│       ├── IntVal(3)
│       └── IntVal(5)
└── Print
    └── Identifier("x")
```

---

## 5. SymbolTable (Tabela de Símbolos)

**Struct:** `SymbolTable`
**Métodos:** `Get(name)`, `Set(name, val)`

### Função
Armazena e recupera o valor atual de cada variável durante a execução do programa. É criada uma única vez em `main()` e passada para todos os nós via `Evaluate(st)`.

### Comportamento
- `Set(name, val)` — cria ou sobrescreve a variável `name` com o valor `val`
- `Get(name)` — retorna o valor de `name`; se a variável não existir, lança `panic("[Semantic] Undefined variable: name")`

---

## 6. Fluxo de Execução — `main()`

```
1. Lê o arquivo .lua fornecido como argumento
2. Concatena "\n" ao final (garante que a última linha seja terminada)
3. Passa pelo PrePro.Filter() para remover comentários
4. Chama run(source):
   a. Cria um Lexer com o código filtrado
   b. Chama parseProgram() → retorna o nó raiz (Block)
5. Cria uma SymbolTable vazia
6. Chama root.Evaluate(st) → executa o programa
```

---

## 7. Tratamento de Erros

Todos os erros são lançados via `panic` com prefixos padronizados:

| Prefixo      | Origem           | Exemplos                                         |
|--------------|------------------|--------------------------------------------------|
| `[Lexer]`    | Lexer            | Caractere inválido no código-fonte               |
| `[Parser]`   | Parser           | Token inesperado, parêntese não fechado          |
| `[Semantic]` | Nós da AST / SymbolTable | Variável não definida, divisão por zero |
| `[Main]`     | main()           | Argumento ausente, arquivo não encontrado        |

---

## 8. Histórico de Versões

| Versão | Roteiro | Funcionalidades adicionadas                                      |
|--------|---------|------------------------------------------------------------------|
| v0.0   | 1       | `+`, `-` com inteiros; pipeline Lexer → Parser direto           |
| v1.0   | 2       | `*`, `/`, operadores unários `+`/`-`, parênteses               |
| v1.1   | 3 (extra) | Potenciação `**` (direita-associativa); regras `POWER`, `ATOM` |
| v1.2   | 4       | AST com nós `IntVal`, `UnOp`, `BinOp`; avaliação separada do parsing |
| v1.2   | 4 (extra) | Operador fatorial `!` (postfix); nó `UnOp("!")`               |
| v2.0   | 5       | Variáveis (`IDEN`), atribuição, `print`, comentários (`--`), `SymbolTable`, `PrePro`, `Block`, `NoOp` |
