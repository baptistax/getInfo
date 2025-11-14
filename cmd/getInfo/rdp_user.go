package main

import (
	"bufio"
	"strings"
)

// Why: 'query session' has a stable column order; username is the 2nd field.
// We also keep 'query user' as a fallback (username is the 1st field).
func getRDPUser(c *collector) string {
	// 1) query session → match lines with rdp-tcp and state Active/Ativo/Conectado
	if out, err := runCMD("query session"); err == nil && strings.TrimSpace(out) != "" {
		sc := bufio.NewScanner(strings.NewReader(out))
		for sc.Scan() {
			l := strings.TrimSpace(strings.TrimLeft(sc.Text(), ">"))
			ll := strings.ToLower(l)
			if !strings.Contains(ll, "rdp-tcp") {
				continue
			}
			// Fields: SESSIONNAME USERNAME ID STATE ...
			fs := strings.Fields(l)
			if len(fs) >= 4 {
				state := strings.ToLower(fs[3])
				if strings.Contains(state, "active") || strings.Contains(state, "ativo") || strings.Contains(state, "conect") {
					// username is fs[1] for 'query session'
					return fs[1]
				}
			}
		}
	}

	// 2) query user → match lines with rdp-tcp and state active/ativo
	if out, err := runCMD("query user"); err == nil && strings.TrimSpace(out) != "" {
		lines := strings.Split(out, "\n")
		for _, ln := range lines[1:] {
			l := strings.TrimSpace(ln)
			if l == "" {
				continue
			}
			ll := strings.ToLower(l)
			if strings.Contains(ll, "rdp-tcp") && (strings.Contains(ll, "active") || strings.Contains(ll, "ativo") || strings.Contains(ll, "conect")) {
				fs := strings.Fields(l)
				if len(fs) > 0 {
					// username is fs[0] for 'query user'
					return fs[0]
				}
			}
		}
	}

	return ""
}
