// Comando "simulador" — CLI nativa para gerenciar o ciclo de vida do
// simulador.jar (HubSaude).
//
// Atende as historias US-03 e US-04.
package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/kyriosdata/runner/internal/cli"
	"github.com/kyriosdata/runner/internal/invoker"
	"github.com/kyriosdata/runner/internal/jdk"
	"github.com/kyriosdata/runner/internal/release"
)

var version = "dev"

const (
	portaSimuladorPadrao = 8443
	timeoutHTTP          = 5 * time.Second
)

func main() {
	codigo := executar(os.Args[1:], os.Stdout, os.Stderr)
	os.Exit(codigo)
}

func executar(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		imprimirUso(stderr)
		return 1
	}
	switch args[0] {
	case "version", "--version", "-v":
		fmt.Fprintf(stdout, "simulador %s\n", version)
		return 0
	case "-h", "--help", "help":
		imprimirUso(stdout)
		return 0
	case "iniciar":
		return rodarIniciar(args[1:], stdout, stderr)
	case "parar":
		return rodarParar(args[1:], stdout, stderr)
	case "status":
		return rodarStatus(args[1:], stdout, stderr)
	case "atualizar":
		return rodarAtualizar(stdout, stderr)
	default:
		fmt.Fprintf(stderr, "comando desconhecido: %s\n\n", args[0])
		imprimirUso(stderr)
		return 1
	}
}

func rodarIniciar(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("iniciar", flag.ContinueOnError)
	flags.SetOutput(stderr)
	porta := flags.Int("porta", portaSimuladorPadrao, "Porta do Simulador")
	jar := flags.String("jar", "", "Caminho para simulador.jar (padrao: ~/.hubsaude/simulador.jar)")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if EmExecucao(*porta) {
		fmt.Fprintf(stdout, "Simulador ja esta em execucao na porta %d.\n", *porta)
		return 0
	}
	if !invoker.PortaDisponivel(*porta) {
		fmt.Fprintf(stderr, "[ERRO]: porta %d ja em uso por outro processo.\n", *porta)
		return 1
	}

	caminhoJar, err := resolverJar(*jar)
	if err != nil {
		// Tentar baixar via release.json
		fmt.Fprintln(stderr, "simulador.jar nao encontrado; baixando a versao mais recente...")
		caminhoJar, err = baixarJarMaisRecente(stderr)
		if err != nil {
			fmt.Fprintf(stderr, "[ERRO]: %v\n", err)
			return 1
		}
	}

	caminhoJava := jdk.LocalizarJava()
	if caminhoJava == "" {
		fmt.Fprintln(stderr, "java nao encontrado; provisionando JRE...")
		caminhoJava, err = provisionarJDK(stderr)
		if err != nil {
			fmt.Fprintf(stderr, "[ERRO]: %v\n", err)
			return 1
		}
	}

	cmd := exec.Command(caminhoJava, "-jar", caminhoJar, "--server.port="+strconv.Itoa(*porta))
	desconectarProcesso(cmd)
	if err := cmd.Start(); err != nil {
		fmt.Fprintf(stderr, "[ERRO]: falha ao iniciar simulador.jar: %v\n", err)
		return 1
	}

	prazo := time.Now().Add(30 * time.Second)
	for time.Now().Before(prazo) {
		if EmExecucao(*porta) {
			estado := invoker.EstadoServidor{
				PID:        cmd.Process.Pid,
				Porta:      *porta,
				Jar:        caminhoJar,
				IniciadoEm: time.Now().UTC().Format(time.RFC3339),
			}
			if err := gravarEstadoSimulador(estado); err != nil {
				fmt.Fprintf(stderr, "[ERRO]: nao foi possivel gravar estado: %v\n", err)
			}
			fmt.Fprintf(stdout, "Simulador em execucao na porta %d.\n", *porta)
			return 0
		}
		time.Sleep(500 * time.Millisecond)
	}
	fmt.Fprintf(stderr, "[ERRO]: simulador nao respondeu na porta %d em 30s.\n", *porta)
	return 1
}

