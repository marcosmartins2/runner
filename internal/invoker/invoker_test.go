package invoker

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestPortaDisponivel(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	porta := l.Addr().(*net.TCPAddr).Port
	if PortaDisponivel(porta) {
		t.Errorf("porta %d deveria estar ocupada", porta)
	}
	l.Close()
	if !PortaDisponivel(porta) {
		t.Errorf("porta %d deveria estar livre apos close", porta)
	}
}

func TestGravarLerRemoverEstado(t *testing.T) {
	dir := t.TempDir()
	caminho := filepath.Join(dir, "estado.json")

	estado := EstadoServidor{PID: 4242, Porta: 8080, Jar: "/tmp/x.jar", IniciadoEm: "2026-05-26T10:00:00Z"}
	if err := GravarEstado(caminho, estado); err != nil {
		t.Fatal(err)
	}

	lido, err := LerEstado(caminho)
	if err != nil {
		t.Fatal(err)
	}
	if lido != estado {
		t.Errorf("estado divergente: %+v", lido)
	}

	if err := RemoverEstado(caminho); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(caminho); !os.IsNotExist(err) {
		t.Errorf("arquivo deveria ter sido removido: %v", err)
	}

	vazio, err := LerEstado(caminho)
	if err != nil {
		t.Fatal(err)
	}
	if vazio != (EstadoServidor{}) {
		t.Errorf("esperado estado zero, recebido %+v", vazio)
	}
}

func TestEmExecucaoTrue(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(Resposta{Status: "sucesso"})
	}))
	defer srv.Close()

	porta := portaDeURL(t, srv.URL)
	if !EmExecucao(porta) {
		t.Errorf("EmExecucao deveria retornar true")
	}
}

func TestEmExecucaoFalse(t *testing.T) {
	// Porta de 0 e invalida — sem listener
	if EmExecucao(1) {
		t.Errorf("EmExecucao deveria retornar false")
	}
}

func TestInvocacaoHTTPSign(t *testing.T) {
	var recebido map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/sign" {
			t.Errorf("metodo/path inesperado: %s %s", r.Method, r.URL.Path)
		}
		json.NewDecoder(r.Body).Decode(&recebido)
		json.NewEncoder(w).Encode(Resposta{Status: "sucesso", Assinatura: "x"})
	}))
	defer srv.Close()

	porta := portaDeURL(t, srv.URL)
	resp, err := InvocacaoHTTPSign(porta, "SGVsbG8=", "cert-001")
	if err != nil {
		t.Fatalf("InvocacaoHTTPSign falhou: %v", err)
	}
	if resp.Status != "sucesso" || resp.Assinatura != "x" {
		t.Errorf("resposta inesperada: %+v", resp)
	}
	if recebido["documento"] != "SGVsbG8=" || recebido["certificado"] != "cert-001" {
		t.Errorf("corpo inesperado: %+v", recebido)
	}
}

func TestInvocacaoHTTPValidate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/validate" {
			t.Errorf("path inesperado: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(Resposta{Status: "sucesso"})
	}))
	defer srv.Close()

	porta := portaDeURL(t, srv.URL)
	if _, err := InvocacaoHTTPValidate(porta, "SGVsbG8=", "dGVzdA=="); err != nil {
		t.Fatalf("InvocacaoHTTPValidate falhou: %v", err)
	}
}

func TestPararServidor(t *testing.T) {
	dir := t.TempDir()
	estado := filepath.Join(dir, "estado.json")
	_ = GravarEstado(estado, EstadoServidor{PID: 1, Porta: 8080, Jar: "x"})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/shutdown" {
			t.Errorf("path inesperado: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(Resposta{Status: "sucesso", Mensagem: "encerrando"})
	}))
	defer srv.Close()

	porta := portaDeURL(t, srv.URL)
	resp, err := PararServidor(porta, estado)
	if err != nil {
		t.Fatalf("PararServidor falhou: %v", err)
	}
	if resp.Status != "sucesso" {
		t.Errorf("status inesperado: %s", resp.Status)
	}
	if _, err := os.Stat(estado); !os.IsNotExist(err) {
		t.Errorf("estado deveria ter sido removido")
	}
}

func TestInvocacaoLocalJarAusente(t *testing.T) {
	_, err := InvocacaoLocal("java", "/jar/inexistente.jar", []string{"--operacao", "criar"})
	if err == nil || !strings.Contains(err.Error(), "nao encontrado") {
		t.Errorf("esperado erro de jar nao encontrado, recebido %v", err)
	}
}

func TestRequisicaoJSONErroHTTP(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(Resposta{Status: "falha", Mensagem: "parametro invalido"})
	}))
	defer srv.Close()
	porta := portaDeURL(t, srv.URL)
	_, err := RequisicaoJSON(http.MethodPost, urlServidor(porta, "/api/sign"), map[string]string{"x": "y"})
	if err == nil || !strings.Contains(err.Error(), "parametro invalido") {
		t.Errorf("erro inesperado: %v", err)
	}
}

func portaDeURL(t *testing.T, u string) int {
	t.Helper()
	// httptest URL no formato http://127.0.0.1:PORT
	parte := u[strings.LastIndex(u, ":")+1:]
	p, err := strconv.Atoi(parte)
	if err != nil {
		t.Fatalf("nao foi possivel extrair porta de %s: %v", u, err)
	}
	return p
}
