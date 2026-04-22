# compilador-logcomp

[![Compilation Status](https://compiler-tester.insper-comp.com.br/svg/EnzoSilva14/compilador-logcomp)](https://compiler-tester.insper-comp.com.br/svg/EnzoSilva14/compilador-logcomp)

This repository is monitored by Compiler Tester for automatic compilation status.

## Grammar (EBNF v2.2)

```ebnf
PROGRAM        = { STATEMENT } ;
STATEMENT      = ( VARDEC | (IDENTIFIER, "=", BOOLEXPRESSION) | (IF, BOOLEXPRESSION, "then", "\n", BLOCK, ["else", "\n", BLOCK], "end") | (PRINT, "(", BOOLEXPRESSION, ")") | (WHILE, BOOLEXPRESSION, "do", "\n", BLOCK, "end") | (FOR, IDENTIFIER, "=", EXPRESSION, ",", EXPRESSION, [",", EXPRESSION], "do", "\n", BLOCK, "end") | (REPEAT, "\n", BLOCK, "until", BOOLEXPRESSION) | ("do", "\n", BLOCK, "end") | Ε ), EOL ;
VARDEC         = "local", IDENTIFIER, TYPE, ["=", BOOLEXPRESSION] ;
BOOLEXPRESSION = BOOLTERM, { "or", BOOLTERM } ;
BOOLTERM       = RELEXPRESSION, { "and", RELEXPRESSION } ;
RELEXPRESSION  = EXPRESSION, [("<" | "==" | ">"), EXPRESSION] ;
EXPRESSION     = TERM, { ("+" | "-"), TERM } ;
TERM           = FACTOR, { ("*" | "/"), FACTOR } ;
FACTOR         = ("+"|"-"), FACTOR | READ, "(", ")" | POWER ;
POWER          = ATOM, ["**", FACTOR] ;
ATOM           = "(", BOOLEXPRESSION, ")" | INT | IDENTIFIER | BOOL | STR | (IF, BOOLEXPRESSION, "then", BOOLEXPRESSION, "else", BOOLEXPRESSION, "end") ;
TYPE           = "number" | "string" | "boolean" ;
BOOL           = "true" | "false" ;
INT            = DIGIT, { DIGIT } ;
STR            = '"', { CHAR }, '"' ;
IDENTIFIER     = LETTER, { LETTER | DIGIT | "_" } ;
DIGIT          = 0 | 1 | ... | 9 ;
LETTER         = a | b | ... | z | A | B | ... | Z ;
```