func rodarParar(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("parar", flag.ContinueOnError)
	flags.SetOutput(stderr)
	porta := flags.Int("porta", portaSimuladorPadrao, "Porta do Simulador")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if !EmExecucao(*porta) {
		fmt.Fprintf(stdout, "Simulador nao esta em execucao na porta %d.\n", *porta)
		return 0
	}
	if _, err := requisicaoSimulador(http.MethodPost, *porta, "/shutdown"); err != nil {
		fmt.Fprintf(stderr, "[ERRO]: %v\n", err)
		return 1
	}
	_ = removerEstadoSimulador()
	fmt.Fprintln(stdout, cli.FormatarStatusSimulador(invoker.Resposta{
		Status: "sucesso", Mensagem: "Simulador interrompido.",
	}))
	return 0
}

func rodarStatus(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("status", flag.ContinueOnError)
	flags.SetOutput(stderr)
	porta := flags.Int("porta", portaSimuladorPadrao, "Porta do Simulador")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if !EmExecucao(*porta) {
		fmt.Fprintln(stdout, cli.FormatarStatusSimulador(invoker.Resposta{
			Status:   "parado",
			Mensagem: fmt.Sprintf("Simulador nao esta em execucao na porta %d.", *porta),
		}))
		return 0
	}
	resp, err := requisicaoSimulador(http.MethodGet, *porta, "/api/info")
	if err != nil {
		fmt.Fprintf(stderr, "[ERRO]: %v\n", err)
		return 1
	}
	resp.Porta = *porta
	fmt.Fprintln(stdout, cli.FormatarStatusSimulador(resp))
	if estado, err := lerEstadoSimulador(); err == nil && estado.Porta == *porta {
		fmt.Fprintf(stdout, "PID local: %d | JAR: %s\n", estado.PID, estado.Jar)
	}
	return 0
}

func rodarAtualizar(stdout, stderr io.Writer) int {
	if _, err := baixarJarMaisRecente(stderr); err != nil {
		fmt.Fprintf(stderr, "[ERRO]: %v\n", err)
		return 1
	}
	caminho, err := caminhoSimuladorJar()
	if err != nil {
		fmt.Fprintf(stderr, "[ERRO]: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "simulador.jar atualizado em %s\n", caminho)
	return 0
}

// EmExecucao testa /api/info em http(s)://127.0.0.1:porta.
// O Simulador pode subir em HTTP ou HTTPS — tentamos ambos.
func EmExecucao(porta int) bool {
	for _, esquema := range []string{"http", "https"} {
		if resp, err := requisicaoEsquema(esquema, http.MethodGet, porta, "/api/info"); err == nil {
			estado := strings.ToLower(resp.Status)
			if estado == "sucesso" || estado == "ok" || estado == "up" || estado == "running" {
				return true
			}
		}
	}
	return false
}

func requisicaoSimulador(metodo string, porta int, caminho string) (invoker.Resposta, error) {
	for _, esquema := range []string{"http", "https"} {
		if resp, err := requisicaoEsquema(esquema, metodo, porta, caminho); err == nil {
			return resp, nil
		}
	}
	return invoker.Resposta{}, fmt.Errorf("nao foi possivel conectar ao Simulador na porta %d", porta)
}

func requisicaoEsquema(esquema, metodo string, porta int, caminho string) (invoker.Resposta, error) {
	var r invoker.Resposta
	url := fmt.Sprintf("%s://127.0.0.1:%d%s", esquema, porta, caminho)
	req, err := http.NewRequest(metodo, url, nil)
	if err != nil {
		return r, err
	}
	req.Header.Set("Accept", "application/json")

	ctx, cancel := context.WithTimeout(context.Background(), timeoutHTTP)
	defer cancel()
	req = req.WithContext(ctx)

	transp := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // localhost self-signed
		DialContext: (&net.Dialer{Timeout: timeoutHTTP}).DialContext,
	}
	cliente := &http.Client{Timeout: timeoutHTTP, Transport: transp}
	resp, err := cliente.Do(req)
	if err != nil {
		return r, err
	}
	defer resp.Body.Close()
	dados, err := io.ReadAll(resp.Body)
	if err != nil {
		return r, err
	}
	if err := json.Unmarshal(dados, &r); err != nil {
		// Permitir respostas em texto puro
		r.Status = "sucesso"
		r.Mensagem = string(dados)
	}
	return r, nil
}

func resolverJar(arg string) (string, error) {
	if arg != "" {
		abs, _ := filepath.Abs(arg)
		if _, err := os.Stat(abs); err == nil {
			return abs, nil
		}
		return "", fmt.Errorf("simulador.jar nao encontrado em %s", abs)
	}
	caminho, err := caminhoSimuladorJar()
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(caminho); err != nil {
		return "", fmt.Errorf("simulador.jar nao encontrado em %s", caminho)
	}
	return caminho, nil
}

func caminhoSimuladorJar() (string, error) {
	dir, err := jdk.DiretorioHubSaude()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "simulador.jar"), nil
}

