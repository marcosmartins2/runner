# Plano de implementação — v2 (retrospectivo)

> **Status: concluído.** Todas as histórias foram entregues. Este documento
> consolida a decomposição final dos épicos US-01..US-05 e o que ficou
> implementado em cada uma. Para a evolução cronológica, veja
> [planejamento.md](planejamento.md).

## Premissas

- CLIs em **Go 1.21** (`cmd/assinatura`, `cmd/simulador`), gerando binários
  nativos para `windows/amd64`, `linux/amd64`, `darwin/amd64` —
  ver [ADR 0001](adr/0001-linguagem-go-para-as-clis.md).
- `assinador.jar` e `simulador.jar` em **Java 17** com Maven, mantidos
  como autoridade única de validação/simulação (ADR-livre, herdado da spec).
- Parser de CLI com `stdlib flag` ([ADR 0003](adr/0003-parser-de-cli-com-stdlib-flag.md)).
- Estratégia híbrida de artefatos: upstream real com fallback local
  ([ADR 0004](adr/0004-estrategia-hibrida-artefatos.md)).
- Integridade por SHA-256 + Cosign keyless
  ([ADR 0005](adr/0005-integridade-sha256-e-cosign.md)).

## Rastreabilidade épicos → histórias

| Épico | Descrição | Histórias |
|-------|-----------|-----------|
| US-01 | Invocar `assinador.jar` via CLI | US-01.1 .. US-01.9 |
| US-02 | Simular assinatura digital com validação | US-02.1 .. US-02.4 |
| US-03 | Gerenciar ciclo de vida do simulador | US-03.1 .. US-03.4 |
| US-04 | Provisionar JDK automaticamente | US-04.1 |
| US-05 | Disponibilizar binários multiplataforma | US-05.1 .. US-05.3 |

---

## Sprint 1 — Fundação e entrega contínua

**Objetivo:** estrutura Go, primeiro binário nativo, CI rodando nas três
plataformas.

### US-01.1 — Estrutura base do CLI em Go
- [x] `go mod init github.com/kyriosdata/runner`
- [x] Layout `cmd/assinatura`, `cmd/simulador`, `internal/*`
- [x] `assinatura version` e `simulador version` injetando tag + SHA via `-ldflags`

### US-05.1 — Pipeline CI multiplataforma
- [x] [`.github/workflows/build.yml`](../.github/workflows/build.yml) com matrix
      `ubuntu-latest` + `windows-latest` + `macos-latest`
- [x] `go vet`, `go test ./...`, `mvn test` por SO

### US-05.2 — Publicação de releases SemVer
- [x] [`.github/workflows/release.yml`](../.github/workflows/release.yml)
      disparado por tag `v*`
- [x] Convenção `assinatura-<versão>-<os>-<arch>[.exe]`
- [x] `checksums.txt` consolidado anexado ao release

