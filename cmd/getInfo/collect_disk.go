package main

import (
	"fmt"
	"os"
	"strings"
)

// getDiskSystemGiB returns total and free space (in GiB) for the system drive.
// Values are returned as strings to keep formatting consistent with other fields.
func getDiskSystemGiB(c *collector) (total string, free string) {
	sysDrive := getenvOr("SystemDrive", "C:")

	// Preferred: PowerShell (fast and reliable on modern Windows)
	psScript := fmt.Sprintf(
		`$d = Get-CimInstance Win32_LogicalDisk -Filter "DeviceID='%s'"; "$($d.Size)`+"\n"+`$($d.FreeSpace)"`,
		sysDrive,
	)

	if out, err := runPS(psScript); err == nil && strings.TrimSpace(out) != "" {
		lines := strings.Split(strings.TrimSpace(out), "\n")
		if len(lines) >= 2 {
			if sz, err1 := parseInt64Any(lines[0]); err1 == nil {
				if fr, err2 := parseInt64Any(lines[1]); err2 == nil {
					return strconvFormatInt(toGiB(sz)), strconvFormatInt(toGiB(fr))
				}
			}
		}
	}

	// Fallback: WMIC (older systems)
	out2, err2 := runCMD(fmt.Sprintf(
		`wmic logicaldisk where "DeviceID='%s'" get Size,FreeSpace /value`,
		sysDrive,
	))
	if err2 == nil && strings.TrimSpace(out2) != "" {
		var szB, frB int64
		for _, ln := range strings.Split(out2, "\n") {
			t := strings.ToLower(strings.TrimSpace(ln))
			switch {
			case strings.HasPrefix(t, "size="):
				if v, err := parseInt64Any(ln); err == nil {
					szB = v
				}
			case strings.HasPrefix(t, "freespace="):
				if v, err := parseInt64Any(ln); err == nil {
					frB = v
				}
			}
		}
		if szB > 0 {
			return strconvFormatInt(toGiB(szB)), strconvFormatInt(toGiB(frB))
		}
	}

	c.addErr("disco", ErrNotFound, "")
	return "", ""
}

// getIsSSD tries to determine whether the primary disk is an SSD.
// Returns "Yes", "No" or empty string when unknown.
func getIsSSD(c *collector) string {
	// Most modern Windows systems with the Storage module available
	if out, err := runPS(`Get-PhysicalDisk | Select-Object -ExpandProperty MediaType`); err == nil && strings.TrimSpace(out) != "" {
		for _, ln := range strings.Split(strings.TrimSpace(out), "\n") {
			if strings.EqualFold(strings.TrimSpace(ln), "SSD") {
				return "Yes"
			}
		}
		return "No"
	}

	// Heuristic using disk model name (older systems)
	if out2, err2 := runCMD(`wmic diskdrive get Model`); err2 == nil && strings.TrimSpace(out2) != "" {
		lines := strings.Split(strings.TrimSpace(out2), "\n")
		if len(lines) > 1 {
			for _, ln := range lines[1:] {
				if strings.Contains(strings.ToUpper(ln), "SSD") {
					return "Yes"
				}
			}
			return "No"
		}
	}

	return ""
}

// getenvOr returns the value of the given environment variable (trimmed),
// or the provided default value when it is not set or empty.
func getenvOr(k, def string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return def
}
