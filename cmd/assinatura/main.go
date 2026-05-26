// Comando "assinatura" — CLI nativa para acionar o assinador.jar.
//
// Atende as historias US-01, US-02 e US-04 da especificacao.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/kyriosdata/runner/internal/cli"
	"github.com/kyriosdata/runner/internal/invoker"
	"github.com/kyriosdata/runner/internal/jdk"
	"github.com/kyriosdata/runner/internal/release"
)

// version e sobrescrita em release.yml via "-ldflags -X main.version=<tag>".
var version = "dev"

const portaAssinadorPadrao = invoker.PortaPadrao

func main() {
	codigo := executar(os.Args[1:], os.Stdout, os.Stderr)
	os.Exit(codigo)
}

// executar e a versao testavel do entrypoint.
func executar(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		imprimirUso(stderr)
		return 1
	}

	switch args[0] {
	case "version", "--version", "-v":
		fmt.Fprintf(stdout, "assinatura %s\n", version)
		return 0
	case "-h", "--help", "help":
		imprimirUso(stdout)
		return 0
	case "criar":
		return rodarCriar(args[1:], stdout, stderr)
	case "validar":
		return rodarValidar(args[1:], stdout, stderr)
	case "servidor":
		return rodarServidor(args[1:], stdout, stderr)
	case "provisionar-jdk":
		return rodarProvisionarJDK(stdout, stderr)
	default:
		fmt.Fprintf(stderr, "comando desconhecido: %s\n\n", args[0])
		imprimirUso(stderr)
		return 1
	}
}

func rodarCriar(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("criar", flag.ContinueOnError)
	flags.SetOutput(stderr)
	documento := flags.String("documento", "", "Documento codificado em Base64")
	certificado := flags.String("certificado", "", "Identificador do certificado")
	jar := flags.String("jar", "", "Caminho para o assinador.jar")
	modo := flags.String("modo", "servidor", "servidor | local")
	porta := flags.Int("porta", portaAssinadorPadrao, "Porta do servidor HTTP")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if *documento == "" || *certificado == "" {
		fmt.Fprintln(stderr, "[ERRO]: --documento e --certificado sao obrigatorios.")
		return 1
	}

	caminhoJar, err := resolverJar(*jar)
	if err != nil {
		fmt.Fprintf(stderr, "[ERRO]: %v\n", err)
		return 1
	}

	if *modo == "local" {
		resp, err := invoker.InvocacaoLocal(jdk.LocalizarJava(), caminhoJar, []string{
			"--operacao", "criar",
			"--documento", *documento,
			"--certificado", *certificado,
		})
		if err != nil {
			fmt.Fprintf(stderr, "[ERRO]: %v\n", err)
			return 1
		}
		fmt.Fprintln(stdout, cli.FormatarResposta(resp))
		return 0
	}

	estado, err := caminhoEstadoServidor()
	if err != nil {
		fmt.Fprintf(stderr, "[ERRO]: %v\n", err)
		return 1
	}
	if err := invoker.IniciarServidor(jdk.LocalizarJava(), caminhoJar, *porta, 0, estado); err != nil {
		fmt.Fprintf(stderr, "[ERRO]: %v\n", err)
		return 1
	}
	resp, err := invoker.InvocacaoHTTPSign(*porta, *documento, *certificado)
	if err != nil {
		fmt.Fprintf(stderr, "[ERRO]: %v\n", err)
		return 1
	}
	fmt.Fprintln(stdout, cli.FormatarResposta(resp))
	return 0
}

func rodarValidar(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("validar", flag.ContinueOnError)
	flags.SetOutput(stderr)
	documento := flags.String("documento", "", "Documento codificado em Base64")
	assinatura := flags.String("assinatura", "", "Assinatura codificada em Base64")
	jar := flags.String("jar", "", "Caminho para o assinador.jar")
	modo := flags.String("modo", "servidor", "servidor | local")
	porta := flags.Int("porta", portaAssinadorPadrao, "Porta do servidor HTTP")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if *documento == "" || *assinatura == "" {
		fmt.Fprintln(stderr, "[ERRO]: --documento e --assinatura sao obrigatorios.")
		return 1
	}

	caminhoJar, err := resolverJar(*jar)
	if err != nil {
		fmt.Fprintf(stderr, "[ERRO]: %v\n", err)
		return 1
	}

	if *modo == "local" {
		resp, err := invoker.InvocacaoLocal(jdk.LocalizarJava(), caminhoJar, []string{
			"--operacao", "validar",
			"--documento", *documento,
			"--assinatura", *assinatura,
		})
		if err != nil {
			fmt.Fprintf(stderr, "[ERRO]: %v\n", err)
			return 1
		}
		fmt.Fprintln(stdout, cli.FormatarResposta(resp))
		return 0
	}

	estado, err := caminhoEstadoServidor()
	if err != nil {
		fmt.Fprintf(stderr, "[ERRO]: %v\n", err)
		return 1
	}
	if err := invoker.IniciarServidor(jdk.LocalizarJava(), caminhoJar, *porta, 0, estado); err != nil {
		fmt.Fprintf(stderr, "[ERRO]: %v\n", err)
		return 1
	}
	resp, err := invoker.InvocacaoHTTPValidate(*porta, *documento, *assinatura)
	if err != nil {
		fmt.Fprintf(stderr, "[ERRO]: %v\n", err)
		return 1
	}
	fmt.Fprintln(stdout, cli.FormatarResposta(resp))
	return 0
}

