# ADR 0001 — Linguagem Go para as CLIs nativas

**Status:** Aceito (2026-05)

## Contexto

A especificação exige binários nativos para Windows, Linux e macOS (US-05),
sem que o usuário precise instalar runtimes (US-04). O design inicial
(ver `design.md`, decisão DD-01 original) cogitava Python, que exigiria
empacotamento adicional (PyInstaller/venv) e um interpretador na máquina-alvo.

## Decisão

Implementar as duas CLIs (`assinatura` e `simulador`) em **Go**:

- compilação cruzada nativa para as três plataformas com um único toolchain;
- binário estático único, sem dependências de runtime no host;
- biblioteca padrão cobre HTTP, `os/exec` (subprocess) e `log/slog`
  (logs estruturados), evitando dependências externas — `go.mod` sem libs de
  terceiros.

## Consequências

- ✅ `go build` com `GOOS/GOARCH` cobre US-05 sem ferramentas extras.
- ✅ Distribuição trivial: um arquivo por plataforma, sem classpath/venv.
- ✅ Cadeia de dependências mínima (critério F de `docs/criterios.md`).
- ⚠️ A equipe assume o ecossistema Go (estilo, `go vet`, `gofmt`).
- ⚠️ Os artefatos Java (assinador/simulador `.jar`) continuam exigindo JRE,
  resolvido pelo provisionamento automático do Temurin (US-04).
