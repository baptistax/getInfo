package main

import "strings"

// getGPUInfo returns:
//   1) "Yes"/"No" indicating if a GPU/video controller was detected
//   2) GPU model name (first relevant entry)
//   3) Approximate dedicated VRAM in GiB (as string), or "" if unknown.
func getGPUInfo(c *collector) (string, string, string) {
	// Preferred: CIM / Win32_VideoController via PowerShell
	ps := `
$g = Get-CimInstance Win32_VideoController -ErrorAction SilentlyContinue |
  Where-Object { $_.Name -and $_.AdapterRAM } |
  Select-Object -First 1 Name, AdapterRAM
if ($g) {
  "$($g.Name)` + "`n" + `$($g.AdapterRAM)"
}
`
	if out, err := runPS(ps); err == nil && strings.TrimSpace(out) != "" {
		lines := strings.Split(strings.TrimSpace(out), "\n")
		if len(lines) >= 1 {
			model := strings.TrimSpace(lines[0])
			var vram string
			if len(lines) >= 2 {
				if v, err := parseInt64Any(lines[1]); err == nil && v > 0 {
					vram = strconvFormatInt(toGiB(v))
				}
			}
			if model != "" {
				return "Yes", model, vram
			}
		}
	}

	// Fallback: WMIC Win32_VideoController
	if out, err := runCMD(`wmic path win32_VideoController get Name,AdapterRAM /value`); err == nil && strings.TrimSpace(out) != "" {
		var model string
		var vramB int64

		for _, ln := range strings.Split(out, "\n") {
			line := strings.TrimSpace(ln)
			if strings.HasPrefix(strings.ToLower(line), "name=") {
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					model = strings.TrimSpace(parts[1])
				}
			} else if strings.HasPrefix(strings.ToLower(line), "adapterram=") {
				if v, err := parseInt64Any(line); err == nil && v > 0 {
					vramB = v
				}
			}
		}

		if model != "" {
			var vram string
			if vramB > 0 {
				vram = strconvFormatInt(toGiB(vramB))
			}
			return "Yes", model, vram
		}
	}

	// No GPU info found
	c.addErr("gpu", ErrNotFound, "")
	return "No", "", ""
}
