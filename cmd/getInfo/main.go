package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	// Ensure base structure (config, logs, etc.) exists before running.
	if err := ensureStructure(); err != nil {
		fmt.Fprintf(os.Stderr, "error during initial setup: %v\n", err)
		os.Exit(1)
	}

	base := exeDir()
	csvPath := filepath.Join(base, CsvName)
	errLog := filepath.Join(base, ErrLogName)

	// Channel to receive system collection result from background goroutine.
	resultCh := make(chan SystemResult, 1)

	// Start system info collection in the background.
	go func() {
		resultCh <- collectSystemInfo()
	}()

	// --- Handle CLI arguments / user input on the main goroutine ---

	args := Args()

	var inAsset, inName, inLocation string
	switch {
	case len(args) == 0:
		// Double-click or no args in terminal → interactive mode
		inAsset, inName, inLocation = PromptForInputs()
	case len(args) >= 3:
		// Positional args: <Asset> <Name> <Location...>
		inAsset = args[0]
		inName = args[1]
		inLocation = strings.TrimSpace(strings.Join(args[2:], " "))
	default:
		// 1 or 2 args → ambiguous, enforce usage for data quality
		PrintUsageAndExit()
	}

	// Wait for system collection to complete.
	sysResult := <-resultCh
	info := sysResult.Info

	now := time.Now().Format("2006-01-02 15:04:05")

	// --- CSV output ---

	f, err := ensureCSVReady(csvPath, Headers)
	if err != nil {
		// If CSV preparation fails, still persist errors from collection if any.
		writeErrors(errLog, append(sysResult.Errors, fmt.Sprintf("csv_prepare: %v", err)))
		fmt.Println("error preparing CSV file:", err)
		waitForEnter()
		return
	}
	defer f.Close()

	row := []string{
		// Identity
		info.SN,
		info.UUID,
		info.MachineGuid,
		inAsset,
		inName,
		inLocation,

		// System / OS / network
		info.Host,
		info.User,
		info.RDP,
		info.IP,
		info.Win,

		// CPU / RAM
		info.CPU,
		info.RAMGB,
		info.RAMType,
		info.SlotUsed,
		info.SlotTotal,
		info.SlotFree,

		// Storage
		info.DiskGB,
		info.FreeGB,
		info.SSD,

		// GPU
		info.GPUPresent,
		info.GPUModel,
		info.GPUVRAM,

		// Remote / security
		info.AnyDeskID,
		info.ADPwdOK,
		info.AVProducts,
		info.BDProduct,

		// Meta
		now,
	}

	if err := appendCSVRow(f, row); err != nil {
		// Append CSV error to the collected error list.
		writeErrors(errLog, append(sysResult.Errors, fmt.Sprintf("csv_append: %v", err)))
		fmt.Println("error writing CSV row:", err)
		waitForEnter()
		return
	}

	// Persist non-fatal errors collected during system info collection.
	writeErrors(errLog, sysResult.Errors)

	// --- One-line-per-column summary in console ---

	printSummary(Headers, row)

	// Keep the terminal open so the user can review the summary.
	waitForEnter()
}

// printSummary prints each column name and its value on a separate line.
func printSummary(headers, row []string) {
	fmt.Println("\n==== Collected data summary ====")
	n := len(headers)
	if len(row) < n {
		n = len(row)
	}
	for i := 0; i < n; i++ {
		fmt.Printf("%s: %s\n", headers[i], row[i])
	}
	fmt.Println("================================")
}

// waitForEnter blocks until ENTER is pressed (keeps terminal open on double-click).
func waitForEnter() {
	fmt.Print("\nPress Enter to exit...")
	_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
}
