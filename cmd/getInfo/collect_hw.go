package main

import (
	"runtime"
	"strings"
)

func getSerial(c *collector) string {
	if out, err := runPS(`(Get-CimInstance -ClassName Win32_BIOS).SerialNumber`); err == nil && strings.TrimSpace(out) != "" {
		return firstLine(out)
	}
	out2, err2 := runCMD(`wmic bios get serialnumber /value`)
	if err2 == nil {
		for _, ln := range strings.Split(out2, "\n") {
			if strings.HasPrefix(strings.ToLower(strings.TrimSpace(ln)), "serialnumber=") {
				return strings.TrimSpace(strings.SplitN(ln, "=", 2)[1])
			}
		}
	}
	c.addErr("serial", ErrNotFound, "")
	return ""
}

func getUUIDSMBIOS(c *collector) string {
	if out, err := runPS(`(Get-CimInstance Win32_ComputerSystemProduct).UUID`); err == nil && strings.TrimSpace(out) != "" {
		return firstLine(out)
	}
	out2, err2 := runCMD(`wmic csproduct get UUID /value`)
	if err2 == nil && strings.TrimSpace(out2) != "" {
		for _, ln := range strings.Split(out2, "\n") {
			if strings.HasPrefix(strings.ToLower(strings.TrimSpace(ln)), "uuid=") {
				return strings.TrimSpace(strings.SplitN(ln, "=", 2)[1])
			}
		}
	}
	c.addErr("uuid_smbios", ErrNotFound, "")
	return ""
}

func getCPU(c *collector) string {
	if out, err := runPS(`(Get-CimInstance Win32_Processor | Select-Object -First 1 -ExpandProperty Name)`); err == nil && strings.TrimSpace(out) != "" {
		return firstLine(out)
	}
	out2, err2 := runCMD(`wmic cpu get Name /value`)
	if err2 == nil && out2 != "" {
		for _, ln := range strings.Split(out2, "\n") {
			if strings.HasPrefix(strings.ToLower(strings.TrimSpace(ln)), "name=") {
				return strings.TrimSpace(strings.SplitN(ln, "=", 2)[1])
			}
		}
	}
	c.addErr("cpu", ErrNotFound, runtime.GOARCH)
	return ""
}

func getTotalRAMGiB(c *collector) string {
	if out, err := runPS(`(Get-CimInstance Win32_ComputerSystem).TotalPhysicalMemory`); err == nil && out != "" {
		if v, err := parseInt64Any(out); err == nil {
			return ToStr(toGiB(v))
		}
	}
	out2, err2 := runCMD(`wmic os get TotalVisibleMemorySize /value`)
	if err2 == nil {
		for _, ln := range strings.Split(out2, "\n") {
			if strings.HasPrefix(strings.ToLower(strings.TrimSpace(ln)), "totalvisiblememorysize=") {
				if vKB, err := parseInt64Any(ln); err == nil {
					return ToStr(toGiB(vKB * 1024))
				}
			}
		}
	}
	c.addErr("ram_total", ErrNotFound, "")
	return ""
}

func getRAMSlots(c *collector) (used int64, total int64, okUsed bool, okTotal bool) {
	// Used
	if out, err := runPS(`(Get-CimInstance Win32_PhysicalMemory | Measure-Object).Count`); err == nil && strings.TrimSpace(out) != "" {
		if v, err := parseInt64Any(out); err == nil {
			used, okUsed = v, true
		}
	}
	// Total
	if out, err := runPS(`(Get-CimInstance Win32_PhysicalMemoryArray | Select-Object -ExpandProperty MemoryDevices)`); err == nil && strings.TrimSpace(out) != "" {
		var sum int64
		for _, ln := range strings.Split(strings.TrimSpace(out), "\n") {
			if v, err := parseInt64Any(ln); err == nil {
				sum += v
			}
		}
		if sum > 0 {
			total, okTotal = sum, true
		}
	}
	// Fallbacks
	if !okUsed {
		if out, err := runCMD(`wmic memorychip get banklabel`); err == nil && strings.TrimSpace(out) != "" {
			cnt := 0
			for _, ln := range strings.Split(strings.TrimSpace(out), "\n")[1:] {
				if strings.TrimSpace(ln) != "" {
					cnt++
				}
			}
			if cnt > 0 {
				used, okUsed = int64(cnt), true
			}
		}
	}
	if !okTotal {
		if out, err := runCMD(`wmic memphysical get memorydevices /value`); err == nil {
			for _, ln := range strings.Split(out, "\n") {
				if strings.HasPrefix(strings.ToLower(strings.TrimSpace(ln)), "memorydevices=") {
					if v, err := parseInt64Any(ln); err == nil && v > 0 {
						total, okTotal = v, true
					}
				}
			}
		}
	}
	if !okUsed {
		c.addErr("ram_slots_usados", ErrNotFound, "")
	}
	if !okTotal {
		c.addErr("ram_slots_total", ErrNotFound, "")
	}
	return
}

// small helpers (local to this file)
func ToStr(i int64) string { return strconvFormatInt(i) }
