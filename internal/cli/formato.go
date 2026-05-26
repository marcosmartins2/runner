// Package cli oferece formatadores compartilhados entre os comandos
// "assinatura" e "simulador".
package cli

import (
	"fmt"
	"strings"

	"github.com/kyriosdata/runner/internal/invoker"
)

const linha = "============================================================"
const meia = "------------------------------------------------------------"

// FormatarResposta devolve uma representacao tabular da Resposta para o terminal.
func FormatarResposta(r invoker.Resposta) string {
	var b strings.Builder
	b.WriteString(linha)
	b.WriteString("\n")

	icone := "[OK]"
	status := strings.ToUpper(r.Status)
	if status == "" {
		status = "DESCONHECIDO"
	}
	if r.Status != "sucesso" {
		icone = "[ERRO]"
	}
	fmt.Fprintf(&b, "  %s Status: %s\n", icone, status)
	b.WriteString(meia)
	b.WriteString("\n")

	if r.Mensagem != "" {
		fmt.Fprintf(&b, "  Mensagem:    %s\n", r.Mensagem)
	}
	if r.Algoritmo != "" {
		fmt.Fprintf(&b, "  Algoritmo:   %s\n", r.Algoritmo)
	}
	if r.Porta != 0 {
		fmt.Fprintf(&b, "  Porta:       %d\n", r.Porta)
	}
	if r.TimeoutInatividadeMinutos != 0 {
		fmt.Fprintf(&b, "  Timeout:     %d min\n", r.TimeoutInatividadeMinutos)
	}
	if r.Certificado != "" {
		fmt.Fprintf(&b, "  Certificado: %s\n", r.Certificado)
	}
	if r.Assinatura != "" {
		assinatura := r.Assinatura
		if len(assinatura) > 50 {
			assinatura = assinatura[:50] + "..."
		}
		fmt.Fprintf(&b, "  Assinatura:  %s\n", assinatura)
	}
	if r.Timestamp != "" {
		fmt.Fprintf(&b, "  Timestamp:   %s\n", r.Timestamp)
	}
	b.WriteString(linha)
	return b.String()
}

// FormatarStatusSimulador imprime a resposta do Simulador com tom "informativo"
// quando o status nao for "sucesso".
func FormatarStatusSimulador(r invoker.Resposta) string {
	var b strings.Builder
	b.WriteString(linha)
	b.WriteString("\n")

	estado := strings.ToUpper(r.Status)
	if estado == "" {
		estado = "DESCONHECIDO"
	}
	icone := "[INFO]"
	if r.Status == "sucesso" || r.Status == "ok" || r.Status == "up" || r.Status == "running" {
		icone = "[OK]"
	}
	fmt.Fprintf(&b, "  %s Status: %s\n", icone, estado)
	b.WriteString(meia)
	b.WriteString("\n")
	if r.Mensagem != "" {
		fmt.Fprintf(&b, "  Mensagem:    %s\n", r.Mensagem)
	}
	if r.Porta != 0 {
		fmt.Fprintf(&b, "  Porta:       %d\n", r.Porta)
	}
	if r.Timestamp != "" {
		fmt.Fprintf(&b, "  Timestamp:   %s\n", r.Timestamp)
	}
	b.WriteString(linha)
	return b.String()
}
