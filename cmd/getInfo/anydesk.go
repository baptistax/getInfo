package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"
)

// startAnyDeskService attempts to start AnyDesk-related services.
// Some installations require the service to be running for CLI operations to succeed.
func startAnyDeskService() {
	_, _ = runPS(`Start-Service -Name "AnyDesk" -ErrorAction SilentlyContinue`)
	_, _ = runPS(`Start-Service -Name "AnyDesk Service" -ErrorAction SilentlyContinue`)
	_, _ = runCMD(`sc start "AnyDesk"`)
	_, _ = runCMD(`sc start "AnyDesk Service"`)
}

// findAnyDeskExe tries to locate the AnyDesk executable in PATH and common install folders.
func findAnyDeskExe() string {
	if p, err := exec.LookPath("anydesk.exe"); err == nil {
		return p
	}
	if p, err := exec.LookPath("anydesk"); err == nil {
		return p
	}

	candidates := []string{
		filepath.Join(os.Getenv("ProgramFiles"), "AnyDesk", "AnyDesk.exe"),
		filepath.Join(os.Getenv("ProgramFiles(x86)"), "AnyDesk", "AnyDesk.exe"),
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	for _, base := range []string{os.Getenv("ProgramFiles"), os.Getenv("ProgramFiles(x86)")} {
		if base == "" {
			continue
		}
		entries, err := os.ReadDir(base)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			if strings.Contains(strings.ToLower(e.Name()), "anydesk") {
				dir := filepath.Join(base, e.Name())
				items, _ := os.ReadDir(dir)
				for _, it := range items {
					nameLower := strings.ToLower(it.Name())
					if !it.IsDir() && strings.HasSuffix(nameLower, ".exe") && strings.Contains(nameLower, "anydesk") {
						return filepath.Join(dir, it.Name())
					}
				}
			}
		}
	}

	return ""
}

// anydeskGetID retrieves the AnyDesk ID using the CLI, if available.
func anydeskGetID(c *collector) string {
	exe := findAnyDeskExe()
	if exe == "" {
		c.addErr("anydesk_id", errors.New("AnyDesk executable not found"), "")
		return ""
	}

	out, err := runCmdTimeout(6, exe, "--get-id")
	if err != nil || strings.TrimSpace(out) == "" {
		c.addErr("anydesk_id", errors.New("failed to obtain AnyDesk ID"), "")
		return ""
	}

	re := regexp.MustCompile(`\b\d{9,10}\b`)
	if m := re.FindString(out); m != "" {
		return m
	}

	return strings.TrimSpace(out)
}

// anydeskSetPassword sets the unattended access password for AnyDesk.
// It returns true if it believes the password was applied successfully.
func anydeskSetPassword(c *collector, pwd string) bool {
	password := strings.TrimSpace(pwd)
	if password == "" {
		c.addErr("anydesk_setpwd", errors.New("missing password"), "")
		return false
	}
	if !isAdmin() {
		c.addErr("anydesk_setpwd", errors.New("administrative privileges required"), "")
		return false
	}

	exe := findAnyDeskExe()
	if exe == "" {
		c.addErr("anydesk_setpwd", errors.New("AnyDesk executable not found"), "")
		return false
	}

	startAnyDeskService() // ensure service is running before invoking CLI

	// 1) Preferred: provide password via stdin (does not expose it on the command line)
	{
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, exe, "--set-password")
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

		stdin, err := cmd.StdinPipe()
		if err != nil {
			c.addErr("anydesk_setpwd", err, "stdin_pipe")
			// fall back to argument mode below
		} else {
			var out bytes.Buffer
			cmd.Stdout, cmd.Stderr = &out, &out

			if err := cmd.Start(); err != nil {
				c.addErr("anydesk_setpwd", err, "stdin_start")
			} else {
				_, _ = stdin.Write([]byte(password + "\n"))
				_ = stdin.Close()

				if err := cmd.Wait(); err == nil {
					return true // success via stdin
				}

				// log exit code for visibility (e.g., 9011)
				if ee, ok := err.(*exec.ExitError); ok {
					c.addErr("anydesk_setpwd", fmt.Errorf("exit %d", ee.ExitCode()), "stdin_wait")
				} else {
					c.addErr("anydesk_setpwd", err, "stdin_wait")
				}
			}
		}
	}

	// 2) Fallback: pass password as argument (may expose password in process list)
	if out, err := runCmdTimeout(8, exe, "--set-password", password); err == nil {
		_ = out // discard noisy output
		return true
	} else {
		if ee, ok := err.(*exec.ExitError); ok {
			c.addErr("anydesk_setpwd", fmt.Errorf("exit %d", ee.ExitCode()), "arg")
		} else {
			c.addErr("anydesk_setpwd", err, "arg")
		}
	}

	return false
}
