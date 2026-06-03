# Sistema Runner - Design

- O registro do design e organizado conforme o modelo C4. Consulte [C4 Model](https://c4model.com/) para detalhes.
- Diagramas usam PlantUml. Consulte [PlantUml](https://plantuml-documentation.readthedocs.io/en/latest/) para detalhes.
- Scripts `geraimagens.sh` e `geraimagens.bat` automatizam a geracao dos diagramas a partir dos arquivos `.puml`.

## 1. Diagrama de Contexto

![](diagramas/imagens/contexto.svg)

**Atores e Sistemas Externos:**

| Elemento | Tipo | Descricao |
|----------|------|-----------|
| Usuario | Ator | Pessoa que interage com o sistema via linha de comandos |
| Dispositivo de Assinatura Digital | Sistema Externo | Hardware criptografico (token USB, smart card) que armazena certificados e executa operacoes de assinatura |
| Simulador do HubSaude | Sistema Externo | Aplicacao Web gerida pelo CLI e que responde a requisicoes de terceiros |

## 2. Diagrama de Conteineres

![](diagramas/imagens/conteineres.svg)

**Comunicacao entre conteineres:**

| Origem | Destino | Protocolo | Descricao |
|--------|---------|-----------|-----------|
| Usuario | assinatura | CLI | Comandos de assinatura (criar, validar) digitados no terminal |
| Usuario | simulador | CLI | Comandos de gerenciamento do Simulador |
| assinatura | assinador.jar | linha de comandos ou HTTP | Invocacao direta (cold start) ou requisicao HTTP (warm start) |
| assinador.jar | Dispositivo Criptografico | PKCS#11 | Interface padrao para comunicacao com tokens e smart cards |
| simulador | Simulador do HubSaude | HTTP | Invoca e monitora o ciclo de vida do Simulador |

## 3. Componentes principais

### 3.1. assinatura (CLI Go)

- Recebe comandos do usuario (`criar`, `validar`, `servidor`, `provisionar-jdk`)
- Suporta dois modos de invocacao: local (subprocess) e servidor (HTTP)
- Detecta automaticamente se um servidor ja esta em execucao na porta alvo
  (health check em `/api/info`)
- Persiste metadados do servidor em `~/.hubsaude/assinador-server.json`
- Formata a resposta JSON do Assinador em saida amigavel
- Logs estruturados (`log/slog`) com `--verbose`/`--quiet`; resultado em stdout,
  diagnostico em stderr

### 3.2. assinador.jar (Java)

- Valida rigorosamente os parametros recebidos (`ValidadorParametros`)
- Simula criacao e validacao de assinaturas digitais (`AssinaturaService`)
- Retorna respostas como JSON (`RespostaAssinatura`)
- Pode operar como CLI de uma vez ou como servidor HTTP (`AssinadorHttpServer`)
- Suporta encerramento programado por inatividade

### 3.3. simulador (CLI Go)

- Gerencia o ciclo de vida do `simulador.jar` (start, stop, status, atualizar)
- Verifica disponibilidade da porta padrao (8443) antes de iniciar
- Consulta `/api/info` e aciona `/shutdown` do `simulador.jar`
- Baixa o `simulador.jar` dinamicamente via `release.json` (artefato real do
  upstream), com verificacao de integridade SHA256
- Provisiona o JRE via Eclipse Temurin (Adoptium) quando ausente

## 4. Decisoes de design

As decisoes nao-obvias sao registradas como ADRs curtos em
[`docs/adr/`](docs/adr/). Resumo das decisoes de design:

| ID | Decisao | Justificativa | ADR |
|----|---------|---------------|-----|
| DD-01 | CLIs em Go (binarios nativos) | Cross-compile sem runtime no host; stdlib cobre HTTP/subprocess/log | [0001](docs/adr/0001-linguagem-go-para-as-clis.md) |
| DD-02 | Assinador em Java 17 com Maven | Compatibilidade com o ecossistema do HubSaude | — |
| DD-03 | Servidor HTTP minimo (`com.sun.net.httpserver`) | Evita framework HTTP pesado, suficiente para o escopo de simulacao | — |
| DD-04 | Estado em `~/.hubsaude/assinador-server.json` | CLI descobre instancia em execucao entre invocacoes | [0002](docs/adr/0002-porta-padrao-e-descoberta-de-instancia.md) |
| DD-05 | Resposta sempre em JSON | Padronizacao para integracao CLI/HTTP e testes | — |
| DD-06 | Versionamento SemVer | Automacao de release e politicas de upgrade | — |
| DD-07 | Parser de CLI com stdlib `flag` | Escopo pequeno; zero dependencias | [0003](docs/adr/0003-parser-de-cli-com-stdlib-flag.md) |
| DD-08 | Artefatos reais do upstream + fallback local | Single source of truth (criterio B) | [0004](docs/adr/0004-estrategia-hibrida-artefatos.md) |
| DD-09 | Integridade SHA256 (runtime) + Cosign (release) | Seguranca da cadeia de suprimentos | [0005](docs/adr/0005-integridade-sha256-e-cosign.md) |