### US-05.3 — Cosign keyless
- [x] Assinatura OIDC + transparency log no `release.yml`
- [x] `.sig` e `.pem` por artefato
- [x] Comando de verificação documentado no [README](../README.md#verificacao-de-artefatos-cosign)

---

## Sprint 2 — Modo local ponta a ponta

**Objetivo:** usuário roda `assinatura criar` e obtém assinatura simulada
sem configurar Java manualmente.

### US-02.1 — Simulação de criação
- [x] `AssinaturaService` + `RespostaAssinatura` no `assinador.jar`
- [x] Cobertura em `AssinadorAppTest`

### US-02.2 — Validação de criação
- [x] `ValidadorParametros` rejeita entradas inválidas (mensagem clara)
- [x] Exit codes distintos para erro de usuário e do sistema

### US-02.3 — Simulação e validação de `validate`
- [x] `assinador validar` com resultado pré-determinado para parâmetros válidos
- [x] Cenários negativos cobertos em `AssinadorAppTest`

### US-01.2 — Parsing de comandos
- [x] Subcomandos `criar`, `validar`, `servidor {iniciar|status|parar}` em
      [`cmd/assinatura/main.go`](../cmd/assinatura/main.go)
- [x] `--help` com exemplos por subcomando

### US-01.3 — Invocação local via `java -jar`
- [x] [`internal/invoker/invoker.go`](../internal/invoker/invoker.go) localiza
      `java` via `JAVA_HOME` / `PATH` / `~/.hubsaude/jdk/`
- [x] Captura de `stdout`/`stderr` e propagação de exit code
- [x] [`integracao_test.go`](../internal/invoker/integracao_test.go)

### US-01.4 — Exibição legível
- [x] [`internal/cli/formato.go`](../internal/cli/formato.go) formata sucesso/erro
- [x] `--quiet` reduz a saída ao essencial

### US-04.1 — Provisionamento de JDK
- [x] [`internal/jdk/jdk.go`](../internal/jdk/jdk.go) detecta JDK 21 instalado
- [x] [`internal/jdk/extrair.go`](../internal/jdk/extrair.go) baixa Temurin para
      `~/.hubsaude/jdk/` quando ausente
- [x] Suporta `windows/{amd64,arm64}`, `linux/{amd64,arm64}`, `darwin/{amd64,arm64}`

---

## Sprint 3 — Modo servidor e PKCS#11

**Objetivo:** servidor HTTP do `assinador.jar` com gestão de ciclo de vida
pela CLI; suporte opcional a dispositivos criptográficos.

### US-02.4 — Endpoints HTTP do assinador
- [x] `AssinadorHttpServer` com `POST /sign`, `POST /validate`, `GET /health`
- [x] Resposta JSON consistente com o modo CLI

### US-01.5 — Iniciar como servidor
- [x] `assinatura servidor iniciar` sobe o JAR em background
- [x] PID/porta/JAR persistidos em `~/.hubsaude/assinador-server.json`

### US-01.6 — Invocar via HTTP por padrão
- [x] CLI usa modo servidor automaticamente; `--modo local` força cold start

### US-01.7 — Detectar instância em execução
- [x] Health check em `/health` antes de iniciar (ADR 0002); reusa se viva

### US-01.8 — Interromper execução
- [x] `servidor parar` encerra na porta padrão ou na indicada via `--porta`

### US-01.9 — Auto-shutdown por inatividade
- [x] `--parar-apos-minutos N` agenda shutdown; timer reinicia a cada request
- [x] Coberto por `MonitorInatividadeTest` no `assinador/`

### US-02 (extensão) — PKCS#11
- [x] `--pkcs11` carrega `SunPKCS11` quando o dispositivo está disponível
- [x] Mensagem amigável quando o dispositivo está ausente

---

## Sprint 4 — Simulador e segurança de artefatos

**Objetivo:** ciclo de vida completo do `simulador.jar`, integridade dos
artefatos e Cosign.

### US-03.1 — Iniciar o simulador
- [x] `simulador iniciar` baixa JAR + JRE quando ausentes
- [x] Health check via `/api/info`

### US-03.2 — Parar e monitorar
- [x] `simulador parar` chama `/shutdown`
- [x] `simulador status` lê `~/.hubsaude/simulador-server.json`

### US-03.3 — Estrutura base do CLI simulador
- [x] [`cmd/simulador/main.go`](../cmd/simulador/main.go) compartilha
      `internal/*` com o CLI `assinatura`
- [x] Mesma matriz de release que o `assinatura`

### US-03.4 — Obter simulador.jar dinamicamente
- [x] [`internal/release/release.go`](../internal/release/release.go) consulta
      `release.json` em `raw.githubusercontent.com/kyriosdata/runner/main/`
- [x] Verifica SHA-256 antes de usar (ADR 0005)
- [x] `simulador atualizar` força nova consulta

---

## Resumo das sprints

| Sprint | Foco | Status |
|--------|------|--------|
| 1 | Fundação + CI multiplataforma + Cosign | Concluída |
| 2 | Fluxo local end-to-end + provisionamento de JDK | Concluída |
| 3 | Modo servidor + PKCS#11 | Concluída |
| 4 | Simulador + integridade de artefatos | Concluída |

Definition of Done global atendida — ver
[`planejamento.md` §5](planejamento.md#5-definition-of-done-atendida).
