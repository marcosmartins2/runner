# ADR 0003 — Parser de CLI com a biblioteca padrão `flag`

**Status:** Aceito (2026-05)

## Contexto

As CLIs têm poucos comandos (`criar`, `validar`, `servidor`, `provisionar-jdk`;
`iniciar`, `parar`, `status`, `atualizar`) e poucas flags por comando. Existem
frameworks populares (Cobra, urfave/cli), mas eles adicionam dependências e
peso de manutenção desproporcionais ao escopo.

## Decisão

Usar a biblioteca padrão **`flag`** com um `flag.FlagSet` por subcomando e um
`switch` sobre `args[0]` no entrypoint testável `executar(args, stdout, stderr)`.
As flags globais `--verbose`/`--quiet` são extraídas antes do `switch`
(`configurarObservabilidade`).

## Consequências

- ✅ Zero dependências externas (coerente com o ADR 0001 e o critério F).
- ✅ `executar` recebe `stdout`/`stderr` como `io.Writer`, tornando o parsing e
  os códigos de saída testáveis sem subprocesso (ver `*_test.go` das CLIs).
- ✅ Separação stdout (resultado) × stderr (diagnóstico) sob controle direto.
- ⚠️ Sem geração automática de `--help`/autocompletar; o texto de uso é mantido
  à mão em `imprimirUso` (com exemplos, conforme critério I).
- ⚠️ `flag` não suporta flags globais nativamente; daí a extração manual de
  `--verbose`/`--quiet`.
