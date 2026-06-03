# ADR 0004 — Estratégia híbrida de artefatos (JARs reais do upstream + fallback local)

**Status:** Aceito (2026-06)

## Contexto

O repositório da especificação (upstream `kyriosdata/runner`) passou a publicar
os artefatos **reais** e a apontá-los no `release.json`:

- `validador` → `hubsaude-validador-api` (o "assinador.jar" real);
- `simulador` → `hubsaude-simulador`;
- com `tag` fixa e `sha256` por artefato, além de alvos de JRE arm64.

O critério **B (Single source of truth)** de `docs/criterios.md` pede para não
duplicar o que pertence ao upstream. Ao mesmo tempo, mantemos projetos Java
próprios (`assinador/`, `simuladorjar/`) úteis para desenvolvimento e testes
offline.

## Decisão

Adotar uma estratégia **híbrida**:

1. **Em runtime, por padrão**, o CLI consome os artefatos reais do upstream.
   O manifesto é buscado em
   `https://raw.githubusercontent.com/kyriosdata/runner/main/release.json`
   e os JARs são baixados sob demanda, com verificação de integridade
   (ver ADR 0005). O campo `validador` tem precedência sobre `jar`
   (`Manifesto.ArtefatoAssinador`).
2. **Fallback local de desenvolvimento:** se houver um `.jar` local
   (`--jar`, `ASSINADOR_JAR`, `*/target/`, `~/.hubsaude/`), ele é usado sem rede.
3. Os projetos Java próprios permanecem no repositório **apenas** como
   implementação de referência/fallback para dev e CI; não são a fonte da
   verdade dos artefatos distribuídos.
4. O `release.json` versionado referencia o upstream por **tag fixa + sha256**;
   o pipeline de release não reescreve essas versões.

## Consequências

- ✅ Alinha-se ao critério B sem perder a capacidade de build/teste offline.
- ✅ O usuário final obtém os artefatos oficiais, verificados.
- ✅ Chaves de JRE seguem a convenção upstream (`mac_arm64`, `linux_arm64`, …),
  corrigindo a incompatibilidade que quebrava o provisionamento em Apple Silicon.
- ⚠️ Há acoplamento à disponibilidade do upstream em runtime (mitigado pelo
  fallback local).
- ⚠️ Manter dois JARs (nosso e o real) pode confundir; documentado aqui e no
  README para evitar ambiguidade.
