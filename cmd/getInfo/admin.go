package main

import "strings"

// isAdmin checks whether the current process is running with administrative privileges.
// With the manifest it should already be elevated; this is just an additional safeguard.
func isAdmin() bool {
	out, err := runPS(`[bool]([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)`)
	if err != nil {
		return false
	}

	trimmed := strings.TrimSpace(out)
	return strings.EqualFold(trimmed, "true")
}
