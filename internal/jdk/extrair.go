package jdk

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Extrair descompacta um .zip ou .tar.gz em destino, criando os diretorios
// necessarios. Detecta o formato pela extensao do nome do arquivo.
func Extrair(arquivo, destino string) error {
	nome := strings.ToLower(arquivo)
	switch {
	case strings.HasSuffix(nome, ".zip"):
		return extrairZip(arquivo, destino)
	case strings.HasSuffix(nome, ".tar.gz"), strings.HasSuffix(nome, ".tgz"):
		return extrairTarGz(arquivo, destino)
	}
	return fmt.Errorf("formato nao suportado: %s", arquivo)
}

func extrairZip(arquivo, destino string) error {
	r, err := zip.OpenReader(arquivo)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		caminho := filepath.Join(destino, f.Name)
		if !strings.HasPrefix(caminho, filepath.Clean(destino)+string(os.PathSeparator)) {
			return fmt.Errorf("entrada zip fora do destino: %s", f.Name)
		}
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(caminho, f.Mode()); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(caminho), 0o755); err != nil {
			return err
		}
		if err := copiarZipEntry(f, caminho); err != nil {
			return err
		}
	}
	return nil
}

func copiarZipEntry(f *zip.File, destino string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()
	w, err := os.OpenFile(destino, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, f.Mode())
	if err != nil {
		return err
	}
	defer w.Close()
	_, err = io.Copy(w, rc)
	return err
}

func extrairTarGz(arquivo, destino string) error {
	f, err := os.Open(arquivo)
	if err != nil {
		return err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()

	r := tar.NewReader(gz)
	for {
		header, err := r.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		caminho := filepath.Join(destino, header.Name)
		if !strings.HasPrefix(caminho, filepath.Clean(destino)+string(os.PathSeparator)) {
			return fmt.Errorf("entrada tar fora do destino: %s", header.Name)
		}
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(caminho, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(caminho), 0o755); err != nil {
				return err
			}
			out, err := os.OpenFile(caminho, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(out, r); err != nil {
				out.Close()
				return err
			}
			out.Close()
		case tar.TypeSymlink:
			os.Symlink(header.Linkname, caminho)
		}
	}
	return nil
}
