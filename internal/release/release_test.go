package release

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestCarregarLocal(t *testing.T) {
	dir := t.TempDir()
	caminho := filepath.Join(dir, "release.json")
	conteudo := `{
        "jar": {"url": "https://example.com/x.jar", "version": "1.2.0"},
        "jre": {"linux_x64": "https://example.com/jre-linux.tgz"}
    }`
	if err := os.WriteFile(caminho, []byte(conteudo), 0o644); err != nil {
		t.Fatal(err)
	}

	m, err := CarregarLocal(caminho)
	if err != nil {
		t.Fatalf("CarregarLocal falhou: %v", err)
	}
	if m.Jar.Version != "1.2.0" {
		t.Errorf("versao esperada 1.2.0, recebido %q", m.Jar.Version)
	}
	if m.Jre["linux_x64"] == "" {
		t.Errorf("url do jre linux_x64 ausente")
	}
}

func TestBaixarManifesto(t *testing.T) {
	body := `{"jar":{"url":"u","version":"9.9.9"},"jre":{}}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(body))
	}))
	defer srv.Close()

	m, err := BaixarManifesto(srv.URL)
	if err != nil {
		t.Fatalf("BaixarManifesto falhou: %v", err)
	}
	if m.Jar.Version != "9.9.9" {
		t.Errorf("versao esperada 9.9.9, recebido %q", m.Jar.Version)
	}
}

func TestChaveJREDaPlataforma(t *testing.T) {
	chave := ChaveJREDaPlataforma()
	switch runtime.GOOS {
	case "windows":
		if chave != "windows_x64" {
			t.Errorf("esperado windows_x64, recebido %q", chave)
		}
	case "linux":
		if chave != "linux_x64" {
			t.Errorf("esperado linux_x64, recebido %q", chave)
		}
	case "darwin":
		if chave != "mac_x64" && chave != "mac_aarch64" {
			t.Errorf("esperado mac_x64 ou mac_aarch64, recebido %q", chave)
		}
	}
}

func TestBaixarArquivo(t *testing.T) {
	conteudoBin := []byte("conteudo-binario")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(conteudoBin)
	}))
	defer srv.Close()

	dir := t.TempDir()
	destino := filepath.Join(dir, "subdir", "artefato.bin")
	if err := BaixarArquivo(srv.URL, destino); err != nil {
		t.Fatalf("BaixarArquivo falhou: %v", err)
	}

	bytes, err := os.ReadFile(destino)
	if err != nil {
		t.Fatal(err)
	}
	if string(bytes) != string(conteudoBin) {
		t.Errorf("conteudo divergente: %q", string(bytes))
	}
}

func TestVersaoInstaladaEGravarVersao(t *testing.T) {
	dir := t.TempDir()
	caminho := filepath.Join(dir, "artefato.jar")

	if v := VersaoInstalada(caminho); v != "" {
		t.Errorf("esperado vazio, recebido %q", v)
	}

	if err := GravarVersao(caminho, "1.0.0"); err != nil {
		t.Fatalf("GravarVersao falhou: %v", err)
	}
	if v := VersaoInstalada(caminho); v != "1.0.0" {
		t.Errorf("esperado 1.0.0, recebido %q", v)
	}
}

func TestBaixarManifestoUsaURLPadraoQuandoVazia(t *testing.T) {
	// Apenas confirma que a constante padrao esta preenchida.
	if URLReleaseManifestoPadrao == "" {
		t.Fatal("URL padrao do manifesto nao deve ser vazia")
	}
}
