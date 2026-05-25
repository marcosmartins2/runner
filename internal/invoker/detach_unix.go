//go:build !windows

package invoker

import (
	"os/exec"
	"syscall"
)

// desconectarProcesso configura o processo filho como lider de uma nova sessao,
// permitindo que ele sobreviva ao termino do CLI pai.
func desconectarProcesso(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.Setsid = true
}
