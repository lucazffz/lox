package main

import (
	"bufio"
	"fmt"
	"github.com/LucazFFz/lox/internal/scan"
	"github.com/urfave/cli/v2"
	"log"
	"os"
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

	for _, token := range tokens {
		fmt.Println(token)
	}
}

func report(line int, location string, msg string) {
	if location != "" {
		fmt.Printf("[%d] error at \"%s\" - %s \n", line, location, msg)
	} else {
		fmt.Printf("[%d] error - %s \n", line, msg)
	}
}
