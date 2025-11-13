package main

import (
	"os"
	"regexp"
	"strconv"
	"strings"
)

func strconvFormatInt(i int64) string { return strconv.FormatInt(i, 10) }

func getHostname(c *collector) string {
	h, err := os.Hostname()
	if err != nil {
		c.addErr("hostname", err, "")
		return ""
	}
	return strings.TrimSpace(h)
}

func getCurrentUser(c *collector) string {
	out, err := runCMD("whoami")
	if err == nil && out != "" {
		return strings.TrimSpace(out)
	}
	u := os.Getenv("USERNAME")
	if u == "" {
		c.addErr("usuario", err, "USERNAME vazio")
	}
	return u
}

func getWindowsVersion(c *collector) string {
	out, err := runCMD(`reg query "HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion" /v ProductName`)
	if err == nil && out != "" {
		for _, ln := range strings.Split(out, "\n") {
			if strings.Contains(ln, "ProductName") {
				idx := strings.Index(ln, "REG_SZ")
				if idx >= 0 {
					product := strings.TrimSpace(ln[idx+len("REG_SZ"):])
					switch {
					case strings.Contains(product, "Windows 11"):
						return "11"
					case strings.Contains(product, "Windows 10"):
						return "10"
					}
				}
			}
		}
	}
	out2, _ := runCMD("ver")
	if strings.Contains(out2, "Version 10.") {
		return "10"
	}
	if strings.Contains(out2, "Version 11.") {
		return "11"
	}
	return ""
}

// --- helpers (shared across files) ---

func firstLine(s string) string {
	parts := strings.Split(strings.TrimSpace(s), "\n")
	if len(parts) == 0 {
		return ""
	}
	return strings.TrimSpace(parts[0])
}

func parseInt64Any(s string) (int64, error) {
	re := regexp.MustCompile(`-?\d+`)
	m := re.FindString(s)
	if m == "" {
		return 0, strconv.ErrSyntax
	}
	return strconv.ParseInt(m, 10, 64)
}

func toGiB(i int64) int64 {
	const g = int64(1024 * 1024 * 1024)
	if i <= 0 {
		return 0
	}
	return (i + g/2) / g // round
}