func rodarServidor(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "uso: assinatura servidor <iniciar|parar|status> [--porta N] [--jar PATH] [--parar-apos-minutos N]")
		return 1
	}
	acao := args[0]
	flags := flag.NewFlagSet("servidor", flag.ContinueOnError)
	flags.SetOutput(stderr)
	porta := flags.Int("porta", portaAssinadorPadrao, "Porta do servidor HTTP")
	jar := flags.String("jar", "", "Caminho para o assinador.jar")
	pararApos := flags.Int("parar-apos-minutos", 0, "Encerra o servidor apos N minutos sem interacao")
	if err := flags.Parse(args[1:]); err != nil {
		return 2
	}

	estado, err := caminhoEstadoServidor()
	if err != nil {
		fmt.Fprintf(stderr, "[ERRO]: %v\n", err)
		return 1
	}

	switch acao {
	case "iniciar":
		caminhoJar, err := resolverJar(*jar)
		if err != nil {
			fmt.Fprintf(stderr, "[ERRO]: %v\n", err)
			return 1
		}
		if err := invoker.IniciarServidor(jdk.LocalizarJava(), caminhoJar, *porta, *pararApos, estado); err != nil {
			fmt.Fprintf(stderr, "[ERRO]: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "Assinador em execucao na porta %d.\n", *porta)
		return 0
	case "parar":
		if !invoker.EmExecucao(*porta) {
			fmt.Fprintf(stdout, "Assinador nao esta em execucao na porta %d.\n", *porta)
			return 0
		}
		resp, err := invoker.PararServidor(*porta, estado)
		if err != nil {
			fmt.Fprintf(stderr, "[ERRO]: %v\n", err)
			return 1
		}
		fmt.Fprintln(stdout, cli.FormatarResposta(resp))
		return 0
	case "status":
		if !invoker.EmExecucao(*porta) {
			fmt.Fprintf(stdout, "Assinador nao esta em execucao na porta %d.\n", *porta)
			return 0
		}
		resp, err := invoker.RequisicaoJSON("GET", fmt.Sprintf("http://127.0.0.1:%d/api/info", *porta), nil)
		if err != nil {
			fmt.Fprintf(stderr, "[ERRO]: %v\n", err)
			return 1
		}
		fmt.Fprintln(stdout, cli.FormatarResposta(resp))
		if estadoServidor, err := invoker.LerEstado(estado); err == nil && estadoServidor.Porta == *porta {
			fmt.Fprintf(stdout, "PID: %d | Porta registrada: %d\n", estadoServidor.PID, estadoServidor.Porta)
		}
		return 0
	default:
		fmt.Fprintf(stderr, "acao desconhecida: %s\n", acao)
		return 1
	}
}

func rodarProvisionarJDK(stdout, stderr io.Writer) int {
	manifesto, err := release.BaixarManifesto("")
	if err != nil {
		fmt.Fprintf(stderr, "[ERRO]: nao foi possivel obter o release.json: %v\n", err)
		return 1
	}
	caminho, err := jdk.Provisionar(manifesto)
	if err != nil {
		fmt.Fprintf(stderr, "[ERRO]: %v\n", err)
		return 1
	}
	fmt.Fprintln(stdout, caminho)
	return 0
}

// resolverJar devolve o caminho absoluto do assinador.jar.
// Ordem de busca: argumento --jar, env ASSINADOR_JAR, ../assinador/target/, ~/.hubsaude/assinador.jar.
func resolverJar(arg string) (string, error) {
	candidatos := []string{}
	if arg != "" {
		candidatos = append(candidatos, arg)
	}
	if v := os.Getenv("ASSINADOR_JAR"); v != "" {
		candidatos = append(candidatos, v)
	}
	if exec, err := os.Executable(); err == nil {
		raiz := filepath.Dir(exec)
		candidatos = append(candidatos,
			filepath.Join(raiz, "assinador.jar"),
			filepath.Join(raiz, "..", "assinador", "target", "assinador-1.0.0.jar"),
		)
	}
	if dir, err := jdk.DiretorioHubSaude(); err == nil {
		candidatos = append(candidatos, filepath.Join(dir, "assinador.jar"))
	}
	if wd, err := os.Getwd(); err == nil {
		candidatos = append(candidatos,
			filepath.Join(wd, "assinador", "target", "assinador-1.0.0.jar"),
			filepath.Join(wd, "..", "assinador", "target", "assinador-1.0.0.jar"),
		)
	}
	for _, c := range candidatos {
		if c == "" {
			continue
		}
		abs, err := filepath.Abs(c)
		if err != nil {
			continue
		}
		if _, err := os.Stat(abs); err == nil {
			return abs, nil
		}
	}
	return "", fmt.Errorf("assinador.jar nao encontrado. Forneca --jar <path>, configure ASSINADOR_JAR, " +
		"ou copie o jar para ~/.hubsaude/assinador.jar")
}

func caminhoEstadoServidor() (string, error) {
	dir, err := jdk.DiretorioHubSaude()
	if err != nil {
		return "", err
	}
	return invoker.CaminhoEstadoServidor(dir), nil
}

func imprimirUso(saida io.Writer) {
	uso := strings.Join([]string{
		"assinatura - CLI do Sistema Runner para assinatura digital",
		"",
		"Uso:",
		"  assinatura version",
		"  assinatura criar    --documento <base64> --certificado <id> [--modo servidor|local] [--porta N] [--jar PATH]",
		"  assinatura validar  --documento <base64> --assinatura  <b64> [--modo servidor|local] [--porta N] [--jar PATH]",
		"  assinatura servidor <iniciar|parar|status> [--porta N] [--jar PATH] [--parar-apos-minutos N]",
		"  assinatura provisionar-jdk",
		"",
		"Exemplos:",
		"  assinatura criar --documento SGVsbG8= --certificado cert-001",
		"  assinatura --modo local validar --documento SGVsbG8= --assinatura dGVzdA==",
		"  assinatura servidor iniciar --parar-apos-minutos 30",
	}, "\n")
	fmt.Fprintln(saida, uso)
}
