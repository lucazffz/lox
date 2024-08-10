package main

import (
	"bufio"
	"fmt"
	"github.com/LucazFFz/lox/internal/scan"
	"os"
)

func main() {
	switch len(os.Args) {
	case 1:
		execPrompt()
	case 2:
		execFile(os.Args[1])
	default:
		fmt.Println("Usage: jlox [script])")
		os.Exit(64)
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

func execFile(path string) {
	text, err := os.ReadFile(path)
	if err == nil {
		exec(string(text))
	} else {
		fmt.Println("Could not read file: ", err)
	}
}

func exec(source string) {
	tokens, err := scan.Scan(source, scan.ScanContext{})

	if err != nil {
		fmt.Println(err)
	}

	for _, token := range tokens {
		fmt.Println(token)
	}
}

func err(line int, err string) {

}
