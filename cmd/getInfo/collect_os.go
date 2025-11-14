package main

import (
	"os"
	"regexp"
	"strconv"
	"strings"
)

// getHostname returns the current machine hostname.
func getHostname(c *collector) string {
	h, err := os.Hostname()
	if err != nil || strings.TrimSpace(h) == "" {
		// Fallback to COMPUTERNAME environment variable
		comp := strings.TrimSpace(os.Getenv("COMPUTERNAME"))
		if comp == "" {
			c.addErr("hostname", err, "")
			return ""
		}
		return comp
	}
	return strings.TrimSpace(h)
}

// getCurrentUser returns the current user in a "DOMAIN\\user" or "user" format.
func getCurrentUser(c *collector) string {
	out, err := runCMD("whoami")
	if err == nil && strings.TrimSpace(out) != "" {
		return strings.TrimSpace(out)
	}

	user := strings.TrimSpace(os.Getenv("USERNAME"))
	domain := strings.TrimSpace(os.Getenv("USERDOMAIN"))

	if user == "" {
		if err != nil {
			c.addErr("usuario", err, "USERNAME is empty")
		} else {
			c.addErr("usuario", ErrNotFound, "USERNAME is empty")
		}
		return ""
	}

	// If domain is available and not already present, combine.
	if domain != "" && !strings.Contains(user, `\`) {
		return domain + `\` + user
	}

	return user
}

// getWindowsVersion returns a simplified Windows version string ("11" / "10" / "").
func getWindowsVersion(c *collector) string {
	// 1) WMI caption is robust: "Microsoft Windows 11 Pro"
	if out, err := runPS(`(Get-CimInstance Win32_OperatingSystem).Caption`); err == nil && strings.TrimSpace(out) != "" {
		cap := strings.ToLower(firstLine(out))
		if strings.Contains(cap, "windows 11") {
			return "11"
		}
		if strings.Contains(cap, "windows 10") {
			return "10"
		}
	}

	// 2) Registry ProductName (works on most builds)
	if out, err := runCMD(`reg query "HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion" /v ProductName`); err == nil && strings.TrimSpace(out) != "" {
		for _, ln := range strings.Split(out, "\n") {
			if strings.Contains(ln, "ProductName") {
				p := strings.ToLower(ln)
				if strings.Contains(p, "windows 11") {
					return "11"
				}
				if strings.Contains(p, "windows 10") {
					return "10"
				}
			}
		}
	}

	// 3) "ver" fallback (old but still useful)
	if out, _ := runCMD("ver"); strings.Contains(out, "Version 11.") {
		return "11"
	} else if strings.Contains(out, "Version 10.") {
		return "10"
	}

	return ""
}

// firstLine returns the first non-empty line of a string, trimmed.
func firstLine(s string) string {
	parts := strings.Split(strings.TrimSpace(s), "\n")
	if len(parts) == 0 {
		return ""
	}
	return strings.TrimSpace(parts[0])
}

// parseInt64Any extracts the first integer found in the string and parses it as int64.
func parseInt64Any(s string) (int64, error) {
	re := regexp.MustCompile(`-?\d+`)
	m := re.FindString(s)
	if m == "" {
		return 0, strconv.ErrSyntax
	}
	return strconv.ParseInt(m, 10, 64)
}

// toGiB converts a byte value to GiB with rounding.
func toGiB(i int64) int64 {
	const g = int64(1024 * 1024 * 1024)
	if i <= 0 {
		return 0
	}
	// Round to nearest GiB
	return (i + g/2) / g
}