func baixarJarMaisRecente(stderr io.Writer) (string, error) {
	manifesto, err := release.BaixarManifesto("")
	if err != nil {
		return "", err
	}
	caminho, err := caminhoSimuladorJar()
	if err != nil {
		return "", err
	}

	url := manifesto.Jar.URL
	versao := manifesto.Jar.Version
	if manifesto.Sim != nil && manifesto.Sim.URL != "" {
		url = manifesto.Sim.URL
		versao = manifesto.Sim.Version
	}
	if url == "" {
		return "", fmt.Errorf("manifesto nao contem URL do simulador.jar")
	}

	instalada := release.VersaoInstalada(caminho)
	if instalada == versao {
		fmt.Fprintf(stderr, "simulador.jar ja esta na versao %s.\n", versao)
		return caminho, nil
	}

	fmt.Fprintf(stderr, "baixando simulador.jar versao %s...\n", versao)
	if err := release.BaixarArquivo(url, caminho); err != nil {
		return "", err
	}
	if err := release.GravarVersao(caminho, versao); err != nil {
		return "", err
	}
	return caminho, nil
}

func provisionarJDK(stderr io.Writer) (string, error) {
	manifesto, err := release.BaixarManifesto("")
	if err != nil {
		return "", err
	}
	fmt.Fprintln(stderr, "baixando JRE Temurin (~250 MB)...")
	return jdk.Provisionar(manifesto)
}

func caminhoEstado() (string, error) {
	dir, err := jdk.DiretorioHubSaude()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "simulador-server.json"), nil
}

func gravarEstadoSimulador(estado invoker.EstadoServidor) error {
	caminho, err := caminhoEstado()
	if err != nil {
		return err
	}
	return invoker.GravarEstado(caminho, estado)
}

func lerEstadoSimulador() (invoker.EstadoServidor, error) {
	caminho, err := caminhoEstado()
	if err != nil {
		return invoker.EstadoServidor{}, err
	}
	return invoker.LerEstado(caminho)
}

func removerEstadoSimulador() error {
	caminho, err := caminhoEstado()
	if err != nil {
		return err
	}
	return invoker.RemoverEstado(caminho)
}

func imprimirUso(saida io.Writer) {
	uso := strings.Join([]string{
		"simulador - CLI do Sistema Runner para o Simulador HubSaude",
		"",
		"Uso:",
		"  simulador version",
		"  simulador iniciar  [--porta N] [--jar PATH]",
		"  simulador parar    [--porta N]",
		"  simulador status   [--porta N]",
		"  simulador atualizar    (baixa a ultima versao do simulador.jar via release.json)",
		"",
		"Exemplos:",
		"  simulador iniciar",
		"  simulador --porta 9443 iniciar --jar /caminho/simulador.jar",
		"  simulador status",
	}, "\n")
	fmt.Fprintln(saida, uso)
}
