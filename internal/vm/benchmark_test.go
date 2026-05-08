package vm

import (
	"fmt"
	"slug/internal/ast"
	"slug/internal/lexer"
	"slug/internal/object"
	"slug/internal/parser"
	"slug/internal/semantic"
	"testing"
)

var vmBenchmarkPrograms = map[string]string{
	"arith_recur": `
val fib = fn(n, a = 0, b = 1) {
  if (n == 0) { a } else { recur(n - 1, b, a + b) }
}
fib(300)
`,
	"match_nested": `
val classify = fn(v) match {
  [[], ...] => 0
  [[x, ...tail], ...] if x > 1000 => 3
  [[x, ...tail], ...] => x + 1
  _ => -1
}
classify([[12, 13, 14], 99])
`,
	"struct_copy": `
val User = struct {
  name,
  @num age,
}
val mk = fn(n, age = 1, acc = User { name: "u0", age: 0 }) {
  if (n == 0) {
    acc
  } else {
    recur(n - 1, age + 1, acc copy { name: "u" + n, age: age })
  }
}
mk(400)
`,
}

func parseProgramForBench(b *testing.B, name, src string) *ast.Program {
	b.Helper()
	l := lexer.New(src)
	p := parser.New(l, fmt.Sprintf("bench_%s.slug", name), src)
	program := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		b.Fatalf("%s parse errors: %v", name, errs)
	}
	if errs := semantic.Analyze(fmt.Sprintf("bench_%s.slug", name), src, program); len(errs) > 0 {
		b.Fatalf("%s semantic errors: %v", name, errs)
	}
	return program
}

func compileProgramForBench(b *testing.B, name, src string) *Chunk {
	b.Helper()
	program := parseProgramForBench(b, name, src)
	chunk, err := Compile(program)
	if err != nil {
		b.Fatalf("%s compile error: %v", name, err)
	}
	return chunk
}

func BenchmarkVMParseCompile(b *testing.B) {
	for name, src := range vmBenchmarkPrograms {
		name, src := name, src
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				program := parseProgramForBench(b, name, src)
				if _, err := Compile(program); err != nil {
					b.Fatalf("%s compile error: %v", name, err)
				}
			}
		})
	}
}

func BenchmarkVMCompileOnly(b *testing.B) {
	for name, src := range vmBenchmarkPrograms {
		name, src := name, src
		program := parseProgramForBench(b, name, src)
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if _, err := Compile(program); err != nil {
					b.Fatalf("%s compile error: %v", name, err)
				}
			}
		})
	}
}

func BenchmarkVMExecuteOnly(b *testing.B) {
	for name, src := range vmBenchmarkPrograms {
		name, src := name, src
		chunk := compileProgramForBench(b, name, src)
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				env := object.NewRootEnvironment(4)
				exec := NewExecutor(env, nil)
				out := exec.run(chunk)
				if out == nil || out.Type() == object.ERROR_OBJ {
					b.Fatalf("%s execute failed: %v", name, out)
				}
			}
		})
	}
}
