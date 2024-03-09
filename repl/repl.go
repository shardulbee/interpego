package repl

import (
	"bufio"
	"fmt"
	"io"

	"interpego/evaluator"
	"interpego/lexer"
	"interpego/object"
	"interpego/parser"
)

const PROMPT = ">> "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	env := object.NewEnvironment()
	builtins := evaluator.NewBuiltins()
	for {
		fmt.Fprintf(out, PROMPT)

		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		lexer := lexer.New(line)
		p := parser.New(lexer)
		program := p.ParseProgram()

		if len(p.Errors()) > 0 {
			printParserErrors(out, p.Errors())
			continue
		} else if len(program.Statements) == 0 {
			continue
		}

		eval := evaluator.Eval(builtins, program, env)

		io.WriteString(out, "=> "+eval.Inspect())
		io.WriteString(out, "\n\n")
	}
}

func printParserErrors(out io.Writer, errors []string) {
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg+"\n")
	}
}
