package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// Keep usage simple for field techs.
func PrintUsageAndExit() {
	exe := filepath.Base(os.Args[0])
	fmt.Fprintf(os.Stderr, "usage: .\\%s <patrimonio> <nome> <local>\n", exe)
	fmt.Fprintln(os.Stderr, "exemplos:")
	fmt.Fprintf(os.Stderr, "  .\\%s 1029382 laura financeiro\n", exe)
	fmt.Fprintf(os.Stderr, "  .\\%s 1029382 joao \"andar 4\"\n", exe)
	os.Exit(2)
}

// Args wrapper to ease testing/splitting later.
func Args() []string { return os.Args[1:] }
