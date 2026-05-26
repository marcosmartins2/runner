# Planejamento de construcao - runner

## 1. Status atual (2026-05)

Apos a observacao do professor no [issue #1](https://github.com/kyriosdata/runner/issues/1)
sobre a necessidade de versoes nativas em 3 plataformas, o repositorio
foi reorganizado para seguir o plano oficial:

- CLIs `assinatura` e `simulador` reescritos em Go, gerando binarios
  nativos para `windows/amd64`, `linux/amd64` e `darwin/amd64`;
- aplicacao `assinador.jar` mantida em Java 17 (Maven), com suporte
  opcional a dispositivos PKCS#11 (token/smart card);
- aplicacao `simulador.jar` adicionada como projeto Java independente
  (`simuladorjar/`), expondo `/api/info` e `/shutdown`;
- pipeline CI/CD com testes em 3 plataformas, cross-compilation, geracao
  de `checksums.txt` e assinatura keyless via Cosign/Sigstore.

## 2. Decisoes tecnicas

| ID | Decisao | Valor |
|---|---|---|
| DT-01 | Modulo Go | `github.com/kyriosdata/runner` |
| DT-02 | Branch principal | `main` |
| DT-03 | Plataformas-alvo | `windows/amd64`, `linux/amd64`, `darwin/amd64` |
| DT-04 | Convencao de nome dos artefatos | `<binario>-<tag>-<os>-<arch>[.exe]` |
| DT-05 | Checksums | `sha256sum` consolidado em `checksums.txt` |
| DT-06 | Layout | `cmd/`, `internal/`, `assinador/`, `simuladorjar/` |
| DT-07 | Distribuicao | GitHub Releases via tag `v*` |
| DT-08 | Assinatura | Cosign keyless (OIDC) com `<artefato>.sig` e `<artefato>.pem` |

## 3. Mapeamento US -> implementacao

| US | Onde foi atendido |
|---|---|
| US-01 | `cmd/assinatura`, `internal/invoker` (CLI local e HTTP) |
| US-02 | `assinador/` (Java), `ValidadorParametros`, `AssinaturaService`, suporte `--pkcs11` |
| US-03 | `cmd/simulador`, `internal/release`, `simuladorjar/` |
| US-04 | `internal/jdk` (deteccao + download Temurin) |
| US-05 | `.github/workflows/release.yml`, Cosign, cross-compile |

## 4. Iteracoes concluidas

- **Iteracao 1 (consolidacao do modo local).** Validacao e simulacao
  do assinador em Java; estrutura testavel.
- **Iteracao 2 (modo servidor).** Endpoints HTTP, start/stop/status,
  parada programada por inatividade.
- **Iteracao 3 (CLI nativa).** Migracao das CLIs para Go, gerando
  binarios para 3 plataformas e abrindo caminho para distribuicao via
  GitHub Releases. Atende explicitamente ao issue #1.
- **Iteracao 4 (distribuicao).** Workflow `release.yml`, Cosign,
  checksums, atualizacao de `release.json`.

## 5. Definition of Done atendida

- [x] `go build ./...` produz `assinatura` e `simulador` para Windows,
      Linux e macOS;
- [x] `go test ./...` passa nas tres plataformas (matriz no
      `build.yml`);
- [x] `mvn test` passa em `assinador/` e `simuladorjar/`;
- [x] Pipeline CI compila e testa todo o conjunto em push/PR para main;
- [x] Tags `v*` publicam release com binarios, JARs, `checksums.txt`,
      `release.json` e arquivos `.sig`/`.pem` por artefato;
- [x] `cosign verify-blob` reconhece os artefatos como autenticos.

## 6. Comandos de trabalho

```bash
go test ./...                          # testes Go
cd assinador && mvn test               # testes do assinador.jar
cd simuladorjar && mvn test            # testes do simulador.jar
go build -o bin/assinatura ./cmd/assinatura
go build -o bin/simulador  ./cmd/simulador
```

## 7. Operacao do CLI em modo servidor

Detalhes em [operacao.md](./operacao.md).
