# compilador-logcomp

[![Compilation Status](https://compiler-tester.insper-comp.com.br/svg/EnzoSilva14/compilador-logcomp)](https://compiler-tester.insper-comp.com.br/svg/EnzoSilva14/compilador-logcomp)

This repository is monitored by Compiler Tester for automatic compilation status.

## Grammar (EBNF v2.1)

```ebnf
PROGRAM        = { STATEMENT } ;
STATEMENT      = ( (IDENTIFIER, "=", BOOLEXPRESSION) | (IF, "(", BOOLEXPRESSION, ")", STATEMENT, ("ELSE", STATEMENT) | Ε) | (PRINT, "(", BOOLEXPRESSION, ")") | (WHILE, "(", BOOLEXPRESSION, ")", STATEMENT) | Ε ), EOL ;
BOOLEXPRESSION = BOOLTERM, { "||", BOOLTERM } ;
BOOLTERM       = RELEXPRESSION, { "&&", RELEXPRESSION } ;
RELEXPRESSION  = EXPRESSION, ("<" | "==" | ">"), EXPRESSION ;
EXPRESSION     = TERM, { ("+" | "-"), TERM } ;
TERM           = FACTOR, { ("*" | "/"), FACTOR } ;
FACTOR         = ("+"|"-"), FACTOR | "(", BOOLEXPRESSION, ")" | NUMBER | READ, "(", ")" ;
NUMBER         = DIGIT, { DIGIT } ;
IDENTIFIER     = LETTER, { LETTER | DIGIT | "_" } ;
DIGIT          = 0 | 1 | ... | 9 ;
LETTER         = a | b | ... | z | A | B | ... | Z ;
```
