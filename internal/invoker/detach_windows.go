//go:build windows

package invoker

import (
	"os/exec"
	"syscall"
)

const (
	detachedProcess     = 0x00000008
	createNewProcessGrp = 0x00000200
)

// desconectarProcesso isola o filho em um novo grupo de processos para que o
// CLI pai possa encerrar sem matar o servidor do assinador.
func desconectarProcesso(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.CreationFlags = detachedProcess | createNewProcessGrp
}
