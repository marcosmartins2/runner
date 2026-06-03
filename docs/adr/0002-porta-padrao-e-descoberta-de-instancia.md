# ADR 0002 — Porta padrão e descoberta de instância por health check

**Status:** Aceito (2026-05)

## Contexto

No modo servidor (US-01), o CLI deve reutilizar uma instância já em execução
em vez de subir outra, e deve falhar de forma clara quando a porta estiver
tomada por **outro** processo. "Porta ocupada" não é o mesmo que "Assinador
vivo": um processo qualquer pode estar escutando a porta.

## Decisão

- **Portas padrão:** `8080` para o Assinador e `8443` para o Simulador
  (este último fixado pela especificação, US-03). Ambas configuráveis via
  `--porta`.
- **Descoberta de instância:** antes de iniciar, o CLI faz um *health check*
  real — `GET /api/info` — e só reutiliza a instância se a resposta indicar
  prontidão (`status` de sucesso). Ver `internal/invoker/invoker.go`
  (`EmExecucao`, `IniciarServidor`).
- **Conflito de porta:** se não há instância viva mas a porta está ocupada
  (`net.Listen` falha), o início aborta com erro explícito
  ("porta N já em uso por outro processo").

## Consequências

- ✅ Idempotência de start e mensagens de erro acionáveis (critérios E2 e
  "Falhar bem").
- ✅ Distinção entre "porta ocupada" e "Assinador pronto" comprovada por teste
  (`TestIniciarServidorReutilizaInstanciaViva`, `TestIniciarServidorPortaOcupada`).
- ⚠️ Portas padrão são constantes no código; mudá-las exige `--porta`. Não há,
  por ora, configuração por arquivo/variável de ambiente.
