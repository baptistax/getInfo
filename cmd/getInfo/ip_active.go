package main

import (
	"errors"
	"net"
	"strings"
	"time"
)

// getActiveIPv4 returns the IPv4 address that is actually used by the default route.
// First it uses a UDP "dial" trick (no real packets are sent), then falls back to
// PowerShell, and finally iterates network interfaces while skipping virtual ones.
func getActiveIPv4(c *collector) string {
	d := net.Dialer{Timeout: 2 * time.Second}

	// Preferred: UDP dial to a public IP (no data is sent; only local routing is resolved).
	conn, err := d.Dial("udp", "1.1.1.1:53")
	if err == nil {
		defer conn.Close()

		if la, ok := conn.LocalAddr().(*net.UDPAddr); ok {
			ip := la.IP.To4()
			if ip != nil && !(ip[0] == 169 && ip[1] == 254) {
				return ip.String()
			}
		}
	} else {
		c.addErr("ip_via_udp", err, "")
	}

	// Fallback: adapter with default IPv4 gateway and status "Up" via PowerShell.
	ps := `(Get-NetIPConfiguration | ? { $_.IPv4DefaultGateway -ne $null -and $_.NetAdapter.Status -eq 'Up' } | select -First 1).IPv4Address.IPAddress`
	if out, err := runPS(ps); err == nil && strings.TrimSpace(out) != "" {
		ip := firstLine(out)
		if strings.Count(ip, ".") == 3 && !strings.HasPrefix(ip, "169.254.") {
			return ip
		}
	}

	// Final fallback: iterate network interfaces, skipping virtual/loopback adapters.
	ifaces, err := net.Interfaces()
	if err == nil {
		bad := []string{
			"vethernet", "virtual", "vmnet", "vmware",
			"hyper-v", "vbox", "tap", "tun", "wsl", "loopback",
		}

		for _, in := range ifaces {
			name := strings.ToLower(in.Name)

			skip := (in.Flags&net.FlagUp) == 0 || (in.Flags&net.FlagLoopback) != 0
			if !skip {
				for _, b := range bad {
					if strings.Contains(name, b) {
						skip = true
						break
					}
				}
			}
			if skip {
				continue
			}

			addrs, _ := in.Addrs()
			for _, a := range addrs {
				var ip net.IP
				switch v := a.(type) {
				case *net.IPNet:
					ip = v.IP
				case *net.IPAddr:
					ip = v.IP
				}
				if ip == nil || ip.IsLoopback() {
					continue
				}
				if v4 := ip.To4(); v4 != nil && !(v4[0] == 169 && v4[1] == 254) {
					return v4.String()
				}
			}
		}
	}

	c.addErr("ip_fallback", errors.New("IPv4 address not found"), "")
	return ""
}
