# Sistema Runner

Sistema para facilitar o acesso a funcionalidade de execucao de aplicacoes Java
via linha de comandos. Atende as historias US-01..US-05 da disciplina IIS
(2026-01) e disponibiliza **binarios nativos para Windows, Linux e macOS**.

## Documentacao

- [Especificacao](./especificacao.md) - escopo, US-01 a US-05, entregaveis
- [Design (C4)](./design.md) - contexto, conteineres, decisoes
- [Sprint 1](./docs/sprint-1-tasks.md) - tarefas operacionais
- [Diagramas PlantUML](./diagramas/)

## Status

Implementacao concluida, atendendo ao [issue #1](https://github.com/kyriosdata/runner/issues/1)
("Sao esperadas versoes nativas em 3 plataformas") e cobrindo:

- CLIs `assinatura` e `simulador` em Go, compilados como binarios nativos
  para `windows/amd64`, `linux/amd64` e `darwin/amd64` (US-05).
- Modo local e modo servidor HTTP para o `assinador.jar`, com parada
  programada por inatividade (US-01, US-02).
- Validacao rigorosa de parametros, simulacao de criacao/validacao de
  assinatura digital e suporte opcional a dispositivos PKCS#11 (US-02).
- CLI `simulador` para gerenciar o ciclo de vida do `simulador.jar`,
  incluindo download dinamico via `release.json` (US-03).
- Provisionamento automatico do JRE Temurin (Adoptium) quando ausente
  na maquina do usuario (US-04).
- Pipeline CI/CD com testes em 3 plataformas, cross-compilation e
  assinatura dos artefatos via Cosign/Sigstore (US-05, secao 9 da
  especificacao).

## Estrutura

```text
runner/
|-- .github/workflows/
|   |-- build.yml       # testes + cross-compile em push/PR
|   `-- release.yml     # tags v* -> cross-compile, Cosign, GitHub Release
|-- cmd/
|   |-- assinatura/     # binario "assinatura" (Go)
|   `-- simulador/      # binario "simulador" (Go)
|-- internal/
|   |-- cli/            # formatadores de saida
|   |-- invoker/        # invocacao local/HTTP do assinador.jar
|   |-- jdk/            # localizacao e provisionamento do JRE
|   `-- release/        # leitura de release.json e download
|-- assinador/          # projeto Java do assinador.jar (Maven)
|-- simuladorjar/       # projeto Java do simulador.jar (Maven)
|-- diagramas/          # PlantUML (C4)
|-- docs/
|-- release.json        # manifesto de artefatos (jar + JRE)
|-- go.mod
|-- especificacao.md
|-- design.md
|-- LICENSE
`-- README.md
```

## Plataformas suportadas

Os binarios sao construidos de forma nativa para as tres plataformas-alvo
da disciplina (DT-03):

| Plataforma   | Arquitetura | Artefatos                                                                 |
|--------------|-------------|---------------------------------------------------------------------------|
| Windows      | amd64       | `assinatura-<versao>-windows-amd64.exe`, `simulador-<versao>-windows-amd64.exe` |
| Linux        | amd64       | `assinatura-<versao>-linux-amd64`, `simulador-<versao>-linux-amd64`             |
| macOS        | amd64       | `assinatura-<versao>-darwin-amd64`, `simulador-<versao>-darwin-amd64`           |

Cada release publica tambem `<artefato>.sig` e `<artefato>.pem` para
verificacao com `cosign verify-blob` (secao 9 da especificacao).

## Pre-requisitos

- Para uso dos binarios pre-compilados: nenhum. O CLI baixa o JRE e o
  `simulador.jar` automaticamente quando ausentes (US-03, US-04).
- Para compilar localmente: Go 1.21+ e Maven 3.8+ com JDK 17+.

## Compilacao local

```bash
# CLIs Go
go build -o bin/assinatura ./cmd/assinatura
go build -o bin/simulador  ./cmd/simulador

# JAR assinador
cd assinador && mvn package -DskipTests

# JAR simulador
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

# Versao do CLI (injetada via -ldflags em release)
assinatura version
```

Estado do servidor: `~/.hubsaude/assinador-server.json` (PID, porta, JAR, inicio).

## Uso - CLI `simulador`

```bash
simulador iniciar              # baixa simulador.jar e JRE se necessario
simulador status
simulador parar

# Forcar atualizacao do simulador.jar via release.json
simulador atualizar
```

## Verificacao de artefatos (Cosign)

```bash
cosign verify-blob \
  --certificate-identity-regexp 'https://github.com/kyriosdata/runner/.github/workflows/release.yml@.*' \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  --certificate assinatura-1.0.0-linux-amd64.pem \
  --signature   assinatura-1.0.0-linux-amd64.sig \
  assinatura-1.0.0-linux-amd64
```

## Testes

```bash
go test ./...                        # Go: CLIs + pacotes internos
cd assinador && mvn test             # Java: assinador.jar
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
3. Usuarios podem baixar o binario nativo para sua plataforma e validar
   a autenticidade via `cosign verify-blob`.
