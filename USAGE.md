# Usage Guide — `getInfo` PC Inventory Tool

This document describes how to run `getInfo.exe`, what it does during execution, and how to interpret the output CSV and logs.

---

## 1. Running the tool

There are two main ways to run `getInfo`:

### 1.1 Interactive mode (double-click / no arguments)

If you double-click `getInfo.exe` in Explorer, or run it from a terminal **without** arguments:

```bash
getInfo.exe
```

the tool will:

1. Start collecting system information in the background (hardware, OS, GPU, AV, AnyDesk, etc.).
2. Prompt the operator for three values:

   - `Asset` — internal asset tag or inventory ID.
   - `Name` — name of the user/client or a human-friendly label.
   - `Location` — physical location (room, branch, region, etc.).

3. After the operator fills these fields and presses ENTER, it will:
   - Wait for background collection to finish.
   - Append a new row to `inventario.csv`.
   - Print a one-line-per-column summary on screen.
   - Wait for ENTER before closing (so the operator can review the output).

This mode is meant for technicians doing manual inventory in front of the machine.

---

### 1.2 CLI mode (with arguments)

You can also pass information directly via command line arguments:

```bash
getInfo.exe <Asset> <Name> <Location...>
```

Examples:

```bash
getInfo.exe PC-001 "John Doe" "Headquarters - Floor 3"
getInfo.exe NB-202 "Maria Silva" "Branch A - Reception"
```

Rules:

- If there are **3 or more arguments**, `getInfo` will consider:
  - `arg[0]` → Asset
  - `arg[1]` → Name
  - `arg[2..]` (joined with spaces) → Location

- If there are **1 or 2 arguments**, the usage is considered ambiguous and the tool will show a usage message and exit (to avoid half-filled rows).

CLI mode is intended for automated runs (e.g. via scripts, RMM, SCCM, GPO, scheduled tasks).

---

## 2. Output files

By default, the tool reads/writes files **relative to the executable directory**.

### 2.1 `inventario.csv`

- **Location:** same directory as `getInfo.exe` (by default).
- **Behavior:**
  - If the file does not exist or is empty:
    - Create it.
    - Write a UTF-8 BOM.
    - Write `sep=,` on the first line (for Excel delimiter awareness).
    - Write the header row with all column names.
  - If the file already exists:
    - Ensure it has `sep=,` and commas as separators (migrating from older `;`-based formats when needed).
    - Append a new row at the end.

You can open this file directly in Excel or import it into another system.

---

### 2.2 `inventario_erros.txt`

- **Location:** same directory as `getInfo.exe` (by default).
- **Content:** one line per non-fatal error encountered during collection, with timestamp and context.

Example entries:

```text
2025-01-10T14:32:15-0300 gpu: not found | no Win32_VideoController entries
2025-01-10T14:32:15-0300 anydesk_id: AnyDesk executable not found
```

If the file cannot be written, a minimal message is printed to `stderr`.

This log is useful when some columns show up empty and you want to know why (e.g. WMI restrictions, missing services, no AnyDesk install, etc.).

---

### 2.3 Config and auxiliary directories

At startup, the tool ensures:

- `config/` exists
  - If `config/config.json` is missing, a default configuration file is created.
- `output/` exists
  - A `.gitignore` is written inside, ignoring all contents except `.gitignore`.
- `logs/` exists
  - A `.gitignore` is written inside, ignoring all contents except `.gitignore`.

These are preparation steps for future extensions (e.g. alternative outputs, structured logs, additional formats).  
For now, the main output remains `inventario.csv` and `inventario_erros.txt`.

---

## 3. CSV columns and meaning

The header row is defined in `headers.go` and looks like this:

```text
SN,UUID,MachineGuid,
Asset,Name,Location,
Host,User,RDP,
IP,Win,
CPU,RAM_GB,RAM_Type,
Slot_Used,Slot_Total,Slot_Free,
Disk_GB,Free_GB,SSD,
GPU,GPU_Model,GPU_VRAM_GB,
AnyDesk_ID,AD_PwdOK,
Antivirus,BD_Product,
Date
```

Below is a concise description of each column:

### Identity

- **SN**  
  BIOS serial number (from `Win32_BIOS` or `wmic bios`).

- **UUID**  
  SMBIOS UUID (from `Win32_ComputerSystemProduct` or `wmic csproduct`).

- **MachineGuid**  
  Windows `MachineGuid` from the registry (`HKLM\SOFTWARE\Microsoft\Cryptography`).

### Asset / operator

- **Asset**  
  Asset tag or inventory ID (user-provided or CLI argument).

- **Name**  
  Name of the user/client or a human-friendly machine label.

- **Location**  
  Physical location, branch, department, etc.

### System / OS / network

- **Host**  
  Hostname returned by `os.Hostname()`.

- **User**  
  Current user (e.g. `DOMAIN\jdoe`).

