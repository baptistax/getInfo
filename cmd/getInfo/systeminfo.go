package main

// SystemInfo groups all automatically collected system data.
// User-provided fields (Asset, Name, Location) stay outside.
type SystemInfo struct {
	SN          string
	UUID        string
	MachineGuid string

	Host string
	User string
	RDP  string
	IP   string
	Win  string

	CPU     string
	RAMGB   string
	RAMType string

	SlotUsed  string
	SlotTotal string
	SlotFree  string

	DiskGB string
	FreeGB string
	SSD    string

	GPUPresent string
	GPUModel   string
	GPUVRAM    string

	AnyDeskID string
	ADPwdOK   string

	AVProducts string
	BDProduct  string
}

// SystemResult bundles SystemInfo with collected non-fatal error logs.
type SystemResult struct {
	Info   SystemInfo
	Errors []string
}

// collectSystemInfo performs all hardware/OS/network/AV collection.
// It is safe to run in a separate goroutine while the main goroutine handles user input.
func collectSystemInfo() SystemResult {
	c := &collector{}
	var info SystemInfo

	// Identity
	info.SN = getSerial(c)
	info.UUID = getUUIDSMBIOS(c)
	info.MachineGuid = getMachineGuid(c)

	// System / OS / network
	info.Host = getHostname(c)
	info.User = getCurrentUser(c)
	info.RDP = getRDPUser(c)
	info.IP = getActiveIPv4(c)
	info.Win = getWindowsVersion(c)

	// CPU / RAM
	info.CPU = getCPU(c)
	info.RAMGB = getTotalRAMGiB(c)
	info.RAMType = getRAMType(c)

	used, total, okUsed, okTotal := getRAMSlots(c)
	if okUsed {
		info.SlotUsed = strconvFormatInt(used)
	}
	if okTotal {
		info.SlotTotal = strconvFormatInt(total)
	}
	if okUsed && okTotal && total >= used {
		info.SlotFree = strconvFormatInt(total - used)
	}

	// Storage
	info.DiskGB, info.FreeGB = getDiskSystemGiB(c)
	info.SSD = getIsSSD(c)

	// GPU
	info.GPUPresent, info.GPUModel, info.GPUVRAM = getGPUInfo(c)

	// AnyDesk
	info.AnyDeskID = anydeskGetID(c)
	info.ADPwdOK = "No"
	if anydeskSetPassword(c, AnyDeskPassword) {
		info.ADPwdOK = "Yes"
	}

	// Antivirus
	info.AVProducts, info.BDProduct = getAntivirusInfo(c)

	return SystemResult{
		Info:   info,
		Errors: c.errs,
	}
}
