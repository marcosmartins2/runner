// Package jdk localiza um Java compativel no sistema do usuario e, quando
// ausente, baixa um JRE Temurin para ~/.hubsaude/jre.
package jdk

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/kyriosdata/runner/internal/release"
)

// VersaoMinima exigida pela disciplina (JDK 17+).
const VersaoMinima = 17

// DiretorioHubSaude e o caminho padrao para artefatos provisionados.
func DiretorioHubSaude() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("nao foi possivel determinar HOME: %w", err)
	}
	return filepath.Join(home, ".hubsaude"), nil
}

// DiretorioJRE e o caminho do JRE provisionado dentro de ~/.hubsaude/jre.
func DiretorioJRE() (string, error) {
	base, err := DiretorioHubSaude()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "jre"), nil
}

// CaminhoExecutavelJava devolve <dir>/bin/java[.exe].
func CaminhoExecutavelJava(diretorio string) string {
	nome := "java"
	if runtime.GOOS == "windows" {
		nome = "java.exe"
	}
	return filepath.Join(diretorio, "bin", nome)
}

// LocalizarJava tenta encontrar uma JVM com versao >= VersaoMinima nas seguintes
// fontes, em ordem: JAVA_HOME, PATH do sistema, ~/.hubsaude/jre.
// Retorna o caminho do executavel java ou string vazia.
func LocalizarJava() string {
	candidatos := []string{}

	if javaHome := os.Getenv("JAVA_HOME"); javaHome != "" {
		candidatos = append(candidatos, CaminhoExecutavelJava(javaHome))
	}

	cmdJava := "java"
	if runtime.GOOS == "windows" {
		cmdJava = "java.exe"
	}
	if caminho, err := exec.LookPath(cmdJava); err == nil {
		candidatos = append(candidatos, caminho)
	}

	if dirJRE, err := DiretorioJRE(); err == nil {
		candidatos = append(candidatos, CaminhoExecutavelJava(dirJRE))
	}

	for _, candidato := range candidatos {
		if VersaoCompativel(candidato) {
			return candidato
		}
	}
	return ""
}

// VersaoCompativel executa "java -version" e devolve true quando a major version
// e >= VersaoMinima.
func VersaoCompativel(caminhoJava string) bool {
	v := ExtrairMajor(caminhoJava)
	return v >= VersaoMinima
}

// ExtrairMajor invoca "java -version" e retorna a major version (ex: 17).
// Retorna 0 quando nao consegue extrair.
func ExtrairMajor(caminhoJava string) int {
	if caminhoJava == "" {
		return 0
	}
	ctx, cancelar := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelar()
	cmd := exec.CommandContext(ctx, caminhoJava, "-version")
	bytes, err := cmd.CombinedOutput()
	if err != nil {
		return 0
	}
	return ParsearMajor(string(bytes))
}

// ParsearMajor extrai a major version de uma string produzida por "java -version".
func ParsearMajor(saida string) int {
	re := regexp.MustCompile(`version "(\d+)`)
	m := re.FindStringSubmatch(saida)
	if len(m) < 2 {
		return 0
	}
	major, err := strconv.Atoi(m[1])
	if err != nil {
		return 0
	}
	// "java -version" do Java 8 reporta "1.8". Tratar o caso legado.
	if major == 1 {
		reCompat := regexp.MustCompile(`version "1\.(\d+)`)
		m = reCompat.FindStringSubmatch(saida)
		if len(m) >= 2 {
			if alvo, err := strconv.Atoi(m[1]); err == nil {
				return alvo
			}
		}
	}
	return major
}

// Provisionar garante que ha um Java compativel disponivel.
// Quando ausente, baixa o JRE Temurin a partir do manifesto fornecido.
// Retorna o caminho do executavel java.
func Provisionar(manifesto release.Manifesto) (string, error) {
	if caminho := LocalizarJava(); caminho != "" {
		return caminho, nil
	}

	chave := release.ChaveJREDaPlataforma()
	url := manifesto.Jre[chave]
	if url == "" {
		return "", fmt.Errorf("nao ha JRE pre-configurado para a plataforma %s", chave)
	}

	dirJRE, err := DiretorioJRE()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dirJRE, 0o755); err != nil {
		return "", err
	}

	nomeArq := "jre.tar.gz"
	if runtime.GOOS == "windows" {
		nomeArq = "jre.zip"
	}
	destino := filepath.Join(dirJRE, "download", nomeArq)
	if err := release.BaixarArquivo(url, destino); err != nil {
		return "", fmt.Errorf("falha ao baixar JRE: %w", err)
	}

	if err := Extrair(destino, dirJRE); err != nil {
		return "", fmt.Errorf("falha ao extrair JRE: %w", err)
	}

	// Apos extracao o Temurin gera um subdiretorio (ex: jdk-21.0.2+13-jre).
	// Procurar o primeiro bin/java[.exe] disponivel.
	caminho, err := EncontrarJavaExtraido(dirJRE)
	if err != nil {
		return "", err
	}
	return caminho, nil
}

// EncontrarJavaExtraido localiza recursivamente o primeiro executavel java
// dentro de raiz. Util apos a extracao do JRE Temurin.
func EncontrarJavaExtraido(raiz string) (string, error) {
	alvo := "java"
	if runtime.GOOS == "windows" {
		alvo = "java.exe"
	}
	var encontrado string
	err := filepath.Walk(raiz, func(caminho string, info os.FileInfo, err error) error {
		if err != nil || info == nil || info.IsDir() {
			return nil
		}
		if strings.EqualFold(info.Name(), alvo) && strings.Contains(caminho, string(os.PathSeparator)+"bin"+string(os.PathSeparator)) {
			encontrado = caminho
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if encontrado == "" {
		return "", fmt.Errorf("nao foi possivel localizar java em %s", raiz)
	}
	return encontrado, nil
}
