# compilador-logcomp

[![Compilation Status](https://compiler-tester.insper-comp.com.br/svg/EnzoSilva14/compilador-logcomp)](https://compiler-tester.insper-comp.com.br/svg/EnzoSilva14/compilador-logcomp)

This repository is monitored by Compiler Tester for automatic compilation status.

## Syntax Diagram (v1.1)

```
expression:

            ┌────────────────────────────────┐
            │                                │
            ▼                                │
──► [TERM] ─┴─┬─ ['+'] ─┬─ [TERM] ──────────┘──► EOF
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
    └─ [NUMBER] ──────────────────────────────────►
```

## Grammar (EBNF)

```ebnf
EXPRESSION = TERM, { ("+" | "-"), TERM } ;
TERM       = FACTOR, { ("*" | "/"), FACTOR } ;
FACTOR     = ("+" | "-"), FACTOR | "(", EXPRESSION, ")" | NUMBER ;
NUMBER     = DIGIT, { DIGIT } ;
DIGIT      = 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 ;
```
