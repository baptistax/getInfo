package main

import "strings"

// MachineGuid is unique per Windows installation (can change after sysprep).
func getMachineGuid(c *collector) string {
	// Fast path: reg.exe query
	out, err := runCMD(`reg query "HKLM\SOFTWARE\Microsoft\Cryptography" /v MachineGuid`)
	if err == nil && strings.TrimSpace(out) != "" {
		for _, ln := range strings.Split(out, "\n") {
			if strings.Contains(ln, "MachineGuid") {
				parts := strings.Fields(ln)
				if len(parts) >= 3 {
					return strings.TrimSpace(parts[len(parts)-1])
				}
			}
		}
	}
	// Fallback: PowerShell
	ps := `(Get-ItemProperty 'HKLM:\SOFTWARE\Microsoft\Cryptography').MachineGuid`
	if o2, e2 := runPS(ps); e2 == nil && strings.TrimSpace(o2) != "" {
		return firstLine(o2)
	}
	c.addErr("machineguid", err, "")
	return ""
}
