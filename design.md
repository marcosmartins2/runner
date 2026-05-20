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

### 3.1. assinatura (CLI Python)

- Recebe comandos do usuario (`criar`, `validar`, `servidor`)
- Valida entrada antes de invocar o Assinador
- Suporta dois modos de invocacao: local (subprocess) e servidor (HTTP)
- Detecta automaticamente se um servidor ja esta em execucao na porta alvo
- Persiste metadados do servidor em `~/.hubsaude/assinador-server.json`
- Formata a resposta JSON do Assinador em saida amigavel

### 3.2. assinador.jar (Java)

- Valida rigorosamente os parametros recebidos (`ValidadorParametros`)
- Simula criacao e validacao de assinaturas digitais (`AssinaturaService`)
- Retorna respostas como JSON (`RespostaAssinatura`)
- Pode operar como CLI de uma vez ou como servidor HTTP (`AssinadorHttpServer`)
- Suporta encerramento programado por inatividade

### 3.3. simulador (CLI Python)

- Gerencia o ciclo de vida do `simulador.jar` (start, stop, status)
- Verifica disponibilidade da porta padrao (8443) antes de iniciar
- Consulta `/api/info` e aciona `/shutdown` do `simulador.jar`
- (Planejado) Baixa o `simulador.jar` dinamicamente via GitHub Releases
- (Planejado) Provisiona o JRE via Eclipse Temurin (Adoptium)

## 4. Decisoes de design

| ID | Decisao | Justificativa |
|----|---------|---------------|
| DD-01 | CLI em Python | Disponivel em todas as plataformas, baixo overhead inicial, biblioteca padrao cobre HTTP/subprocess |
| DD-02 | Assinador em Java 17 com Maven | Compatibilidade com o ecossistema do HubSaude e com bibliotecas de criptografia |
| DD-03 | Servidor HTTP minimo (`com.sun.net.httpserver`) | Evita dependencia de framework HTTP pesado, suficiente para escopo de simulacao |
| DD-04 | Persistir estado em `~/.hubsaude/assinador-server.json` | Permite ao CLI descobrir instancia ja em execucao entre invocacoes |
| DD-05 | Resposta sempre em JSON | Padronizacao para integracao com CLI/HTTP e testes automatizados |
| DD-06 | Versionamento SemVer | Compatibilidade com automacao de release e politicas de upgrade dos integradores |
