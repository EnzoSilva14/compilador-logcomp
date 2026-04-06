# compilador-logcomp

[![Compilation Status](https://compiler-tester.insper-comp.com.br/svg/EnzoSilva14/compilador-logcomp)](https://compiler-tester.insper-comp.com.br/svg/EnzoSilva14/compilador-logcomp)

This repository is monitored by Compiler Tester for automatic compilation status.

## Syntax Diagram (v2.0)

```
program:

──► { STATEMENT } ──► EOF

statement:

──►─┬─ [IDENTIFIER] ── '=' ── [EXPRESSION] ── '\n' ──►
    ├─ 'print' ── '(' ── [EXPRESSION] ── ')' ── '\n' ──►
    └─ '\n' ──────────────────────────────────────────►

expression:

            ┌────────────────────────────────┐
            │                                │
            ▼                                │
──► [TERM] ─┴─┬─ ['+'] ─┬─ [TERM] ──────────┘
              └─ ['-'] ─┘

term:

              ┌────────────────────────────────┐
              │                                │
              ▼                                │
──► [FACTOR] ─┴─┬─ ['*'] ─┬─ [FACTOR] ────────┘
                └─ ['/'] ─┘

factor:

──►─┬─ ['+'] ─┬─ [FACTOR] ──────────────────────►
    ├─ ['-'] ─┘
    ├─ '(' ── [EXPRESSION] ── ')' ───────────────►
    ├─ [NUMBER] ────────────────────────────────►
    └─ [IDENTIFIER] ────────────────────────────►
```

## Grammar (EBNF)

```ebnf
PROGRAM    = { STATEMENT } ;
STATEMENT  = ( (IDENTIFIER, "=", EXPRESSION) | (PRINT, "(", EXPRESSION, ")") | Ε ), EOL ;
EXPRESSION = TERM, { ("+" | "-"), TERM } ;
TERM       = FACTOR, { ("*" | "/"), FACTOR } ;
FACTOR     = ("+" | "-"), FACTOR | "(", EXPRESSION, ")" | NUMBER ;
NUMBER     = DIGIT, { DIGIT } ;
IDENTIFIER = LETTER, { LETTER | DIGIT | "_" } ;
DIGIT      = 0 | 1 | ... | 9 ;
LETTER     = a | b | ... | z | A | B | ... | Z ;
```
