package main

import (
	"errors"
	"fmt"
	"os"
	"time"
)

// collector is used across files to accumulate non-fatal errors.
type collector struct {
	errs []string
}

// addErr records a contextual error message with a timestamp.
// ctx    -> logical context (e.g. "cpu", "ip_fallback")
// err    -> the error that occurred
// detail -> optional extra detail for debugging
func (c *collector) addErr(ctx string, err error, detail string) {
	if err == nil {
		return
	}

	msg := ctx + ": " + err.Error()
	if detail != "" {
		msg += " | " + detail
	}

	timestamp := time.Now().Format("2006-01-02T15:04:05-0700")
	c.errs = append(c.errs, timestamp+" "+msg)
}

// writeErrors persists collected errors to a log file.
// The actual filename can be chosen by the caller (e.g. local language for operators).
func writeErrors(path string, errs []string) {
	if len(errs) == 0 {
		return
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		// Minimal console output only if the log file cannot be written.
		fmt.Fprintln(os.Stderr, "failed to write error log:", err)
		return
	}
	defer f.Close()

	for _, e := range errs {
		_, _ = f.WriteString(e + "\r\n")
	}
}

// ErrNotFound is a shared sentinel error for "not found" cases.
var ErrNotFound = errors.New("not found")
