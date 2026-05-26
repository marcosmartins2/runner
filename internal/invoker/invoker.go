// Package invoker integra com o assinador.jar atraves dos modos local (CLI)
// e servidor (HTTP).
package invoker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

// PortaPadrao do Assinador em modo HTTP.
const PortaPadrao = 8080

// TimeoutHTTP de cada requisicao HTTP a /api/info, /api/sign, /api/validate.
const TimeoutHTTP = 3 * time.Second

// Resposta representa o JSON devolvido pelo Assinador.
type Resposta struct {
	Status                       string `json:"status,omitempty"`
	Mensagem                     string `json:"mensagem,omitempty"`
	Assinatura                   string `json:"assinatura,omitempty"`
	Timestamp                    string `json:"timestamp,omitempty"`
	Certificado                  string `json:"certificado,omitempty"`
	Algoritmo                    string `json:"algoritmo,omitempty"`
	Porta                        int    `json:"porta,omitempty"`
	TimeoutInatividadeMinutos    int    `json:"timeoutInatividadeMinutos,omitempty"`
}

// EstadoServidor descreve metadados locais do Assinador HTTP iniciado pelo CLI.
type EstadoServidor struct {
	PID         int    `json:"pid"`
	Porta       int    `json:"porta"`
	Jar         string `json:"jar"`
	IniciadoEm  string `json:"iniciadoEm"`
}

// CaminhoEstadoServidor retorna ~/.hubsaude/assinador-server.json.
func CaminhoEstadoServidor(dirHubSaude string) string {
	return filepath.Join(dirHubSaude, "assinador-server.json")
}

// LerEstado carrega o estado salvo do servidor; devolve EstadoServidor zero
// quando o arquivo nao existir.
func LerEstado(caminho string) (EstadoServidor, error) {
	var e EstadoServidor
	bytes, err := os.ReadFile(caminho)
	if err != nil {
		if os.IsNotExist(err) {
			return e, nil
		}
		return e, err
	}
	if err := json.Unmarshal(bytes, &e); err != nil {
		return e, err
	}
	return e, nil
}

