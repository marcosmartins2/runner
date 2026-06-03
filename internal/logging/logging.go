// Package logging oferece logs estruturados (via log/slog) para as CLIs, com
// nivel ajustavel por --verbose / --quiet. Diagnosticos vao sempre para stderr,
// mantendo stdout reservado ao resultado da operacao (separacao exigida pela
// especificacao). Ver docs/criterios.md, secoes D e I.
package logging

import (
	"io"
	"log/slog"
)

// Nivel de verbosidade selecionado pela linha de comando.
type Nivel int

const (
	// Normal exibe info, avisos e erros.
	Normal Nivel = iota
	// Verbose (--verbose) inclui mensagens de depuracao.
	Verbose
	// Quiet (--quiet) suprime tudo abaixo de erro.
	Quiet
)

// atual e o logger global; por padrao descarta a saida ate Configurar ser
// chamado pela CLI.
var atual = slog.New(slog.NewTextHandler(io.Discard, nil))

// Configurar define o destino (em geral stderr) e o nivel do logger global.
func Configurar(nivel Nivel, w io.Writer) {
	var lvl slog.Level
	switch nivel {
	case Verbose:
		lvl = slog.LevelDebug
	case Quiet:
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}
	atual = slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{Level: lvl}))
}

// Logger devolve o logger configurado.
func Logger() *slog.Logger { return atual }

// Debug registra uma mensagem de depuracao (visivel apenas em --verbose).
func Debug(msg string, args ...any) { atual.Debug(msg, args...) }

// Info registra uma mensagem informativa.
func Info(msg string, args ...any) { atual.Info(msg, args...) }

// Warn registra um aviso.
func Warn(msg string, args ...any) { atual.Warn(msg, args...) }

// Error registra um erro.
func Error(msg string, args ...any) { atual.Error(msg, args...) }
