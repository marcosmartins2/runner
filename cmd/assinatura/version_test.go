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

// TestVersionViaGoRun executa "go run . version" e confere se a saida contem
// o valor padrao da variavel version ("dev"), conforme T-01.1.6.
func TestVersionViaGoRun(t *testing.T) {
	if testing.Short() {
		t.Skip("modo curto: pula teste de aceitacao com go run")
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("comando go indisponivel; saltando teste de aceitacao")
	}

	_, arquivoCorrente, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("nao foi possivel determinar diretorio do teste")
	}
	dir := filepath.Dir(arquivoCorrente)

	cmd := exec.Command("go", "run", ".", "version")
	cmd.Dir = dir
	cmd.Env = os.Environ()

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("go run falhou: %v\nstdout=%s\nstderr=%s", err, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "dev") {
		t.Errorf("esperado 'dev' na saida, recebido: %q", stdout.String())
	}
}

func TestVersionDevDefault(t *testing.T) {
	if version != "dev" {
		t.Errorf("variavel version deveria ser 'dev' por padrao, recebido %q", version)
	}
}

func TestVersionFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if codigo := executar([]string{"--version"}, &stdout, &stderr); codigo != 0 {
		t.Errorf("esperado codigo 0, recebido %d", codigo)
	}
	if !strings.Contains(stdout.String(), "dev") {
		t.Errorf("esperado 'dev' na saida, recebido: %q", stdout.String())
	}
}

func TestComandoVersion(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if codigo := executar([]string{"version"}, &stdout, &stderr); codigo != 0 {
		t.Errorf("esperado codigo 0, recebido %d", codigo)
	}
	if !strings.Contains(stdout.String(), "dev") {
		t.Errorf("esperado 'dev' na saida, recebido: %q", stdout.String())
	}
}

func TestComandoDesconhecidoExibeUso(t *testing.T) {
	var stdout, stderr bytes.Buffer
	codigo := executar([]string{"xyz"}, &stdout, &stderr)
	if codigo == 0 {
		t.Errorf("esperado codigo nao-zero para comando desconhecido")
	}
	if !strings.Contains(stderr.String(), "desconhecido") {
		t.Errorf("esperado mensagem de comando desconhecido em stderr, recebido: %q", stderr.String())
	}
}

func TestSemArgumentosExibeUso(t *testing.T) {
	var stdout, stderr bytes.Buffer
	codigo := executar([]string{}, &stdout, &stderr)
	if codigo != 1 {
		t.Errorf("esperado codigo 1 para chamada sem argumentos, recebido %d", codigo)
	}
}

func TestHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if codigo := executar([]string{"--help"}, &stdout, &stderr); codigo != 0 {
		t.Errorf("--help deveria retornar 0, recebido %d", codigo)
	}
	if !strings.Contains(stdout.String(), "Uso:") {
		t.Errorf("esperado mensagem de uso, recebido: %q", stdout.String())
	}
}

func TestCriarSemParametrosObrigatoriosFalha(t *testing.T) {
	var stdout, stderr bytes.Buffer
	codigo := executar([]string{"criar"}, &stdout, &stderr)
	if codigo == 0 {
		t.Errorf("esperado codigo nao-zero quando faltam parametros")
	}
}

func TestValidarSemParametrosObrigatoriosFalha(t *testing.T) {
	var stdout, stderr bytes.Buffer
	codigo := executar([]string{"validar"}, &stdout, &stderr)
	if codigo == 0 {
		t.Errorf("esperado codigo nao-zero quando faltam parametros")
	}
}
