package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestVersionViaGoRun(t *testing.T) {
	if testing.Short() {
		t.Skip("modo curto")
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("comando go indisponivel")
	}
	_, arquivo, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("nao foi possivel determinar diretorio")
	}
	cmd := exec.Command("go", "run", ".", "version")
	cmd.Dir = filepath.Dir(arquivo)
	cmd.Env = os.Environ()
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("go run falhou: %v\nstdout=%s\nstderr=%s", err, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "dev") {
		t.Errorf("esperado 'dev' na saida, recebido %q", stdout.String())
	}
}

func TestVersionDevDefault(t *testing.T) {
	if version != "dev" {
		t.Errorf("variavel version deveria ser 'dev' por padrao, recebido %q", version)
	}
}

func TestComandoVersion(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if codigo := executar([]string{"version"}, &stdout, &stderr); codigo != 0 {
		t.Errorf("esperado codigo 0, recebido %d", codigo)
	}
	if !strings.Contains(stdout.String(), "dev") {
		t.Errorf("esperado 'dev' na saida, recebido %q", stdout.String())
	}
}

func TestComandoDesconhecido(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if codigo := executar([]string{"abc"}, &stdout, &stderr); codigo == 0 {
		t.Errorf("esperado codigo nao-zero")
	}
}

func TestHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if codigo := executar([]string{"--help"}, &stdout, &stderr); codigo != 0 {
		t.Errorf("--help deveria retornar 0, recebido %d", codigo)
	}
	if !strings.Contains(stdout.String(), "Uso:") {
		t.Errorf("esperado mensagem de uso")
	}
}

func TestStatusQuandoSimuladorOffline(t *testing.T) {
	var stdout, stderr bytes.Buffer
	codigo := executar([]string{"status", "--porta", "1"}, &stdout, &stderr)
	if codigo != 0 {
		t.Errorf("status offline deveria retornar 0, recebido %d", codigo)
	}
	if !strings.Contains(stdout.String(), "PARADO") && !strings.Contains(stdout.String(), "execucao") {
		t.Errorf("esperado mensagem de status, recebido %q", stdout.String())
	}
}

func TestPararQuandoSimuladorOffline(t *testing.T) {
	var stdout, stderr bytes.Buffer
	codigo := executar([]string{"parar", "--porta", "1"}, &stdout, &stderr)
	if codigo != 0 {
		t.Errorf("parar quando offline deveria retornar 0, recebido %d", codigo)
	}
}
