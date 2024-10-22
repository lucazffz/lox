package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/LucazFFz/lox/internal/parse"
	"github.com/LucazFFz/lox/internal/scan"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:        "Lox interpreter",
		Usage:       "",
		Description: "A interpreter for the lox programming language.",
		UsageText:   "lox [script] - Script might be omitted to enter interactive mode.",
		Action: func(cCtx *cli.Context) error {
			if cCtx.Args().Len() == 0 {
				execPrompt()
				return nil
			} else if cCtx.Args().Len() == 1 {
				err := execFile(cCtx.Args().First())
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

func execPrompt() {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(">")
		text, _ := reader.ReadString('\n')
		exec(text)
	}
}

func execFile(path string) error {
	if text, err := os.ReadFile(path); err != nil {
		return err
	} else {
		exec(string(text))
		return nil
	}
}

func exec(source string) {
    tokens := scan.Scan(source, report, scan.ScanContext{})
	// for _, token := range tokens {
	// 	fmt.Println(token)
	// }

	stmts, err := parse.Parse(tokens, report)
	if err != nil {
		return
	}
	// for _, token := range tokens {
	// 	fmt.Println(token)
	// }
	//
	for _, stmt := range stmts {
        err := stmt.Evaluate()
        if err != nil {
            fmt.Println(err)
        }
	}


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
