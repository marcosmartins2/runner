// Package release lida com o manifesto release.json publicado pelo repositorio
// da disciplina e com o download de artefatos (assinador/validador, simulador.jar
// e JRE), incluindo verificacao de integridade via SHA256.
package release

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// URLReleaseManifestoPadrao aponta para o manifesto estavel publicado na branch
// main do repositorio da disciplina (upstream). E a unica fonte da verdade dos
// artefatos reais consumidos em runtime (ver docs/adr/0004-estrategia-hibrida-artefatos.md).
const URLReleaseManifestoPadrao = "https://raw.githubusercontent.com/kyriosdata/runner/main/release.json"

// Manifesto representa a estrutura do release.json.
//
// O campo "validador" corresponde ao assinador.jar real publicado pelo upstream
// (hubsaude-validador-api). "jar" e mantido apenas por compatibilidade com
// manifestos antigos. "simulador" descreve o simulador.jar (hubsaude-simulador).
type Manifesto struct {
	Validador *Artefato         `json:"validador,omitempty"`
	Jar       Artefato          `json:"jar,omitempty"`
	Sim       *Artefato         `json:"simulador,omitempty"`
	Jre       map[string]string `json:"jre"`
}

// Artefato descreve uma entrada do manifesto (validador, simulador ou jar).
type Artefato struct {
	URL     string `json:"url"`
	Version string `json:"version"`
	Tag     string `json:"tag,omitempty"`
	SHA256  string `json:"sha256,omitempty"`
}

// ArtefatoAssinador devolve o artefato do assinador a ser consumido, preferindo
// "validador" (artefato real do upstream) e caindo para "jar" (compatibilidade).
// O segundo retorno e false quando nenhum esta disponivel.
func (m Manifesto) ArtefatoAssinador() (Artefato, bool) {
	if m.Validador != nil && m.Validador.URL != "" {
		return *m.Validador, true
	}
	if m.Jar.URL != "" {
		return m.Jar, true
	}
	return Artefato{}, false
}

// ArtefatoSimulador devolve o artefato do simulador, quando presente.
func (m Manifesto) ArtefatoSimulador() (Artefato, bool) {
	if m.Sim != nil && m.Sim.URL != "" {
		return *m.Sim, true
	}
	return Artefato{}, false
}

// CarregarLocal le um release.json a partir do caminho informado.
func CarregarLocal(caminho string) (Manifesto, error) {
	var m Manifesto
	conteudo, err := os.ReadFile(caminho)
	if err != nil {
		return m, fmt.Errorf("falha ao ler release local: %w", err)
	}
	if err := json.Unmarshal(conteudo, &m); err != nil {
		return m, fmt.Errorf("release.json invalido: %w", err)
	}
	return m, nil
}

// BaixarManifesto recupera o release.json a partir da URL fornecida (ou da
// URLReleaseManifestoPadrao se a string for vazia).
func BaixarManifesto(url string) (Manifesto, error) {
	if url == "" {
		url = URLReleaseManifestoPadrao
	}
	var m Manifesto
	cliente := &http.Client{Timeout: 30 * time.Second}
	resp, err := cliente.Get(url)
	if err != nil {
		return m, fmt.Errorf("falha ao baixar manifesto: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return m, fmt.Errorf("manifesto indisponivel: HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return m, fmt.Errorf("falha ao ler manifesto: %w", err)
	}
	if err := json.Unmarshal(body, &m); err != nil {
		return m, fmt.Errorf("manifesto invalido: %w", err)
	}
	return m, nil
}

// ChaveJREDaPlataforma retorna a chave do mapa "jre" para o GOOS/GOARCH atual.
// As chaves seguem a convencao do manifesto upstream (ex.: windows_x64,
// linux_arm64, mac_arm64).
func ChaveJREDaPlataforma() string {
	arch := "x64"
	if runtime.GOARCH == "arm64" {
		arch = "arm64"
	}
	switch runtime.GOOS {
	case "windows":
		return "windows_" + arch
	case "linux":
		return "linux_" + arch
	case "darwin":
		return "mac_" + arch
	}
	return runtime.GOOS + "_" + runtime.GOARCH
}

// BaixarArquivo grava o conteudo de url em destino, criando os diretorios
// intermediarios. Retorna erro se o status HTTP nao for 2xx.
func BaixarArquivo(url, destino string) error {
	return BaixarArquivoVerificado(url, destino, "")
}

// BaixarArquivoVerificado baixa url para destino e, quando sha256Esperado nao
// for vazio, confere o digest do arquivo baixado. Em caso de divergencia o
// arquivo e removido e um erro explicito e retornado (protecao de supply chain).
func BaixarArquivoVerificado(url, destino, sha256Esperado string) error {
	if err := os.MkdirAll(filepath.Dir(destino), 0o755); err != nil {
		return fmt.Errorf("falha ao criar diretorio: %w", err)
	}

	cliente := &http.Client{Timeout: 5 * time.Minute}
	resp, err := cliente.Get(url)
	if err != nil {
		return fmt.Errorf("falha ao baixar %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("download %s retornou HTTP %d", url, resp.StatusCode)
	}

	temp := destino + ".part"
	out, err := os.Create(temp)
	if err != nil {
		return fmt.Errorf("falha ao criar arquivo: %w", err)
	}
	if _, err := io.Copy(out, resp.Body); err != nil {
		out.Close()
		os.Remove(temp)
		return fmt.Errorf("falha ao gravar arquivo: %w", err)
	}
	if err := out.Close(); err != nil {
		os.Remove(temp)
		return fmt.Errorf("falha ao fechar arquivo: %w", err)
	}

	if sha256Esperado != "" {
		if err := VerificarSHA256(temp, sha256Esperado); err != nil {
			os.Remove(temp)
			return err
		}
	}

	if err := os.Rename(temp, destino); err != nil {
		os.Remove(temp)
		return fmt.Errorf("falha ao mover arquivo: %w", err)
	}
	return nil
}

// VerificarSHA256 calcula o digest SHA256 do arquivo em caminho e o compara,
// sem diferenciar maiusculas/minusculas, com o valor esperado.
func VerificarSHA256(caminho, esperado string) error {
	f, err := os.Open(caminho)
	if err != nil {
		return fmt.Errorf("falha ao abrir arquivo para verificacao: %w", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return fmt.Errorf("falha ao calcular sha256: %w", err)
	}
	obtido := hex.EncodeToString(h.Sum(nil))
	if !strings.EqualFold(obtido, esperado) {
		return fmt.Errorf("integridade comprometida: sha256 esperado %s, obtido %s", esperado, obtido)
	}
	return nil
}

// VersaoInstalada le um arquivo "<artefato>.version" lado-a-lado com o artefato.
// Retorna string vazia se o arquivo nao existir.
func VersaoInstalada(caminhoArtefato string) string {
	versao, err := os.ReadFile(caminhoArtefato + ".version")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(versao))
}

// GravarVersao registra a versao instalada do artefato.
func GravarVersao(caminhoArtefato, versao string) error {
	return os.WriteFile(caminhoArtefato+".version", []byte(versao), 0o644)
}
