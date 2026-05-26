//go:build windows

package main

import (
	"os/exec"
	"syscall"
)

const (
	detachedProcess     = 0x00000008
	createNewProcessGrp = 0x00000200
)

func desconectarProcesso(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.CreationFlags = detachedProcess | createNewProcessGrp
}
