package main

import (
	"errors"
	"fmt"
	"os"
	"time"
)

// Used across files to collect non-fatal errors.
type collector struct{ errs []string }

func (c *collector) addErr(ctx string, err error, detail string) {
	if err == nil {
		return
	}
	msg := ctx + ": " + err.Error()
	if detail != "" {
		msg += " | " + detail
	}
	c.errs = append(c.errs, time.Now().Format("2006-01-02T15:04:05-0700")+" "+msg)
}

// Single place to persist errors (Portuguese filename kept for operators).
func writeErrors(path string, errs []string) {
	if len(errs) == 0 {
		return
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		// minimal console output only if log can't be written
		fmt.Fprintln(os.Stderr, "erro ao gravar log de erros:", err)
		return
	}
	defer f.Close()
	for _, e := range errs {
		f.WriteString(e + "\r\n")
	}
}

// Shared sentinel error for "not found" cases.
var ErrNotFound = errors.New("nao encontrado")
