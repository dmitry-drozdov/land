package main

import (
	"os"
	"testing"

	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/parser"
)

var (
	sources *ast.Source
)

func init() {
	// Чтение файла — ВНЕ таймера benchmark'а (init выполняется до запуска бенча)
	b, err := os.ReadFile("e:\\phd\\ts\\test\\3\\schema.graphql")
	if err != nil {
		panic(err)
	}
	sources = &ast.Source{
		Name:  "schema.graphql",
		Input: string(b),
	}

}

func BenchmarkParseSchema(b *testing.B) {
	// Прогрев (не обязателен, но можно)
	if _, err := parser.ParseSchema(sources); err != nil {
		b.Fatalf("parse error (warmup): %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer() // Важно: всё, что выше — вне замера

	for i := 0; i < b.N; i++ {
		if _, err := parser.ParseSchema(sources); err != nil {
			b.Fatalf("parse error: %v", err)
		}
	}
}
