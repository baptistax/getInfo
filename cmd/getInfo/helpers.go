package main

import (
	"os"
	"strconv"
)

// Small wrappers used across files (avoid importing strconv/os everywhere).

// strconvFormatInt converts an int64 to its base-10 string representation.
func strconvFormatInt(i int64) string {
	return strconv.FormatInt(i, 10)
}

// osGetenv is kept for compatibility with older code that may still call it.
func osGetenv(k string) string {
	return os.Getenv(k)
}
