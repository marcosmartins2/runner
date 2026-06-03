# Architecture Decision Records (ADRs)

Um **ADR** registra uma decisão arquitetural relevante, seu contexto e suas
consequências. São curtos (≈1 página), imutáveis depois de aceitos e numerados
sequencialmente. Quando uma decisão é revista, cria-se um novo ADR que
*supersede* o anterior (em vez de editá-lo).

Formato adotado (inspirado em Michael Nygard):
**Status · Contexto · Decisão · Consequências**.

| ADR | Título | Status |
|-----|--------|--------|
| [0001](0001-linguagem-go-para-as-clis.md) | Linguagem Go para as CLIs nativas | Aceito |
| [0002](0002-porta-padrao-e-descoberta-de-instancia.md) | Porta padrão e descoberta de instância por health check | Aceito |
| [0003](0003-parser-de-cli-com-stdlib-flag.md) | Parser de CLI com a biblioteca padrão `flag` | Aceito |
| [0004](0004-estrategia-hibrida-artefatos.md) | Estratégia híbrida de artefatos (JARs reais do upstream + fallback local) | Aceito |
| [0005](0005-integridade-sha256-e-cosign.md) | Integridade por SHA256 (runtime) e assinatura Cosign (release) | Aceito |
