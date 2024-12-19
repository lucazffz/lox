package main

import (
	"bufio"
	"fmt"
	"github.com/LucazFFz/lox/internal/ast"
	"github.com/LucazFFz/lox/internal/parse"
	"github.com/LucazFFz/lox/internal/scan"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"strings"
)

func main() {
	app := &cli.App{
		Name:        "Lox interpreter",
		Usage:       "",
		Description: "A interpreter for the lox programming language.",
		UsageText:   "lox [script] - Script might be omitted to enter interactive mode.",
		Action: func(cCtx *cli.Context) error {
			if cCtx.Args().Len() == 0 {
				runRepl()
				print("Leaving Lox REPL")
				return cli.Exit("", 0)
			} else if cCtx.Args().Len() == 1 {
				err := runFile(cCtx.Args().First())
				if err != nil {
					return cli.Exit(err.Error(), 64)
				}
			}

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func runRepl() {
	block_mode := false
	reader := bufio.NewReader(os.Stdin)
	var text string
	for {
		if block_mode {
			var block strings.Builder
			for {
				fmt.Print("lox|")
				text, _ = reader.ReadString('\n')
				block.WriteString(text)
				if len(text) < 2 {
					block_mode = false
					text = block.String()
					break
				}
			}
		} else {
			fmt.Print("lox>")
			text, _ = reader.ReadString('\n')
		}

		text = strings.Trim(text, "\n ")

		if text == "" {
			continue
		}
		// if the first character is a colon, it is a command
		if text[0] == ':' {
			// be careful not to index out of range
			if len(text) == 2 {
				// exit command
				if text[1] == 'q' {
					return
				}
			} else if len(text) == 4 {
				if text[1:4] == "blk" {
					block_mode = true
					continue
				}
			}

			println("unrecognized command")
			continue
		}

		if text[len(text)-1] != ';' && text[len(text)-1] != '}' {
			// execute expression
			execExpr(string(text))
			continue
		}

		// execute statement
		exec(string(text))
	}
}

func runFile(path string) error {
	if text, err := os.ReadFile(path); err != nil {
		return err
	} else {
		exec(string(text))
		return nil
	}
}

func execExpr(source string) {
	// allow REPL to parse only expressions and print the evaluated value,
	// done for user convenience
	tokens, _ := scan.Scan(source, report, scan.ScanContext{})
	expr, err := parse.ParseExpression(tokens, report)
	if err != nil {
		return
	}

	val, err := expr.Evaluate()
	if err != nil {
		return
	}

	println(val.DebugPrint())
}

func exec(source string) {
	tokens, _ := scan.Scan(source, report, scan.ScanContext{})
	// for _, token := range tokens {
	// 	fmt.Println(token)
	// }

	stmts, err := parse.Parse(tokens, report)
	for _, stmt := range stmts {
		println(stmt.DebugPrint())

	}
	if err != nil {
		return
	}

	ast.Interpret(stmts, report)
	// for _, token := range tokens {
	// 	fmt.Println(token)
	// }
	//

	// fmt.Println(expr.Print())

	// value, err := expr.Evaluate()
	// if err != nil {
	// 	fmt.Println(err)
	// } else {
	// 	fmt.Println(value.Print())
	// }
}

func report(err error) {
	switch e := err.(type) {
	default:
		fmt.Print(e)
	}
}
