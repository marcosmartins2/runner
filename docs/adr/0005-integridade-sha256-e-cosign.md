# ADR 0005 — Integridade por SHA256 (runtime) e assinatura Cosign (release)

**Status:** Aceito (2026-06)

## Contexto

O sistema baixa artefatos executáveis (JARs e JRE) e distribui binários. A
seção 9 da especificação e o critério **F (supply chain)** de
`docs/criterios.md` exigem proteção contra adulteração tanto no consumo quanto
na distribuição.

## Decisão

Defesa em duas camadas:

1. **No consumo (runtime):** quando o `release.json` traz `sha256`, o download
   é verificado antes de promover o arquivo (`BaixarArquivoVerificado` →
   `VerificarSHA256` em `internal/release/release.go`). Em divergência, o
   arquivo `.part` é descartado e o erro é explícito. A comparação é
   *case-insensitive* e o arquivo só é renomeado para o destino final após
   passar na verificação (download atômico via `.part`).
2. **Na distribuição (release):** o pipeline assina **todos** os artefatos com
   **Cosign keyless (OIDC/Sigstore)** e publica `<artefato>.sig` + `<artefato>.pem`
   e um `checksums.txt`, conforme a seção 9 da especificação
   (`.github/workflows/release.yml`).

O JRE do Adoptium é baixado por URL "latest" e não traz `sha256` no manifesto;
nesse caso a verificação por digest é omitida (limitação documentada).

## Consequências

- ✅ Adulteração de JAR é detectada em runtime; artefatos de release são
  verificáveis por qualquer usuário com `cosign verify-blob`.
- ✅ Coberto por testes (`TestBaixarArquivoVerificadoSHA256OK`,
  `TestBaixarArquivoVerificadoSHA256Divergente`, `TestVerificarSHA256`).
- ⚠️ O JRE permanece sem verificação por digest enquanto o manifesto não
  fornecer `sha256` para ele.
