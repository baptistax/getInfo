package main

import "strings"

// With manifest we should always be elevated; this is just a safety log.
func isAdmin() bool {
	out, err := runPS(`[bool]([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)`)
	if err != nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(out), "True") || strings.EqualFold(strings.TrimSpace(out), "true")
}
