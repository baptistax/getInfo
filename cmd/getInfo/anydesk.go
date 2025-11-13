package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"
)

func findAnyDeskExe() string {
	if p, err := exec.LookPath("anydesk.exe"); err == nil {
		return p
	}
	if p, err := exec.LookPath("anydesk"); err == nil {
		return p
	}
	cands := []string{
		filepath.Join(os.Getenv("ProgramFiles"), "AnyDesk", "AnyDesk.exe"),
		filepath.Join(os.Getenv("ProgramFiles(x86)"), "AnyDesk", "AnyDesk.exe"),
	}
	for _, p := range cands {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	// Lightweight scan first level (helps branded installers)
	for _, base := range []string{os.Getenv("ProgramFiles"), os.Getenv("ProgramFiles(x86)")} {
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
					n := strings.ToLower(it.Name())
					if !it.IsDir() && strings.HasSuffix(n, ".exe") && strings.Contains(n, "anydesk") {
						return filepath.Join(dir, it.Name())
					}
				}
			}
		}
	}
	return ""
}

func anydeskGetID(c *collector) string {
	exe := findAnyDeskExe()
	if exe == "" {
		c.addErr("anydesk_id", errors.New("anydesk nao encontrado"), "")
		return ""
	}
	out, err := runCmdTimeout(6, exe, "--get-id")
	if err != nil || strings.TrimSpace(out) == "" {
		c.addErr("anydesk_id", errors.New("falha ao obter id"), "")
		return ""
	}
	re := regexp.MustCompile(`\b\d{9,10}\b`)
	if m := re.FindString(out); m != "" {
		return m
	}
	return strings.TrimSpace(out)
}

func anydeskSetPassword(c *collector, pwd string) {
	if strings.TrimSpace(pwd) == "" {
		c.addErr("anydesk_setpwd", errors.New("senha ausente"), "")
		return
	}
	if !isAdmin() {
		// With manifest this should not happen; log just in case.
		c.addErr("anydesk_setpwd", errors.New("requer administrador"), "")
		return
	}
	exe := findAnyDeskExe()
	if exe == "" {
		c.addErr("anydesk_setpwd", errors.New("anydesk nao encontrado"), "")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, exe, "--set-password")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		c.addErr("anydesk_setpwd", err, "stdin")
		return
	}
	var out bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &out

	if err := cmd.Start(); err != nil {
		c.addErr("anydesk_setpwd", err, "start")
		return
	}
	_, _ = stdin.Write([]byte(pwd + "\n"))
	_ = stdin.Close()
	if err := cmd.Wait(); err != nil {
		c.addErr("anydesk_setpwd", err, "")
	}
}
