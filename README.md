# PC Inventory Tool (`getInfo`)

`getInfo` is a small Windows-only inventory tool written in Go.  
It collects basic hardware, OS and security information from a workstation and appends everything to a single CSV file, one row per run.

The goal is to be:

- Simple to deploy (single `getInfo.exe` + optional manifest).
- Easy to use for non-technical operators (interactive mode).
- Safe for spreadsheets (stable, consistent column layout).
- Reasonably transparent about what it collects.

---

## Features

### Data collected

On each run, `getInfo` collects:

#### Identity

- **SN** — BIOS serial number.
- **UUID** — SMBIOS UUID.
- **MachineGuid** — Windows `MachineGuid`.

#### Asset / operator

- **Asset** — asset tag or inventory code (user input or CLI argument).
- **Name** — person / client / machine label (user input or CLI argument).
- **Location** — physical location (room, branch, region, etc.).

#### System / OS / network

- **Host** — machine hostname.
- **User** — current user (`DOMAIN\user` or local).
- **RDP** — Remote Desktop / MSTSC information (current RDP user, when available).
- **IP** — currently active IPv4 used by the default route.
- **Win** — Windows major version (`10` or `11`).

#### CPU / RAM

- **CPU** — CPU model (e.g. `Intel(R) Core(TM) i5-8600K`).
- **RAM_GB** — total physical memory in GiB (rounded).
- **RAM_Type** — memory type (`DDR2`, `DDR3`, `DDR4`, `DDR5`, `LPDDR`, etc.) when SMBIOS exposes it.
- **Slot_Used** — number of RAM slots currently populated.
- **Slot_Total** — total number of RAM slots on the board.
- **Slot_Free** — estimated free slots (`total - used` when both are known).

#### Storage

- **Disk_GB** — system drive size in GiB (rounded).
- **Free_GB** — free space on the system drive in GiB (rounded).
- **SSD** — `"Yes"` if the primary disk is reported as SSD, `"No"` otherwise (best-effort).

#### GPU

- **GPU** — `"Yes"` if a video controller was detected, `"No"` otherwise.
- **GPU_Model** — GPU model name (e.g. `NVIDIA GeForce GTX 1660 Ti`, `Intel(R) UHD Graphics`).
- **GPU_VRAM_GB** — approximate dedicated VRAM in GiB (rounded), when reported.

#### Remote / security

- **AnyDesk_ID** — AnyDesk ID (when AnyDesk is installed and the CLI is available).
- **AD_PwdOK** — `"Yes"` if `getInfo` believes it successfully set the AnyDesk unattended password, `"No"` otherwise.
- **Antivirus** — semicolon-separated list of antivirus products reported by Windows Security Center.
- **BD_Product** — Bitdefender product/edition name, when Bitdefender is present (e.g. `Bitdefender Total Security`).

#### Meta

- **Date** — local timestamp of the inventory run (`YYYY-MM-DD HH:MM:SS`).

All of these fields are always written in the same order, with the same headers, so the CSV remains stable over time.

---

## How it works (high level)

1. On startup, `getInfo` ensures some basic structure exists:
   - `config/` with a default `config.json` (if missing).
   - `output/` and `logs/` directories, each with a local `.gitignore` to avoid committing runtime files.

2. It starts a background goroutine that:
   - Collects all system/OS/hardware/security data (SN, UUID, RAM, GPU, AnyDesk, AV, etc.).
   - Uses PowerShell / CIM / WMI / WMIC as needed.
   - Accumulates non-fatal errors in memory for logging.

3. Meanwhile, on the main goroutine:
   - If there are no CLI arguments, it prompts the operator for:
     - `Asset`
     - `Name`
     - `Location`
   - If there are enough CLI arguments, it uses them directly.

4. Once user input is ready, it waits for the background collection to finish and:
   - Opens or creates `inventario.csv` in the executable directory.
   - Ensures the file has:
     - UTF-8 BOM for Excel compatibility.
     - `sep=,` header line.
     - The CSV header row defined in `headers.go`.
   - Appends a new row with all collected data.

