package main

import (
	"bufio"
	"strings"
)

func getRDPUser(c *collector) string {
	out, err := runCMD("query user")
	if err == nil && out != "" {
		lines := strings.Split(out, "\n")
		for _, ln := range lines[1:] {
			l := strings.TrimSpace(ln)
			if l == "" {
				continue
			}
			ll := strings.ToLower(l)
			if strings.Contains(ll, "rdp") && (strings.Contains(ll, "active") || strings.Contains(ll, "ativo")) {
				fs := strings.Fields(l)
				if len(fs) > 0 {
					return fs[0]
				}
			}
		}
	}
	out2, err2 := runCMD("query session")
	if err2 == nil && out2 != "" {
		sc := bufio.NewScanner(strings.NewReader(out2))
		for sc.Scan() {
			l := strings.ToLower(sc.Text())
			if strings.Contains(l, "rdp-tcp") && (strings.Contains(l, "active") || strings.Contains(l, "ativo")) {
				fs := strings.Fields(sc.Text())
				if len(fs) > 0 {
					return fs[0]
				}
			}
		}
	}
	return ""
}