- **RDP**  
  Remote Desktop information (current RDP user, when available).

- **IP**  
  IPv4 address in use on the default route (best-effort, non-virtual interface when possible).

- **Win**  
  `"10"` or `"11"` depending on the detected Windows edition.

### CPU / RAM

- **CPU**  
  CPU model string (e.g. `Intel(R) Core(TM) i7-8700 CPU @ 3.20GHz`).

- **RAM_GB**  
  Total physical memory (GiB), rounded.

- **RAM_Type**  
  RAM type derived from SMBIOS or memory chip info when available. Examples:
  - `DDR3`
  - `DDR4`
  - `DDR5`
  - `LPDDR3`
  - `LPDDR4`

- **Slot_Used**  
  Number of populated memory slots.

- **Slot_Total**  
  Total number of memory slots on the motherboard.

- **Slot_Free**  
  `Slot_Total - Slot_Used` when both values are available.

### Storage

- **Disk_GB**  
  Size of the system drive (e.g. `C:`) in GiB, rounded.

- **Free_GB**  
  Free space on the system drive in GiB, rounded.

- **SSD**  
  `"Yes"` if the primary disk is reported as SSD by `Get-PhysicalDisk` or model name heuristics, `"No"` otherwise. Empty string if unknown.

### GPU

- **GPU**  
  `"Yes"` if a `Win32_VideoController` entry is found, `"No"` if none is detected.

- **GPU_Model**  
  Video controller name (e.g. `NVIDIA GeForce RTX 3060`, `AMD Radeon RX 580`, `Intel(R) HD Graphics`).

- **GPU_VRAM_GB**  
  Approximate dedicated VRAM in GiB, rounded (derived from `AdapterRAM`).

### Remote / security

- **AnyDesk_ID**  
  AnyDesk ID obtained via the `anydesk.exe --get-id` CLI, if AnyDesk is installed and the CLI is accessible.

- **AD_PwdOK**  
  `"Yes"` if `getInfo` believes it successfully set the AnyDesk unattended password via:
  - stdin to `--set-password`, or
  - the fallback argument-based mode.

  `"No"` if all attempts failed, or if AnyDesk is not installed.

- **Antivirus**  
  Semicolon-separated list of antivirus products registered in Windows Security Center (namespace `root/SecurityCenter2` or legacy `root/SecurityCenter`).

- **BD_Product**  
  Bitdefender product/edition name when any entry contains `"Bitdefender"`.  
  Empty string when Bitdefender is not present.

### Meta

- **Date**  
  Local timestamp for the inventory run (`YYYY-MM-DD HH:MM:SS`).

---

## 4. AnyDesk password handling

The default AnyDesk password is defined in `config.go` as:

```go
const AnyDeskPassword = "CHANGE_ME"
```

You must:

1. Change this value before deploying the tool.
2. Consider moving it out of source code (e.g. to an environment variable or a private `config.json`) if the repository is public.

The tool tries:

1. `anydesk.exe --set-password` with the password written to stdin (preferred).
2. A fallback mode where the password is passed as a CLI argument (less private; used only if necessary).

If both attempts fail, `AD_PwdOK` is `"No"`, and an error is recorded in `inventario_erros.txt`.

---

## 5. Antivirus and Bitdefender detection

The tool uses the `AntiVirusProduct` class from:

- `root/SecurityCenter2` (Windows 10/11), or
- `root/SecurityCenter` (older / legacy).

It collects:

- All `displayName` entries → joined into the **Antivirus** column.
- The first entry containing `"Bitdefender"` (case-insensitive) → **BD_Product**.

This is a high-level view; it does not attempt to introspect license keys, subscription status, or configuration details.

---

## 6. Automation scenarios

A few possible ways to use `getInfo` in larger environments:

- **Logon scripts**: run `getInfo.exe "%COMPUTERNAME%" "%USERNAME%" "Branch X"` at login to track active users.
- **RMM / SCCM / Intune**: deploy `getInfo.exe` and run it with CLI arguments so that:
  - `Asset` comes from the RMM asset ID.
  - `Name` and `Location` are pre-populated from your CMDB.
- **On-site audits**: technicians can use the interactive mode, typing only human-friendly labels while the tool collects all technical details in the background.

In all cases, the result is a single, growing `inventario.csv` that can be imported into your own systems.

---

## 7. Cleaning before publishing

If you plan to publish this repository (e.g. on GitHub), consider:

- Adding a root `.gitignore` that excludes:
  - `inventario.csv`
  - `inventario_*.txt`
  - `*.log`
  - `config/config.json`
  - `output/`
  - `logs/`
  - `getInfo.exe` and other binaries

- Resetting any test data:
  - Remove CSV files, error logs and temporary configs before committing.
- Ensuring that the AnyDesk password constant **is not a real production password**.

---
