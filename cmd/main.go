package main

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func main() {
	// Require 3 positional args: <Patrimonio> <Nome> <Local...>
	args := Args()
	if len(args) < 3 {
		PrintUsageAndExit()
	}
	inPatr := args[0]
	inNome := args[1]
	inLocal := strings.TrimSpace(strings.Join(args[2:], " "))

	base := exeDir()
	csvPath := filepath.Join(base, CsvName)
	errLog := filepath.Join(base, ErrLogName)

	c := &collector{}

	// --- Coleta (cada função falhando retorna "" e loga o erro) ---
	sn := getSerial(c)
	uuid := getUUIDSMBIOS(c)
	mguid := getMachineGuid(c)
	host := getHostname(c)
	user := getCurrentUser(c)
	mstsc := getRDPUser(c)
	ip := getActiveIPv4(c) // correct active IPv4 (default route)
	win := getWindowsVersion(c)
	cpu := getCPU(c)
	ramGB := getTotalRAMGiB(c)
	used, total, okUsed, okTotal := getRAMSlots(c)

	var slotsUs, slotsTot, slotsLiv string
	if okUsed {
		slotsUs = strconv.FormatInt(used, 10)
	}
	if okTotal {
		slotsTot = strconv.FormatInt(total, 10)
	}
	if okUsed && okTotal && total >= used {
		slotsLiv = strconv.FormatInt(total-used, 10)
	}

	diskGB, freeGB := getDiskSystemGiB(c)
	ssd := getIsSSD(c)
	adID := anydeskGetID(c)

	// Set AnyDesk password (requires admin; manifest should ensure elevation)
	anydeskSetPassword(c, AnyDeskPassword)

	now := time.Now().Format("2006-01-02 15:04:05")

	// --- CSV ---
	f, err := ensureCSVReady(csvPath, Headers)
	if err != nil {
		c.addErr("csv_prepare", err, csvPath)
		writeErrors(errLog, c.errs)
		fmt.Println("erro ao preparar CSV:", err)
		return
	}
	defer f.Close()

	row := []string{
		sn, uuid, mguid, inPatr, inNome, inLocal,
		host, user, mstsc, ip, win, cpu, ramGB,
		slotsUs, slotsTot, slotsLiv, diskGB, freeGB, ssd,
		adID, now,
	}
	if err := appendCSVRow(f, row); err != nil {
		c.addErr("csv_append", err, csvPath)
		writeErrors(errLog, c.errs)
		fmt.Println("erro ao gravar linha no CSV:", err)
		return
	}

	writeErrors(errLog, c.errs)
}
