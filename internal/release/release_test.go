package release

import (
	"crypto/sha256"
	"encoding/hex"
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
        "validador": {"url": "https://example.com/validador.jar", "version": "0.1.8", "tag": "hubsaude-validador-api-v0.1.8", "sha256": "abc"},
        "simulador": {"url": "https://example.com/sim.jar", "version": "0.1.7"},
        "jre": {"linux_x64": "https://example.com/jre-linux.tgz", "mac_arm64": "https://example.com/jre-mac.tgz"}
    }`
	if err := os.WriteFile(caminho, []byte(conteudo), 0o644); err != nil {
		t.Fatal(err)
	}

	m, err := CarregarLocal(caminho)
	if err != nil {
		t.Fatalf("CarregarLocal falhou: %v", err)
	}
	art, ok := m.ArtefatoAssinador()
	if !ok || art.Version != "0.1.8" || art.Tag != "hubsaude-validador-api-v0.1.8" {
		t.Errorf("artefato do assinador inesperado: %+v (ok=%v)", art, ok)
	}
	if art.SHA256 != "abc" {
		t.Errorf("sha256 esperado abc, recebido %q", art.SHA256)
	}
	sim, ok := m.ArtefatoSimulador()
	if !ok || sim.Version != "0.1.7" {
		t.Errorf("artefato do simulador inesperado: %+v (ok=%v)", sim, ok)
	}
	if m.Jre["mac_arm64"] == "" {
		t.Errorf("url do jre mac_arm64 ausente")
	}
}

func TestArtefatoAssinadorCompatibilidadeJar(t *testing.T) {
	// Manifesto antigo (sem "validador", apenas "jar") deve continuar funcionando.
	m := Manifesto{Jar: Artefato{URL: "https://example.com/x.jar", Version: "1.2.0"}}
	art, ok := m.ArtefatoAssinador()
	if !ok || art.Version != "1.2.0" {
		t.Errorf("esperado fallback para jar 1.2.0, recebido %+v (ok=%v)", art, ok)
	}
}

func TestBaixarManifesto(t *testing.T) {
	body := `{"validador":{"url":"u","version":"9.9.9"},"jre":{}}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(body))
	}))
	defer srv.Close()

	m, err := BaixarManifesto(srv.URL)
	if err != nil {
		t.Fatalf("BaixarManifesto falhou: %v", err)
	}
	art, ok := m.ArtefatoAssinador()
	if !ok || art.Version != "9.9.9" {
		t.Errorf("versao esperada 9.9.9, recebido %+v (ok=%v)", art, ok)
	}
}

func TestChaveJREDaPlataforma(t *testing.T) {
	chave := ChaveJREDaPlataforma()
	arch := "x64"
	if runtime.GOARCH == "arm64" {
		arch = "arm64"
	}
	var esperado string
	switch runtime.GOOS {
	case "windows":
		esperado = "windows_" + arch
	case "linux":
		esperado = "linux_" + arch
	case "darwin":
		esperado = "mac_" + arch
	default:
		return
	}
	if chave != esperado {
		t.Errorf("esperado %q, recebido %q", esperado, chave)
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

	dados, err := os.ReadFile(destino)
	if err != nil {
		t.Fatal(err)
	}
	if string(dados) != string(conteudoBin) {
		t.Errorf("conteudo divergente: %q", string(dados))
	}
}

func TestBaixarArquivoVerificadoSHA256OK(t *testing.T) {
	conteudoBin := []byte("artefato-integro")
	soma := sha256.Sum256(conteudoBin)
	hexSoma := hex.EncodeToString(soma[:])

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(conteudoBin)
	}))
	defer srv.Close()

	dir := t.TempDir()
	destino := filepath.Join(dir, "artefato.bin")
	if err := BaixarArquivoVerificado(srv.URL, destino, hexSoma); err != nil {
		t.Fatalf("verificacao deveria passar: %v", err)
	}
	if _, err := os.Stat(destino); err != nil {
		t.Errorf("arquivo deveria existir apos download verificado: %v", err)
	}
}

func TestBaixarArquivoVerificadoSHA256Divergente(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("conteudo-adulterado"))
	}))
	defer srv.Close()

	dir := t.TempDir()
	destino := filepath.Join(dir, "artefato.bin")
	shaErrado := "0000000000000000000000000000000000000000000000000000000000000000"
	err := BaixarArquivoVerificado(srv.URL, destino, shaErrado)
	if err == nil {
		t.Fatal("esperado erro de integridade")
	}
	if _, statErr := os.Stat(destino); !os.IsNotExist(statErr) {
		t.Errorf("arquivo corrompido nao deveria ter sido mantido")
	}
	if _, statErr := os.Stat(destino + ".part"); !os.IsNotExist(statErr) {
		t.Errorf("arquivo .part deveria ter sido removido")
	}
}

func TestVerificarSHA256(t *testing.T) {
	dir := t.TempDir()
	caminho := filepath.Join(dir, "x.bin")
	conteudo := []byte("abc")
	if err := os.WriteFile(caminho, conteudo, 0o644); err != nil {
		t.Fatal(err)
	}
	soma := sha256.Sum256(conteudo)
	hexSoma := hex.EncodeToString(soma[:])

	if err := VerificarSHA256(caminho, hexSoma); err != nil {
		t.Errorf("digest correto deveria passar: %v", err)
	}
	// Comparacao deve ser insensivel a maiusculas/minusculas.
	if err := VerificarSHA256(caminho, "DEAD"); err == nil {
		t.Errorf("digest incorreto deveria falhar")
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
