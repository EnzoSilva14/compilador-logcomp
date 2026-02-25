package main

import (
	"fmt"
	"os"
	"strconv"
	"unicode"
)

func main() {
	if len(os.Args) < 2 {
		panic("Nenhum argumento fornecido. Uso: go run main.go 'expressao'")
	}

	input := os.Args[1]
	result := evaluate(input)	
	fmt.Println(result)
}

func evaluate(input string) int {
	var tokens []string
	i := 0

	// --- LEXER ---
	for i < len(input) {
		char := rune(input[i])

		if unicode.IsSpace(char) {
			i++
			continue
		}

		if unicode.IsDigit(char) {
			start := i
			for i < len(input) && unicode.IsDigit(rune(input[i])) {
				i++
			}
			tokens = append(tokens, input[start:i])
		} else if char == '+' || char == '-' {
			tokens = append(tokens, string(char))
			i++
		} else {
			panic(fmt.Sprintf("Caractere inválido: %c", char))
		}
	}

	if len(tokens) == 0 {
		panic("Expressão vazia")
	}

	// --- PARSER / EVALUATOR ---
	val, err := strconv.Atoi(tokens[0])
	if err != nil {
		panic("A expressão deve começar com um número")
	}
	
	result := val
	idx := 1

	for idx < len(tokens) {
		op := tokens[idx]
		idx++

		if op != "+" && op != "-" {
			panic("Esperado operador (+ ou -) entre números")
		}

		if idx >= len(tokens) {
			panic("Expressão incompleta: operador sem operando")
		}

		nextValStr := tokens[idx]
		idx++

		nextVal, err := strconv.Atoi(nextValStr)
		if err != nil {
			panic("Esperado número após operador")
		}

		if op == "+" {
			result += nextVal
		} else {
			result -= nextVal
		}
	}

	return result
}