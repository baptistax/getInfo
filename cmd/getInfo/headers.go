package main

// Headers defines the CSV column order.
// Short English names are used to keep spreadsheets clean and readable.
var Headers = []string{
	// Identity
	"SN",          // BIOS serial number
	"UUID",        // SMBIOS UUID
	"MachineGuid", // Windows MachineGuid

	"Asset",    // Asset tag / inventory code
	"Name",     // Person / client / machine label
	"Location", // Physical location (room, branch, etc.)

	// System / OS / network
	"Host", // Hostname
	"User", // Current user (DOMAIN\user)
	"RDP",  // Remote Desktop (MSTSC) status

	"IP",  // Active IP address
	"Win", // Windows version (10 / 11)

	// CPU / RAM
	"CPU",      // CPU model
	"RAM_GB",   // Total RAM (GiB, rounded)
	"RAM_Type", // DDR, DDR2, DDR3, DDR4, DDR5, etc.

	"Slot_Used",  // RAM slots in use
	"Slot_Total", // Total RAM slots
	"Slot_Free",  // Estimated free RAM slots

	// Storage
	"Disk_GB", // System disk size (GiB, rounded)
	"Free_GB", // Free space on system disk (GiB, rounded)
	"SSD",     // "Yes" / "No" if primary disk is SSD

	// GPU
	"GPU",         // "Yes" / "No" if a video controller is present
	"GPU_Model",   // GPU model name
	"GPU_VRAM_GB", // Approximate dedicated VRAM (GiB, rounded)

	// Remote / security
	"AnyDesk_ID", // AnyDesk ID
	"AD_PwdOK",   // "Yes" if AnyDesk password was set successfully

	"Antivirus",  // AV products reported by Windows Security Center
	"BD_Product", // Bitdefender product/edition (if present)

	// Meta
	"Date", // Inventory date
}
