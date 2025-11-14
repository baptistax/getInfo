package main

import "strings"

// getAntivirusInfo returns two values:
//   1) a semicolon-separated list of antivirus products detected by Windows Security Center
//   2) the Bitdefender product/edition string, if Bitdefender is present (otherwise empty).
func getAntivirusInfo(c *collector) (string, string) {
	script := `
$av = Get-CimInstance -Namespace root/SecurityCenter2 -ClassName AntiVirusProduct -ErrorAction SilentlyContinue
if (-not $av) {
  $av = Get-CimInstance -Namespace root/SecurityCenter -ClassName AntiVirusProduct -ErrorAction SilentlyContinue
}
$av | Select-Object -ExpandProperty displayName
`
	out, err := runPS(script)
	if err != nil || strings.TrimSpace(out) == "" {
		c.addErr("antivirus", ErrNotFound, "no AntiVirusProduct entries")
		return "", ""
	}

	lines := strings.Split(strings.TrimSpace(out), "\n")
	var products []string
	bitdefender := ""

	for _, ln := range lines {
		name := strings.TrimSpace(ln)
		if name == "" {
			continue
		}
		products = append(products, name)
		if bitdefender == "" && strings.Contains(strings.ToLower(name), "bitdefender") {
			bitdefender = name
		}
	}

	if len(products) == 0 {
		c.addErr("antivirus", ErrNotFound, "parsed zero entries")
		return "", bitdefender
	}

	return strings.Join(products, "; "), bitdefender
}
