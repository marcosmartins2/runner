# Critérios de aceitação — Sistema Runner

Critérios usados para considerar cada incremento como **pronto** neste
repositório. Derivados da versão upstream
([kyriosdata/runner](https://github.com/kyriosdata/runner)) e adaptados às
decisões registradas em [docs/adr/](adr/).

## A. Princípios transversais

- **Rastreabilidade.** Especificação → ADR → história (US-XX) → commit → código → teste.
  Cada US aparece no [README](../README.md#rastreabilidade-requisito---codigo---teste) e nos commits.
- **Single source of truth.** A especificação pertence ao upstream e é
  referenciada por commit fixo em [`especificacao.md`](../especificacao.md);
  nunca duplicada.
- **Reprodutibilidade.** `go test ./...` + `mvn test` produzem verde em
  qualquer clone. CI cobre Windows + Linux + macOS.
- **Falhar bem.** Exit codes coerentes, mensagens com *o quê / por quê /
  como resolver*; `stdout` reservado ao resultado, `stderr` ao diagnóstico.
- **Decisões registradas.** Cada decisão não trivial vira ADR em
  [`docs/adr/`](adr/) (porta padrão, parser CLI, artefatos, integridade).

## B. Organização do repositório

- Layout multi-módulo: `cmd/<binário>/`, `internal/<pacote>/`, `assinador/`,
  `simuladorjar/`, `diagramas/`, `docs/` — registrado em
  [planejamento.md DT-06](planejamento.md).
- `.gitignore` cobre `bin/`, `dist/`, `target/`, `.idea/`, `.DS_Store`,
  caches do Go.
- `.gitattributes` normaliza EOL (impede `\r\n` em arquivos `.go`/`.java`).
- Zero artefato compilado versionado.
- `LICENSE` Apache-2.0, compatível com as dependências.
- Idioma único em paths (português sem acento), nomes em snake_case nos
  arquivos `.go` (`integracao_test.go`) e PascalCase nos `.java`.

## C. Documentação (mínima, específica, viva)

- [`README.md`](../README.md) é o contrato: o que é, como compilar, como
  executar, como testar, status atual, plataformas suportadas.
- [`especificacao.md`](../especificacao.md) referencia o upstream por
  commit fixo (`4d7d40f...`), não por `main`.
- ADRs curtas registram decisões com **Status / Contexto / Decisão /
  Consequências** ([índice](adr/README.md)).
- [`docs/planejamento.md`](planejamento.md) reflete o estado real, com
  US-XX marcadas e DTs documentadas.
- [`docs/operacao.md`](operacao.md) documenta o ciclo de vida do servidor
  HTTP (start, status, stop, inatividade).

## D. Qualidade de código

- Separação **transporte / domínio / interface**:
  - transporte: [`internal/invoker/`](../internal/invoker/) (subprocess + HTTP);
  - domínio: [`assinador/src/main/java/...`](../assinador/src/main/java/com/runner/assinador/) (`AssinaturaService`, `ValidadorParametros`);
  - interface: [`cmd/assinatura/`](../cmd/assinatura/), [`cmd/simulador/`](../cmd/simulador/).
- Parser de CLI com `stdlib flag` (ADR 0003) — sem dependência externa.
- Logging estruturado via `slog` em [`internal/logging/`](../internal/logging/),
  com `--verbose` e `--quiet` previsíveis.
- Sem `catch (Throwable)` ou `if err == nil` mudo; erro propagado com
  contexto (`fmt.Errorf("...%w...", err)`).
- Encoding UTF-8 declarado; EOL normalizado pelo `.gitattributes`.
- Sem segredos, IPs ou portas hardcoded fora do `release.json`.

## E. Requisitos funcionais (comportamento observável)

### E1. Invocação local do `assinador.jar` (US-01)

- Executável funciona independentemente do `cwd`.
- Argumentos preservam espaços, acentos e aspas (validado em
  `internal/invoker/invoker_test.go`).
- `exit code` propagado; `stdout` carrega o JSON; `stderr` o diagnóstico.

### E2. Modo servidor HTTP (US-01.5 a US-01.9)

- `iniciar` é idempotente: faz health check real em `/health`, não só
  verifica porta ocupada (ADR 0002).
- Porta padrão `9099` (assinador) e `8443` (simulador) configurável via
  `--porta`; falha clara quando ocupada.
- `parar` aceita porta corrente e a indicada via `--porta`.
- Auto-shutdown por inatividade (`--parar-apos-minutos N`) com timer que
  reinicia a cada requisição — coberto por
  [`MonitorInatividadeTest`](../assinador/src/test/java/com/runner/assinador/).
- Modo servidor é o padrão; `--modo local` ativa explicitamente cold start.

### E3. Validação de parâmetros (US-02.2, US-02.3)

- Feita **dentro do `assinador.jar`** (autoridade única), não replicada na CLI Go.
- Exit codes distintos para erro de usuário e erro do sistema.

### E4. Simulador do HubSaúde (US-03)

- Ciclo de vida em [`cmd/simulador/main.go`](../cmd/simulador/main.go):
  `iniciar`, `status`, `parar`, `atualizar`.
- Health check via `/api/info`; "processo subiu" ≠ "pronto para requisição".
- `simulador atualizar` consulta [`release.json`](../release.json),
  baixa o JAR e **verifica SHA-256** (ADR 0005); aborta em divergência.

### E5. PKCS#11 (US-02 — opcional)

- `assinador --pkcs11` carrega `SunPKCS11` quando disponível; mensagem
  amigável se o dispositivo está ausente.

### E6. Provisionamento de JDK (US-04)

- Detecção via `JAVA_HOME` e `PATH`; download Temurin 21 quando ausente.
- Cache em `~/.hubsaude/jdk/`; reuso entre execuções.
- Suporta `windows/amd64`, `windows/arm64`, `linux/amd64`, `linux/arm64`,
  `darwin/amd64`, `darwin/arm64` via `release.json`.

### E7. Portabilidade real (US-05)

- Cross-compile para `windows/amd64`, `linux/amd64`, `darwin/amd64` em
  [`.github/workflows/release.yml`](../.github/workflows/release.yml).
- Testes na matriz dos 3 SO em [`build.yml`](../.github/workflows/build.yml).

## F. Build, dependências, supply chain

- `go.mod` declara `go 1.21`; dependências do Go limitadas à stdlib.
- `assinador.jar` e `simulador.jar` empacotados como *fat JAR* com
  `Main-Class` no manifest.
- `release.json` referencia artefatos upstream por **tag fixa** e **SHA-256**
  (ADR 0004, ADR 0005). Verificação em
  [`internal/release/release.go`](../internal/release/release.go).
- Cosign keyless (OIDC + transparency log) assina cada artefato no release;
  `.sig` e `.pem` publicados junto.

## G. Testes

- Pirâmide saudável: unitários por pacote, integração em
  [`integracao_test.go`](../internal/invoker/integracao_test.go),
  end-to-end via CI nos 3 SO.
- Cenários negativos como cidadãos de primeira classe: porta ocupada,
  JAR ausente, JVM ausente, timeout, payload inválido, race no start.
- Sem testes flaky tolerados.
- Cobertura como sinal, não meta.

## H. Engenharia de processo

- Commits atômicos em Conventional Commits, citando a US: `feat(invoker):
  ... (US-01)`.
- PRs ligados a issues; CI obrigatório bloqueia merge.
- Tags `vX.Y.Z` coerentes com `release.json`; release notes do `gh release`
  geradas a partir dos commits.

## I. Operabilidade

- `--help` com exemplos por subcomando (não só lista de flags).
- `--version` retorna `<tag> (<sha-curto>)` — injetado via
  `-ldflags "-X main.version=... -X main.commit=..."`.
- `--verbose` aumenta verbosidade do `slog`; `--quiet` silencia tudo
  exceto erros.
- Estado local rastreável em `~/.hubsaude/assinador-server.json` e
  `~/.hubsaude/simulador-server.json` (PID, porta, JAR, início).
