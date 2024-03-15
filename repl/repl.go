package repl

import (
	"bufio"
	"fmt"
	"io"

	"interpego/compiler"
	"interpego/lexer"
	"interpego/object"
	"interpego/parser"
	"interpego/vm"
)

const PROMPT = ">> "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	symbols := compiler.NewSymbolTable()
	globals := make([]object.Object, vm.GLOBALS_SIZE)
	// env := object.NewEnvironment()
	// builtins := evaluator.NewBuiltins()
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

		compiler := compiler.NewWithSymbols(symbols)
		err := compiler.Compile(program)
		if err != nil {
			fmt.Fprintf(out, "Woops! Compilation failed:\n %s\n", err)
			continue
		}

		vm := vm.NewWithGlobals(globals, compiler.Bytecode())
		err = vm.Run()
		if err != nil {
			fmt.Fprintf(out, "Woops! Executing bytecode failed:\n %s\n", err)
			continue
		}

		io.WriteString(out, "=> "+vm.LastPoppedStackElement().Inspect())
		io.WriteString(out, "\n\n")
	}
}

func printParserErrors(out io.Writer, errors []string) {
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg+"\n")
	}
}
