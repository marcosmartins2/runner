# Sistema Runner

Sistema para facilitar o acesso a funcionalidade de execucao de aplicacoes Java
via linha de comandos. Atende as historias US-01..US-05 da disciplina IIS
(2026-01) e disponibiliza **binarios nativos para Windows, Linux e macOS**.

As CLIs sao escritas em Go; por padrao consomem os **artefatos reais do
upstream** (`hubsaude-validador-api` e `hubsaude-simulador`) baixados via
`release.json` com verificacao de integridade SHA256, mantendo um *fallback*
local para desenvolvimento (ver [ADR 0004](docs/adr/0004-estrategia-hibrida-artefatos.md)).

## Documentacao

A **especificacao, o design upstream e os criterios** pertencem ao repositorio
da disciplina e sao referenciados por **commit fixo** (nunca `main`), para nao
quebrar a rastreabilidade:

- [Especificacao (upstream, commit fixo)](https://github.com/kyriosdata/runner/blob/4d7d40fff32b3b50372e7fbe41fe713b2bbddb4c/especificacao.md)
- [Criterios de aceitacao (upstream, commit fixo)](https://github.com/kyriosdata/runner/blob/4d7d40fff32b3b50372e7fbe41fe713b2bbddb4c/docs/criterios.md)

Documentacao **especifica desta implementacao**:

- [Design (C4)](./design.md) - contexto, conteineres, decisoes
- [ADRs](./docs/adr/) - decisoes arquiteturais (porta padrao, parser CLI, artefatos, integridade)
- [Operacao do servidor](./docs/operacao.md) - ciclo de vida do Assinador HTTP
- [Planejamento](./docs/planejamento.md) - tarefas da construcao
- [Diagramas PlantUML](./diagramas/)

## Status

Implementacao concluida, cobrindo:

- CLIs `assinatura` e `simulador` em Go, compiladas como binarios nativos
  para `windows/amd64`, `linux/amd64` e `darwin/amd64` (US-05).
- Modo local e modo servidor HTTP para o assinador, com parada programada por
  inatividade (US-01, US-02).
- Validacao rigorosa de parametros feita **dentro do JAR** (autoridade unica),
  simulacao de criacao/validacao e suporte opcional a PKCS#11 (US-02).
- Download dinamico dos JARs reais via `release.json` com verificacao SHA256
  (US-03) e provisionamento automatico do JRE Temurin (US-04).
- Pipeline CI/CD com lint, testes em 3 plataformas, cross-compilation e
  assinatura dos artefatos via Cosign/Sigstore (US-05, secao 9 da especificacao).

## Rastreabilidade (requisito -> codigo -> teste)

| US | Onde esta implementado | Testes |
|----|------------------------|--------|
| US-01 invocar assinador (local/HTTP) | [internal/invoker](internal/invoker/), [cmd/assinatura](cmd/assinatura/main.go) | `invoker_test.go`, `integracao_test.go` |
| US-02 validar parametros + simular | [assinador/](assinador/src/main/java/com/runner/assinador/) | `AssinadorAppTest.java`, `MonitorInatividadeTest.java` |
| US-03 ciclo de vida do simulador + download | [cmd/simulador](cmd/simulador/main.go), [internal/release](internal/release/) | `version_test.go`, `release_test.go` |
| US-04 provisionar JRE | [internal/jdk](internal/jdk/) | `jdk_test.go` |
| US-05 binarios + assinatura | [.github/workflows](.github/workflows/) | matriz Win/Linux/macOS |

## Estrutura

```text
runner/
|-- .github/workflows/
|   |-- build.yml       # lint + testes + cross-compile em push/PR
|   `-- release.yml     # tags v* -> cross-compile, Cosign, GitHub Release
|-- cmd/
|   |-- assinatura/     # binario "assinatura" (Go)
|   `-- simulador/      # binario "simulador" (Go)
|-- internal/
|   |-- cli/            # formatadores de saida
|   |-- invoker/        # invocacao local/HTTP do assinador
|   |-- jdk/            # localizacao e provisionamento do JRE
|   |-- logging/        # logs estruturados (slog) com --verbose/--quiet
|   `-- release/        # release.json, download e verificacao SHA256
|-- assinador/          # JAR de referencia/fallback (Maven)
|-- simuladorjar/       # JAR de referencia/fallback (Maven)
|-- diagramas/          # PlantUML (C4)
|-- docs/
|   |-- adr/            # Architecture Decision Records
|   |-- operacao.md
|   `-- planejamento.md
|-- release.json        # manifesto (artefatos reais do upstream + sha256)
|-- .gitattributes      # normalizacao de EOL/encoding
|-- go.mod
|-- design.md
|-- especificacao.md    # ponteiro para a spec upstream (commit fixo)
|-- LICENSE
`-- README.md
```

## Plataformas suportadas

| Plataforma   | Arquitetura | Artefatos                                                                 |
|--------------|-------------|---------------------------------------------------------------------------|
| Windows      | amd64       | `assinatura-<versao>-windows-amd64.exe`, `simulador-<versao>-windows-amd64.exe` |
| Linux        | amd64       | `assinatura-<versao>-linux-amd64`, `simulador-<versao>-linux-amd64`             |
| macOS        | amd64       | `assinatura-<versao>-darwin-amd64`, `simulador-<versao>-darwin-amd64`           |

Cada release publica tambem `<artefato>.sig` e `<artefato>.pem` para
verificacao com `cosign verify-blob` (secao 9 da especificacao). O provisionamento
de JRE ja suporta tambem alvos arm64 (Windows/Linux/macOS) via `release.json`.

## Pre-requisitos

- Para uso dos binarios pre-compilados: nenhum. O CLI baixa o JRE e os JARs
  automaticamente quando ausentes (US-03, US-04).
- Para compilar localmente: Go 1.21+ e Maven 3.8+ com JDK 17+.

## Compilacao local

```bash
# CLIs Go
go build -o bin/assinatura ./cmd/assinatura
go build -o bin/simulador  ./cmd/simulador

# JARs de referencia/fallback
cd assinador    && mvn package -DskipTests
cd simuladorjar && mvn package -DskipTests
```

## Uso - CLI `assinatura`

```bash
# Criar assinatura digital (simulacao)
assinatura criar --documento SGVsbG8= --certificado cert-001

# Validar uma assinatura
assinatura validar --documento SGVsbG8= --assinatura <assinatura>

# Forcar invocacao local em vez do modo servidor (cold start)
assinatura --modo local criar --documento SGVsbG8= --certificado cert-001

# Servidor HTTP
assinatura servidor iniciar
assinatura servidor iniciar --parar-apos-minutos 30
assinatura servidor status
assinatura servidor parar

# Diagnostico detalhado / silencioso (flags globais)
assinatura --verbose criar --documento SGVsbG8= --certificado cert-001
assinatura --quiet   servidor status

# Versao (tag + SHA curto, injetados via -ldflags em release)
assinatura version        # ex.: "assinatura v1.0.0 (a1b2c3d)"
```

Estado do servidor: `~/.hubsaude/assinador-server.json` (PID, porta, JAR, inicio).

## Uso - CLI `simulador`

```bash
simulador iniciar              # baixa simulador.jar e JRE se necessario
simulador status
simulador parar
simulador atualizar            # forca atualizacao do simulador.jar via release.json
simulador --verbose iniciar    # diagnostico detalhado
```

## Estrategia de artefatos e integridade

- O `release.json` referencia os artefatos **reais do upstream** por tag fixa e
  `sha256`. Em runtime o CLI busca o manifesto em
  `raw.githubusercontent.com/kyriosdata/runner/main/release.json`, baixa o JAR e
  **verifica o SHA256** antes de usar; em divergencia, aborta com erro explicito.
- Havendo um JAR local (`--jar`, `ASSINADOR_JAR`, `*/target/`, `~/.hubsaude/`),
  ele e usado sem rede (fallback de desenvolvimento).
- Detalhes em [ADR 0004](docs/adr/0004-estrategia-hibrida-artefatos.md) e
  [ADR 0005](docs/adr/0005-integridade-sha256-e-cosign.md).

## Verificacao de artefatos (Cosign)

```bash
cosign verify-blob \
  --certificate-identity-regexp 'https://github.com/marcosmartins2/runner/.github/workflows/release.yml@.*' \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  --certificate assinatura-v1.0.0-linux-amd64.pem \
  --signature   assinatura-v1.0.0-linux-amd64.sig \
  assinatura-v1.0.0-linux-amd64
```

## Testes

```bash
go test ./...                        # Go: CLIs + pacotes internos (inclui testes de integracao/negativos)
cd assinador    && mvn test          # Java: assinador.jar (inclui MonitorInatividade)
cd simuladorjar && mvn test          # Java: simulador.jar
```

## Diagramas (C4)

Para gerar os SVGs a partir dos `.puml`:

```bash
./geraimagens.sh        # Linux/macOS
geraimagens.bat         # Windows
```

## Releases multiplataforma

1. Crie uma tag `v*` (ex.: `git tag v1.0.0 && git push origin v1.0.0`).
2. O workflow `release.yml` executa testes em 3 plataformas, faz
   cross-compilation, assina os artefatos com Cosign e publica no
   GitHub Releases junto com `checksums.txt` e `release.json`.
3. Usuarios baixam o binario nativo da sua plataforma e validam a autenticidade
   via `cosign verify-blob`.
