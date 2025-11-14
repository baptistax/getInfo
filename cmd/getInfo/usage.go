package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// When 1 or 2 args are passed, show usage (to avoid partial metadata).
func PrintUsageAndExit() {
	exe := filepath.Base(os.Args[0])
	fmt.Fprintf(os.Stderr, "usage: .\\%s <patrimonio> <nome> <local>\n", exe)
	fmt.Fprintln(os.Stderr, "exemplos:")
	fmt.Fprintf(os.Stderr, "  .\\%s 1029382 laura financeiro\n", exe)
	fmt.Fprintf(os.Stderr, "  .\\%s 1029382 joao \"andar 4\"\n", exe)
	os.Exit(2)
}

func Args() []string { return os.Args[1:] }

// Minimal interactive prompt for double-click (no args).
func PromptForInputs() (patr, nome, local string) {
	r := bufio.NewReader(os.Stdin)

	fmt.Print("pat: ")
	patr, _ = r.ReadString('\n')

	fmt.Print("nome: ")
	nome, _ = r.ReadString('\n')

	fmt.Print("local: ")
	local, _ = r.ReadString('\n')

	trim := func(s string) string { return strings.TrimSpace(s) }
	return trim(patr), trim(nome), trim(local)
}