// GravarEstado persiste o estado.
func GravarEstado(caminho string, estado EstadoServidor) error {
	if err := os.MkdirAll(filepath.Dir(caminho), 0o755); err != nil {
		return err
	}
	bytes, err := json.MarshalIndent(estado, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(caminho, bytes, 0o644)
}

// RemoverEstado apaga o arquivo de estado se existir.
func RemoverEstado(caminho string) error {
	err := os.Remove(caminho)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// PortaDisponivel verifica se a porta TCP esta livre para escutar em 127.0.0.1.
func PortaDisponivel(porta int) bool {
	l, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(porta))
	if err != nil {
		return false
	}
	l.Close()
	return true
}

// EmExecucao retorna true se /api/info responder com status sucesso.
func EmExecucao(porta int) bool {
	resp, err := RequisicaoJSON(http.MethodGet, urlServidor(porta, "/api/info"), nil)
	if err != nil {
		return false
	}
	return resp.Status == "sucesso"
}

// InvocacaoLocal chama o assinador.jar via java -jar (cold start).
func InvocacaoLocal(caminhoJava, caminhoJar string, args []string) (Resposta, error) {
	var r Resposta
	if _, err := os.Stat(caminhoJar); err != nil {
		return r, fmt.Errorf("assinador.jar nao encontrado em %s", caminhoJar)
	}
	if caminhoJava == "" {
		return r, fmt.Errorf("java nao encontrado: configure JAVA_HOME ou execute 'assinatura provisionar-jdk'")
	}

	ctx, cancelar := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelar()

	completos := append([]string{"-jar", caminhoJar}, args...)
	cmd := exec.CommandContext(ctx, caminhoJava, completos...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		mensagem := stderr.String()
		if mensagem == "" {
			mensagem = err.Error()
		}
		return r, fmt.Errorf("assinador.jar retornou erro: %s", mensagem)
	}

	saida := bytes.TrimSpace(stdout.Bytes())
	if len(saida) == 0 {
		return r, fmt.Errorf("assinador.jar nao retornou resposta")
	}
	if err := json.Unmarshal(saida, &r); err != nil {
		return r, fmt.Errorf("resposta nao e JSON valido: %w", err)
	}
	return r, nil
}

// IniciarServidor garante que o assinador.jar esta rodando em modo HTTP.
// Se ja estiver em execucao, retorna sem erro. Caso contrario, dispara em
// background e aguarda ate ~10s pela primeira resposta de /api/info.
func IniciarServidor(caminhoJava, caminhoJar string, porta, pararAposMinutos int, arquivoEstado string) error {
	if EmExecucao(porta) {
		return nil
	}
	if !PortaDisponivel(porta) {
		return fmt.Errorf("porta %d ja em uso por outro processo", porta)
	}
	if _, err := os.Stat(caminhoJar); err != nil {
		return fmt.Errorf("assinador.jar nao encontrado em %s", caminhoJar)
	}
	if caminhoJava == "" {
		return fmt.Errorf("java nao encontrado: configure JAVA_HOME ou execute 'assinatura provisionar-jdk'")
	}

	args := []string{"-jar", caminhoJar, "--server", "--port", strconv.Itoa(porta)}
	if pararAposMinutos > 0 {
		args = append(args, "--parar-apos-minutos", strconv.Itoa(pararAposMinutos))
	}
	cmd := exec.Command(caminhoJava, args...)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	desconectarProcesso(cmd)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("falha ao iniciar assinador.jar: %w", err)
	}
	pid := cmd.Process.Pid

	prazo := time.Now().Add(10 * time.Second)
	for time.Now().Before(prazo) {
		if EmExecucao(porta) {
			estado := EstadoServidor{
				PID:        pid,
				Porta:      porta,
				Jar:        caminhoJar,
				IniciadoEm: time.Now().UTC().Format(time.RFC3339),
			}
			if err := GravarEstado(arquivoEstado, estado); err != nil {
				return fmt.Errorf("nao foi possivel gravar estado local: %w", err)
			}
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}
	return fmt.Errorf("assinador.jar nao respondeu na porta %d em 10s", porta)
}

// PararServidor envia POST /shutdown.
func PararServidor(porta int, arquivoEstado string) (Resposta, error) {
	resp, err := RequisicaoJSON(http.MethodPost, urlServidor(porta, "/shutdown"), nil)
	if err != nil {
		return resp, err
	}
	_ = RemoverEstado(arquivoEstado)
	return resp, nil
}

// InvocacaoHTTPSign chama POST /api/sign.
func InvocacaoHTTPSign(porta int, documento, certificado string) (Resposta, error) {
	corpo := map[string]string{"documento": documento, "certificado": certificado}
	return RequisicaoJSON(http.MethodPost, urlServidor(porta, "/api/sign"), corpo)
}

// InvocacaoHTTPValidate chama POST /api/validate.
func InvocacaoHTTPValidate(porta int, documento, assinatura string) (Resposta, error) {
	corpo := map[string]string{"documento": documento, "assinatura": assinatura}
	return RequisicaoJSON(http.MethodPost, urlServidor(porta, "/api/validate"), corpo)
}

// RequisicaoJSON executa metodo + url, opcionalmente com corpo JSON.
func RequisicaoJSON(metodo, url string, corpo any) (Resposta, error) {
	var r Resposta
	var body io.Reader
	if corpo != nil {
		dadosCorpo, err := json.Marshal(corpo)
		if err != nil {
			return r, err
		}
		body = bytes.NewReader(dadosCorpo)
	}
	req, err := http.NewRequest(metodo, url, body)
	if err != nil {
		return r, err
	}
	req.Header.Set("Accept", "application/json")
	if corpo != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	cliente := &http.Client{Timeout: TimeoutHTTP}
	resp, err := cliente.Do(req)
	if err != nil {
		return r, fmt.Errorf("falha ao conectar ao assinador HTTP: %w", err)
	}
	defer resp.Body.Close()

	dados, err := io.ReadAll(resp.Body)
	if err != nil {
		return r, err
	}
	if err := json.Unmarshal(dados, &r); err != nil {
		return r, fmt.Errorf("resposta HTTP nao e JSON valido: %w", err)
	}
	if resp.StatusCode/100 != 2 {
		mensagem := r.Mensagem
		if mensagem == "" {
			mensagem = resp.Status
		}
		return r, fmt.Errorf("assinador HTTP retornou %d: %s", resp.StatusCode, mensagem)
	}
	return r, nil
}

func urlServidor(porta int, caminho string) string {
	return "http://127.0.0.1:" + strconv.Itoa(porta) + caminho
}
