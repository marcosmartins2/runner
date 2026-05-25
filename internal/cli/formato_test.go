package cli

import (
	"strings"
	"testing"

	"github.com/kyriosdata/runner/internal/invoker"
)

func TestFormatarRespostaSucesso(t *testing.T) {
	r := invoker.Resposta{
		Status:      "sucesso",
		Mensagem:    "Assinatura criada.",
		Algoritmo:   "SHA256withRSA",
		Certificado: "cert-001",
		Assinatura:  "YWJj",
		Timestamp:   "2026-05-26T10:00:00Z",
	}
	saida := FormatarResposta(r)
	if !strings.Contains(saida, "[OK]") {
		t.Error("esperado icone [OK]")
	}
	if !strings.Contains(saida, "SUCESSO") {
		t.Error("esperado SUCESSO em maiusculas")
	}
	if !strings.Contains(saida, "cert-001") {
		t.Error("esperado certificado")
	}
}

func TestFormatarRespostaFalha(t *testing.T) {
	r := invoker.Resposta{Status: "falha", Mensagem: "parametro invalido"}
	saida := FormatarResposta(r)
	if !strings.Contains(saida, "[ERRO]") {
		t.Error("esperado [ERRO]")
	}
	if !strings.Contains(saida, "FALHA") {
		t.Error("esperado FALHA")
	}
}

func TestFormatarRespostaTruncaAssinaturaLonga(t *testing.T) {
	r := invoker.Resposta{Status: "sucesso", Assinatura: strings.Repeat("A", 80)}
	saida := FormatarResposta(r)
	if !strings.Contains(saida, "...") {
		t.Error("esperado truncamento com ...")
	}
}

func TestFormatarStatusSimuladorParado(t *testing.T) {
	r := invoker.Resposta{Status: "parado", Mensagem: "Simulador nao esta em execucao"}
	saida := FormatarStatusSimulador(r)
	if !strings.Contains(saida, "[INFO]") {
		t.Error("esperado [INFO] para status nao sucesso")
	}
}

func TestFormatarStatusSimuladorOk(t *testing.T) {
	r := invoker.Resposta{Status: "sucesso", Mensagem: "Tudo certo", Porta: 8443}
	saida := FormatarStatusSimulador(r)
	if !strings.Contains(saida, "[OK]") {
		t.Error("esperado [OK]")
	}
	if !strings.Contains(saida, "8443") {
		t.Error("esperado porta 8443")
	}
}