5. It saves non-fatal errors to `inventario_erros.txt`, if any.

6. Finally, it prints a one-line-per-column summary in the console and waits for ENTER before closing (so operators can review data when running via double-click).

---

## Project structure

A typical layout for this repository looks like:

```text
.
├─ cmd/
│  └─ getinfo/
│     ├─ main.go
│     ├─ admin.go
│     ├─ anydesk.go
│     ├─ collect_disk.go
│     ├─ collect_hw.go
│     ├─ collect_os.go
│     ├─ collect_av.go
│     ├─ collect_gpu.go
│     ├─ systeminfo.go
│     ├─ csvfile.go
│     ├─ headers.go
│     ├─ helpers.go
│     ├─ ip_active.go
│     ├─ logging.go
│     ├─ machineguid.go
│     └─ getInfo.exe.manifest (optional, for elevation)
├─ config/
│  └─ config.example.json
├─ output/
├─ logs/
├─ go.mod
└─ README.md
```

Runtime files like `inventario.csv`, `inventario_erros.txt`, `config/config.json`, and logs are generated at run time and should generally be ignored by Git.

---

## Requirements

- **OS**: Windows 10 or 11.
- **Runtime**:
  - PowerShell available (default on Windows 10/11).
  - WMI/CIM not heavily restricted by security policy (for hardware queries).
- **Build**:
  - Go 1.20+ (older versions may also work, but are not tested).

---

## Building

From the repository root:

```bash
go mod tidy
go build -o getInfo.exe ./cmd/getinfo
```

This will:

- Resolve Go module dependencies (only standard library is required).
- Produce a `getInfo.exe` binary in the repository root.

If you prefer to build inside the `cmd/getinfo` directory:

```bash
cd cmd/getinfo
go build -o getInfo.exe .
```

> **Note:** do not use `go build main.go` – always build the whole package (`go build .` or `go build ./cmd/getinfo`).

---

## Elevation (running as Administrator)

To always run `getInfo.exe` as Administrator:

1. Place `getInfo.exe` and `getInfo.exe.manifest` side by side.
2. The example manifest in this repository:
   - Requests `requireAdministrator`.
   - Declares compatibility with Windows 10/11.

Windows automatically associates the manifest file with the executable when they share the same base name in the same directory.

---

## Configuration

On first run, `getInfo` writes a default `config/config.json` if none exists.  
The default content looks like:

```json
{
  "collection": {
    "hardware": true,
    "gpu": true,
    "ram": true,
    "osInfo": true,
    "installedSoftware": false,
    "remoteDesktop": true,
    "antivirus": true,
    "network": false
  },
  "output": {
    "excel": true,
    "csv": false,
    "json": false,
    "showSummaryInConsole": true
  },
  "ui": {
    "interactiveIfNoArgs": true
  }
}
```

At the moment, these fields are mostly reserved for future expansion and documentation.  
The core behavior is:

- Always append to `inventario.csv`.
- Always print a summary to the console.
- Always collect the fields described above.

If you need to keep a template under version control, use `config/config.example.json` in the repo and **ignore** `config/config.json` in `.gitignore`.

---

## Security notes

- The AnyDesk password used by the tool is defined as a constant in `config.go`:

  ```go
  const AnyDeskPassword = "CHANGE_ME"
  ```

  You **must** change this value before using the tool in a real environment.

- Consider moving the password to:
  - `config/config.json` (not committed to Git), or
  - an environment variable or another secure input channel.

- The tool attempts to set the AnyDesk password using:
  - `anydesk.exe --set-password` with stdin (preferred, avoids password in process list).
  - A fallback passing the password as an argument (only when necessary).

- The output CSV and error log may contain:
  - hostnames, usernames, IPs, serial numbers, etc.,
  - antivirus and Bitdefender product names,
  - AnyDesk IDs.

  These files should be treated as internal inventory data and **not** committed to a public repository.

---

## License

Add your preferred license here (MIT, Apache-2.0, etc.).
