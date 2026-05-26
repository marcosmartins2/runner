// Package release lida com o manifesto release.json publicado pelo repositorio
// da disciplina e com o download de artefatos (simulador.jar e JRE).
package release

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// URLReleaseManifestoPadrao aponta para o manifesto estavel publicado na branch
// main do repositorio da disciplina.
const URLReleaseManifestoPadrao = "https://raw.githubusercontent.com/kyriosdata/runner/main/release.json"

// Manifesto representa a estrutura do release.json.
type Manifesto struct {
	Jar Artefato            `json:"jar"`
	Jre map[string]string   `json:"jre"`
	Sim *Artefato           `json:"simulador,omitempty"`
}

// Artefato descreve uma entrada (jar ou jre).
type Artefato struct {
	URL     string `json:"url"`
	Version string `json:"version"`
}

// CarregarLocal le um release.json a partir do caminho informado.
func CarregarLocal(caminho string) (Manifesto, error) {
	var m Manifesto
	bytes, err := os.ReadFile(caminho)
	if err != nil {
		return m, fmt.Errorf("falha ao ler release local: %w", err)
	}
	if err := json.Unmarshal(bytes, &m); err != nil {
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
func ChaveJREDaPlataforma() string {
	switch runtime.GOOS {
	case "windows":
		return "windows_x64"
	case "linux":
		return "linux_x64"
	case "darwin":
		if runtime.GOARCH == "arm64" {
			return "mac_aarch64"
		}
		return "mac_x64"
	}
	return runtime.GOOS + "_" + runtime.GOARCH
}

// BaixarArquivo grava o conteudo de url em destino, criando os diretorios
// intermediarios. Retorna erro se o status HTTP nao for 2xx.
func BaixarArquivo(url, destino string) error {
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
	if err := os.Rename(temp, destino); err != nil {
		os.Remove(temp)
		return fmt.Errorf("falha ao mover arquivo: %w", err)
	}
	return nil
}

// VersaoInstalada le um arquivo "VERSION" lado-a-lado com o artefato.
// Retorna string vazia se o arquivo nao existir.
func VersaoInstalada(caminhoArtefato string) string {
	versao, err := os.ReadFile(caminhoArtefato + ".version")
	if err != nil {
		return ""
	}
	return string(versao)
}

// GravarVersao registra a versao instalada do artefato.
func GravarVersao(caminhoArtefato, versao string) error {
	return os.WriteFile(caminhoArtefato+".version", []byte(versao), 0o644)
}
