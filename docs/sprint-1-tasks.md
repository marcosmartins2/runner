# Sprint 1 — Tarefas Operacionais

Decomposição operacional da Sprint 1 (fundação Go + CI/CD), tal como foi
executada neste repositório. Para o histórico cronológico veja
[planejamento.md](planejamento.md); para as decisões transversais veja
[adr/](adr/).

## Decisões técnicas (DT)

| # | Decisão | Valor | Referência |
|---|---------|-------|------------|
| DT-01 | Módulo Go | `github.com/kyriosdata/runner` | [go.mod](../go.mod) |
| DT-02 | Branch principal | `main` | — |
| DT-03 | Plataformas-alvo | `windows/amd64`, `linux/amd64`, `darwin/amd64` | [release.yml](../.github/workflows/release.yml) |
| DT-04 | Nome dos artefatos | `<binário>-<tag>-<os>-<arch>[.exe]` | [release.yml](../.github/workflows/release.yml) |
| DT-05 | Checksums | `sha256sum` consolidado em `checksums.txt` | [release.yml](../.github/workflows/release.yml) |
| DT-06 | Layout | `cmd/`, `internal/`, `assinador/`, `simuladorjar/` | [README](../README.md#estrutura) |
| DT-07 | Distribuição | GitHub Releases via tag `v*` | [release.yml](../.github/workflows/release.yml) |
| DT-08 | Assinatura | Cosign keyless (OIDC) + transparency log | [ADR 0005](adr/0005-integridade-sha256-e-cosign.md) |
| DT-09 | Parser CLI | `stdlib flag` | [ADR 0003](adr/0003-parser-de-cli-com-stdlib-flag.md) |

### DT-06 — layout efetivo

```
runner/
├── cmd/
│   ├── assinatura/       ← binário principal
│   │   ├── main.go
│   │   └── version_test.go
│   └── simulador/        ← segundo binário
│       ├── main.go
│       ├── detach_unix.go
│       ├── detach_windows.go
│       └── version_test.go
├── internal/
│   ├── cli/              ← formatadores de saída
│   ├── invoker/          ← invocação local + HTTP do assinador.jar
│   ├── jdk/              ← detecção e provisionamento do JRE
│   ├── logging/          ← slog estruturado (--verbose / --quiet)
│   └── release/          ← release.json, download e verificação SHA-256
├── assinador/            ← JAR Maven (autoridade da simulação)
├── simuladorjar/         ← JAR Maven (Simulador do HubSaúde)
├── .github/workflows/
│   ├── build.yml         ← lint + testes em push/PR
│   └── release.yml       ← cross-compile + Cosign em tag v*
└── go.mod
```

Justificativa: dois binários em `cmd/` compartilham `internal/*`. Os dois
projetos Java vivem no mesmo repositório para CI unificada — `build.yml`
roda `mvn test` nos dois.

---

## US-01.1 — Estrutura base do CLI em Go

### T-01.1.1 — Inicializar módulo
- `go mod init github.com/kyriosdata/runner`
- Branch padrão `main` no GitHub (DT-02)

### T-01.1.2 — Criar layout (DT-06)
- Diretórios `cmd/{assinatura,simulador}` e `internal/{cli,invoker,jdk,logging,release}`
- `assinador/` e `simuladorjar/` já populados pela base anterior

### T-01.1.3 — Comando `version` com SHA injetado
- Em [`cmd/assinatura/main.go`](../cmd/assinatura/main.go) declarar
  `var version = "dev"` e `var commit = "none"` como **variáveis** (não
  `const`, senão o `-ldflags` não sobrescreve)
- Subcomando `version` imprime `assinatura <tag> (<sha-curto>)`
- Mesma estrutura em [`cmd/simulador/main.go`](../cmd/simulador/main.go)

### T-01.1.4 — Parser com stdlib `flag`
- Sem dependência externa (ADR 0003)
- Subcomandos detectados por `os.Args[1]`, depois `flag.NewFlagSet(...)`
- Flags globais (`--verbose`, `--quiet`, `--modo`) parseadas antes do subcomando

### T-01.1.5 — Verificar compilação
- `go vet ./...` sem warnings
- `go build -o bin/assinatura ./cmd/assinatura`
- `go build -o bin/simulador  ./cmd/simulador`

### T-01.1.6 — Teste de aceitação do `version`
- [`cmd/assinatura/version_test.go`](../cmd/assinatura/version_test.go) usa
  `os/exec` para executar `go run .` no diretório do binário
- Verifica que a saída contém `"dev"` (valor padrão sem injeção)
- Mesma cobertura em `cmd/simulador/version_test.go`

---

## US-05.1 — Pipeline CI

### T-05.1.1 — Workflow `build.yml`
- Trigger: `push` e `pull_request` com `branches: [main]`
  (a restrição evita disparo em tag `v*`, que pertence ao `release.yml`)
- `actions/setup-go@v5` com `go-version: '1.21'`
- `actions/setup-java@v4` com `temurin` 17

### T-05.1.2 — Job `test` em matrix
- Runners: `ubuntu-latest`, `windows-latest`, `macos-latest`
- Etapas: `go vet ./...`, `go test ./...`, `cd assinador && mvn -B test`,
  `cd simuladorjar && mvn -B test`
- Não gera artefato; serve como gate

### T-05.1.3 — Job `build` (cross-compile)
- Roda só em `ubuntu-latest`
- `needs: test`
- Para cada `(GOOS, GOARCH)` de DT-03:
  `GOOS=<os> GOARCH=<arch> go build -o dist/<bin>-<os>-<arch>[.exe] ./cmd/<bin>`

### T-05.1.4 — `upload-artifact`
- Um artifact por plataforma, nomeado por DT-04 sem a tag
  (a tag entra no `release.yml`)

---

## US-05.2 — Releases SemVer

### T-05.2.1 — Workflow `release.yml`
- Trigger: `push` de tag `v*`
- Jobs em sequência: `test` → `build` → `publish`

### T-05.2.2 — Job `test` (3 SO)
- Mesma matrix do `build.yml`; release não avança se falhar

### T-05.2.3 — Job `build` com versão injetada
- Versão da tag via `${{ github.ref_name }}`
- `-ldflags "-X main.version=<tag> -X main.commit=<sha-curto>"` em cada compile
- Nomes finais: `assinatura-<tag>-<os>-<arch>[.exe]`, idem `simulador`
- Geração de `checksums.txt`:
  `sha256sum dist/* > dist/checksums.txt` (Linux/macOS) ou equivalente

### T-05.2.4 — Job `publish`
- `needs: build`
- `actions/download-artifact@v4` para coletar binários + `checksums.txt`
- `softprops/action-gh-release@v2` cria o release com os binários, os JARs
  do `assinador` e `simulador`, `checksums.txt`, `release.json` e os
  arquivos `.sig`/`.pem` gerados pelo Cosign
- Permissão `contents: write` + `id-token: write` (Cosign OIDC)

---

## US-05.3 — Cosign keyless

### T-05.3.1 — Instalar `cosign`
- `sigstore/cosign-installer@v3` no job `publish`

### T-05.3.2 — Assinar cada artefato
- Para cada binário e cada JAR:
  ```
  cosign sign-blob --yes \
    --output-signature <artefato>.sig \
    --output-certificate <artefato>.pem \
    <artefato>
  ```

### T-05.3.3 — Documentar verificação
- README publica o comando `cosign verify-blob` com
  `--certificate-identity-regexp` apontando para o workflow do fork

---

## Definition of Done atendida

- [x] `go build ./...` sem erros
- [x] `go vet ./...` sem warnings
- [x] `assinatura version` e `simulador version` exibem versão correta
- [x] `build.yml` verde em `push` para `main` nos 3 SO
- [x] `release.yml` publica binários, JARs, `checksums.txt`, `release.json`
      e arquivos `.sig`/`.pem` ao criar tag `v*`
- [x] `cosign verify-blob` reconhece os artefatos como autênticos
