package invoker

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// Cenarios negativos e de contrato exigidos pelo criterio G de docs/criterios.md:
// JVM ausente, porta ocupada, reuso de instancia viva, conexao recusada,
// resposta malformada e contrato real via subprocess.

func TestInvocacaoLocalJavaAusente(t *testing.T) {
	dir := t.TempDir()
	jar := filepath.Join(dir, "assinador.jar")
	if err := os.WriteFile(jar, []byte("conteudo"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := InvocacaoLocal("", jar, []string{"--operacao", "criar"})
	if err == nil || !strings.Contains(err.Error(), "java nao encontrado") {
		t.Errorf("esperado erro de java ausente, recebido %v", err)
	}
}

func TestIniciarServidorPortaOcupada(t *testing.T) {
	// Servidor que NAO responde /api/info como Assinador, apenas ocupa a porta.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()
	porta := portaDeURL(t, srv.URL)

	dir := t.TempDir()
	estado := filepath.Join(dir, "estado.json")
	err := IniciarServidor("java", filepath.Join(dir, "x.jar"), porta, 0, estado)
	if err == nil || !strings.Contains(err.Error(), "ja em uso") {
		t.Errorf("esperado erro de porta ocupada, recebido %v", err)
	}
}

func TestIniciarServidorReutilizaInstanciaViva(t *testing.T) {
	// Idempotencia de start (criterio E2): instancia viva deve ser reutilizada
	// sem tentar iniciar processo novo (java/jar propositalmente invalidos).
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/info" {
			json.NewEncoder(w).Encode(Resposta{Status: "sucesso"})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()
	porta := portaDeURL(t, srv.URL)

	dir := t.TempDir()
	estado := filepath.Join(dir, "estado.json")
	if err := IniciarServidor("/caminho/inexistente/java", filepath.Join(dir, "x.jar"), porta, 0, estado); err != nil {
		t.Errorf("instancia viva deveria ser reutilizada, recebido erro: %v", err)
	}
}

func TestRequisicaoJSONConexaoRecusada(t *testing.T) {
	// Reserva e libera uma porta para garantir que nao ha servidor escutando.
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	porta := l.Addr().(*net.TCPAddr).Port
	l.Close()

	_, err = RequisicaoJSON(http.MethodGet, urlServidor(porta, "/api/info"), nil)
	if err == nil || !strings.Contains(err.Error(), "falha ao conectar") {
		t.Errorf("esperado erro de conexao recusada, recebido %v", err)
	}
}

func TestRequisicaoJSONRespostaMalformada(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("isto nao e json"))
	}))
	defer srv.Close()
	porta := portaDeURL(t, srv.URL)

	_, err := RequisicaoJSON(http.MethodGet, urlServidor(porta, "/api/info"), nil)
	if err == nil || !strings.Contains(err.Error(), "JSON valido") {
		t.Errorf("esperado erro de resposta malformada, recebido %v", err)
	}
}

// TestInvocacaoLocalSubprocessReal exercita o contrato CLI<->JAR via subprocess
// real: um script faz o papel do "java", recebe argumentos e devolve JSON em
// stdout. Em Windows e ignorado (usa shell script POSIX); o contrato em Windows
// fica coberto pelos testes HTTP, que rodam nas tres plataformas.
func TestInvocacaoLocalSubprocessReal(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("subprocess via shell script POSIX; coberto por testes HTTP no Windows")
	}
	dir := t.TempDir()
	jar := filepath.Join(dir, "assinador.jar")
	if err := os.WriteFile(jar, []byte("fake"), 0o644); err != nil {
		t.Fatal(err)
	}

	fakeJava := filepath.Join(dir, "java")
	script := "#!/bin/sh\n" +
		"cat <<'JSON'\n" +
		`{"status":"sucesso","assinatura":"ABC123","algoritmo":"SHA256"}` + "\n" +
		"JSON\n"
	if err := os.WriteFile(fakeJava, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	resp, err := InvocacaoLocal(fakeJava, jar, []string{"--operacao", "criar", "--documento", "SGVsbG8="})
	if err != nil {
		t.Fatalf("InvocacaoLocal (subprocess real) falhou: %v", err)
	}
	if resp.Status != "sucesso" || resp.Assinatura != "ABC123" {
		t.Errorf("resposta inesperada do subprocess: %+v", resp)
	}
}
