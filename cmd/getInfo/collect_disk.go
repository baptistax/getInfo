package main

import (
	"fmt"
	"strings"
)

func getDiskSystemGiB(c *collector) (total string, free string) {
	sysDrive := getenvOr("SystemDrive", "C:")
	// PS first (fast & reliable)
	if out, err := runPS(fmt.Sprintf(`$d = Get-CimInstance Win32_LogicalDisk -Filter "DeviceID='%s'"; "$($d.Size)`+"\n"+`$($d.FreeSpace)"`, sysDrive)); err == nil && out != "" {
		lines := strings.Split(strings.TrimSpace(out), "\n")
		if len(lines) >= 2 {
			if sz, err1 := parseInt64Any(lines[0]); err1 == nil {
				if fr, err2 := parseInt64Any(lines[1]); err2 == nil {
					return strconvFormatInt(toGiB(sz)), strconvFormatInt(toGiB(fr))
				}
			}
		}
	}
	// Fallback: wmic
	out2, err2 := runCMD(fmt.Sprintf(`wmic logicaldisk where "DeviceID='%s'" get Size,FreeSpace /value`, sysDrive))
	if err2 == nil {
		var szB, frB int64
		for _, ln := range strings.Split(out2, "\n") {
			t := strings.ToLower(strings.TrimSpace(ln))
			if strings.HasPrefix(t, "size=") {
				if v, err := parseInt64Any(ln); err == nil {
					szB = v
				}
			} else if strings.HasPrefix(t, "freespace=") {
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

func getIsSSD(c *collector) string {
	// Works on most modern Windows with Storage module available
	if out, err := runPS(`Get-PhysicalDisk | Select-Object -ExpandProperty MediaType`); err == nil && strings.TrimSpace(out) != "" {
		for _, ln := range strings.Split(strings.TrimSpace(out), "\n") {
			if strings.EqualFold(strings.TrimSpace(ln), "SSD") {
				return "Sim"
			}
		}
		return "Nao"
	}
	// Heuristic using model name
	if out2, err2 := runCMD(`wmic diskdrive get Model`); err2 == nil && strings.TrimSpace(out2) != "" {
		for _, ln := range strings.Split(strings.TrimSpace(out2), "\n")[1:] {
			if strings.Contains(strings.ToUpper(ln), "SSD") {
				return "Sim"
			}
		}
		return "Nao"
	}
	return ""
}

// small env helper
func getenvOr(k, def string) string {
	if v := strings.TrimSpace(Getenv(k)); v != "" {
		return v
	}
	return def
}
