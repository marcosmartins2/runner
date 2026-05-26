package jdk

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestParsearMajor(t *testing.T) {
	testes := []struct {
		nome     string
		entrada  string
		esperado int
	}{
		{"java 17", `openjdk version "17.0.10" 2024-01-16`, 17},
		{"java 21", `openjdk version "21.0.2" 2024-01-16`, 21},
		{"java 8 legado", `java version "1.8.0_392"`, 8},
		{"sem versao", "outra coisa qualquer", 0},
	}
	for _, tt := range testes {
		t.Run(tt.nome, func(t *testing.T) {
			if got := ParsearMajor(tt.entrada); got != tt.esperado {
				t.Errorf("ParsearMajor(%q) = %d, esperado %d", tt.entrada, got, tt.esperado)
			}
		})
	}
}

func TestCaminhoExecutavelJava(t *testing.T) {
	caminho := CaminhoExecutavelJava("/opt/jdk")
	if runtime.GOOS == "windows" {
		if caminho != filepath.Join("/opt/jdk", "bin", "java.exe") {
			t.Errorf("caminho windows incorreto: %s", caminho)
		}
	} else {
		if caminho != filepath.Join("/opt/jdk", "bin", "java") {
			t.Errorf("caminho unix incorreto: %s", caminho)
		}
	}
}

func TestDiretorioHubSaude(t *testing.T) {
	dir, err := DiretorioHubSaude()
	if err != nil {
		t.Fatal(err)
	}
	if !filepath.IsAbs(dir) {
		t.Errorf("esperado caminho absoluto, recebido %q", dir)
	}
	if filepath.Base(dir) != ".hubsaude" {
		t.Errorf("esperado .hubsaude, recebido %q", filepath.Base(dir))
	}
}

func TestExtrairZip(t *testing.T) {
	dirArq := t.TempDir()
	dirDestino := t.TempDir()
	zipPath := filepath.Join(dirArq, "exemplo.zip")

	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	w := zip.NewWriter(f)
	entry, err := w.Create("subdir/arquivo.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := entry.Write([]byte("conteudo")); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	f.Close()

	if err := Extrair(zipPath, dirDestino); err != nil {
		t.Fatalf("Extrair falhou: %v", err)
	}

	bytes, err := os.ReadFile(filepath.Join(dirDestino, "subdir", "arquivo.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(bytes) != "conteudo" {
		t.Errorf("conteudo divergente: %q", string(bytes))
	}
}

func TestExtrairTarGz(t *testing.T) {
	dirArq := t.TempDir()
	dirDestino := t.TempDir()
	tgz := filepath.Join(dirArq, "exemplo.tar.gz")

	f, err := os.Create(tgz)
	if err != nil {
		t.Fatal(err)
	}
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)
	dados := []byte("conteudo-tar")
	hdr := &tar.Header{Name: "subdir/arq.txt", Mode: 0o644, Size: int64(len(dados)), Typeflag: tar.TypeReg}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(dados); err != nil {
		t.Fatal(err)
	}
	tw.Close()
	gz.Close()
	f.Close()

	if err := Extrair(tgz, dirDestino); err != nil {
		t.Fatalf("Extrair tar.gz falhou: %v", err)
	}
	bytes, err := os.ReadFile(filepath.Join(dirDestino, "subdir", "arq.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(bytes) != string(dados) {
		t.Errorf("conteudo divergente: %q", string(bytes))
	}
}

func TestLocalizarJavaUsaJavaHome(t *testing.T) {
	if testing.Short() {
		t.Skip("short")
	}
	// Sem java instalado no ambiente, LocalizarJava deve ao menos nao panicar.
	_ = LocalizarJava()
}
