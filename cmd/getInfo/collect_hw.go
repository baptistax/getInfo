package main

import (
	"runtime"
	"strings"
)

// getSerial retrieves the BIOS serial number using CIM first, then WMIC as fallback.
func getSerial(c *collector) string {
	if out, err := runPS(`(Get-CimInstance -ClassName Win32_BIOS).SerialNumber`); err == nil && strings.TrimSpace(out) != "" {
		return firstLine(out)
	}

	out2, err2 := runCMD(`wmic bios get serialnumber /value`)
	if err2 == nil && strings.TrimSpace(out2) != "" {
		for _, ln := range strings.Split(out2, "\n") {
			line := strings.TrimSpace(ln)
			if strings.HasPrefix(strings.ToLower(line), "serialnumber=") {
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					return strings.TrimSpace(parts[1])
				}
			}
		}
	}

	c.addErr("serial", ErrNotFound, "")
	return ""
}

// getUUIDSMBIOS retrieves the SMBIOS UUID using CIM first, then WMIC as fallback.
func getUUIDSMBIOS(c *collector) string {
	if out, err := runPS(`(Get-CimInstance Win32_ComputerSystemProduct).UUID`); err == nil && strings.TrimSpace(out) != "" {
		return firstLine(out)
	}

	out2, err2 := runCMD(`wmic csproduct get UUID /value`)
	if err2 == nil && strings.TrimSpace(out2) != "" {
		for _, ln := range strings.Split(out2, "\n") {
			line := strings.TrimSpace(ln)
			if strings.HasPrefix(strings.ToLower(line), "uuid=") {
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					return strings.TrimSpace(parts[1])
				}
			}
		}
	}

	c.addErr("uuid_smbios", ErrNotFound, "")
	return ""
}

// getCPU returns the CPU name (model string).
func getCPU(c *collector) string {
	if out, err := runPS(`(Get-CimInstance Win32_Processor | Select-Object -First 1 -ExpandProperty Name)`); err == nil && strings.TrimSpace(out) != "" {
		return firstLine(out)
	}

	out2, err2 := runCMD(`wmic cpu get Name /value`)
	if err2 == nil && strings.TrimSpace(out2) != "" {
		for _, ln := range strings.Split(out2, "\n") {
			line := strings.TrimSpace(ln)
			if strings.HasPrefix(strings.ToLower(line), "name=") {
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					return strings.TrimSpace(parts[1])
				}
			}
		}
	}

	c.addErr("cpu", ErrNotFound, runtime.GOARCH)
	return ""
}

// getTotalRAMGiB returns the total physical memory in GiB as a string.
func getTotalRAMGiB(c *collector) string {
	if out, err := runPS(`(Get-CimInstance Win32_ComputerSystem).TotalPhysicalMemory`); err == nil && strings.TrimSpace(out) != "" {
		if v, err := parseInt64Any(out); err == nil {
			return strconvFormatInt(toGiB(v))
		}
	}

	out2, err2 := runCMD(`wmic os get TotalVisibleMemorySize /value`)
	if err2 == nil && strings.TrimSpace(out2) != "" {
		for _, ln := range strings.Split(out2, "\n") {
			line := strings.TrimSpace(ln)
			if strings.HasPrefix(strings.ToLower(line), "totalvisiblememorysize=") {
				if vKB, err := parseInt64Any(line); err == nil {
					return strconvFormatInt(toGiB(vKB * 1024))
				}
			}
		}
	}

	c.addErr("ram_total", ErrNotFound, "")
	return ""
}

// getRAMType returns a human-readable RAM type string (e.g. "DDR3", "DDR4").
func getRAMType(c *collector) string {
	// Preferred: SMBIOSMemoryType via CIM
	if out, err := runPS(`(Get-CimInstance Win32_PhysicalMemory | Select-Object -First 1 -ExpandProperty SMBIOSMemoryType)`); err == nil && strings.TrimSpace(out) != "" {
		if v, err := parseInt64Any(out); err == nil {
			if t := ramTypeFromCode(v); t != "" {
				return t
			}
		}
	}

	// Fallback: MemoryType via CIM
	if out, err := runPS(`(Get-CimInstance Win32_PhysicalMemory | Select-Object -First 1 -ExpandProperty MemoryType)`); err == nil && strings.TrimSpace(out) != "" {
		if v, err := parseInt64Any(out); err == nil {
			if t := ramTypeFromCode(v); t != "" {
				return t
			}
		}
	}

	// Fallback: WMIC SMBIOSMemoryType
	if out, err := runCMD(`wmic memorychip get SMBIOSMemoryType /value`); err == nil && strings.TrimSpace(out) != "" {
		for _, ln := range strings.Split(out, "\n") {
			line := strings.TrimSpace(ln)
			if strings.HasPrefix(strings.ToLower(line), "smbiosmemorytype=") {
				if v, err := parseInt64Any(line); err == nil {
					if t := ramTypeFromCode(v); t != "" {
						return t
					}
				}
			}
		}
	}

	// Fallback: WMIC MemoryType
	if out, err := runCMD(`wmic memorychip get MemoryType /value`); err == nil && strings.TrimSpace(out) != "" {
		for _, ln := range strings.Split(out, "\n") {
			line := strings.TrimSpace(ln)
			if strings.HasPrefix(strings.ToLower(line), "memorytype=") {
				if v, err := parseInt64Any(line); err == nil {
					if t := ramTypeFromCode(v); t != "" {
						return t
					}
				}
			}
		}
	}

	c.addErr("ram_type", ErrNotFound, "")
	return ""
}

// ramTypeFromCode converts SMBIOS/MemoryType numeric codes to a DDR label.
func ramTypeFromCode(code int64) string {
	switch code {
	case 20:
		return "DDR"
	case 21:
		return "DDR2"
	case 22:
		return "DDR2 FB-DIMM"
	case 24:
		return "DDR3"
	case 26:
		return "DDR4"
	case 27:
		return "LPDDR"
	case 28:
		return "LPDDR2"
	case 29:
		return "LPDDR3"
	case 30:
		return "LPDDR4"
	case 31:
		return "LPDDR4X"
	case 34:
		return "DDR5"
	default:
		return ""
	}
}

// getRAMSlots returns the number of used and total RAM slots, with flags for each value.
func getRAMSlots(c *collector) (used int64, total int64, okUsed bool, okTotal bool) {
	// Used slots via CIM
	if out, err := runPS(`(Get-CimInstance Win32_PhysicalMemory | Measure-Object).Count`); err == nil && strings.TrimSpace(out) != "" {
		if v, err := parseInt64Any(out); err == nil {
			used, okUsed = v, true
		}
	}

	// Total slots via CIM
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

	// Fallbacks for "used"
	if !okUsed {
		if out, err := runCMD(`wmic memorychip get banklabel`); err == nil && strings.TrimSpace(out) != "" {
			lines := strings.Split(strings.TrimSpace(out), "\n")
			if len(lines) > 1 {
				cnt := 0
				for _, ln := range lines[1:] {
					if strings.TrimSpace(ln) != "" {
						cnt++
					}
				}
				if cnt > 0 {
					used, okUsed = int64(cnt), true
				}
			}
		}
	}

	// Fallbacks for "total"
	if !okTotal {
		if out, err := runCMD(`wmic memphysical get memorydevices /value`); err == nil && strings.TrimSpace(out) != "" {
			for _, ln := range strings.Split(out, "\n") {
				line := strings.TrimSpace(ln)
				if strings.HasPrefix(strings.ToLower(line), "memorydevices=") {
					if v, err := parseInt64Any(line); err == nil && v > 0 {
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
