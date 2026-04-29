# Planejamento de construcao - runner

## 1. Base consultada em 2026-04-07

Este planejamento foi produzido a partir de tres referencias:

- solicitacao de atualizacao do repositorio, com foco em consultar o conteudo atual do portal e acrescentar o planejamento ao projeto;
- repositorio oficial `kyriosdata/runner`, cujo `README.md` informa que, desde 2026-03-18, o foco passou a ser o planejamento da construcao;
- documentos `docs/planejamento.md` e `docs/plano-revisitado-v2.md` do projeto oficial, consultados na branch `main`, com release mais recente `v0.0.6` publicada em 2026-04-01.

Referencias:

- <https://github.com/kyriosdata/runner>
- <https://github.com/kyriosdata/runner/blob/main/docs/planejamento.md>
- <https://github.com/kyriosdata/runner/blob/main/docs/plano-revisitado-v2.md>

## 2. Diretrizes consideradas

Da documentacao mais recente do projeto de referencia, ficam claras quatro diretrizes para o repositorio:

1. registrar explicitamente a ambientacao adotada;
2. assumir um processo iterativo e incremental;
3. planejar a construcao em etapas que contemplem design detalhado, implementacao, testes de unidade, revisao e refatoracao;
4. manter rastreabilidade entre o que ja existe no projeto e o que ainda falta para atender a especificacao.

## 3. Diagnostico do estado atual do repositorio

### O que ja existe

- CLI `assinatura` em Python com comandos `criar` e `validar`, opcao `--version` e suporte a `--jar`;
- aplicacao `assinador` em Java com validacao de parametros, simulacao de assinatura e resposta JSON;
- testes unitarios em Java para validacao de parametros e servico de assinatura;
- testes do CLI em Python escritos com `unittest`.

### O que motivou esta atualizacao

- o repositorio nao tinha um documento de planejamento proprio;
- a ambientacao tecnica nao estava registrada;
- artefatos gerados (`target/`, `__pycache__/`) estavam sendo versionados, o que conflita com um fluxo minimo de construcao e revisao.

### Lacunas em relacao ao plano oficial

| Item | Situacao no repositorio atual | Observacao |
|---|---|---|
| US-02.1, US-02.2 e parte de US-02.3 | parcialmente atendidos | o `assinador` local ja simula criacao/validacao e valida parametros |
| US-01.2, US-01.3 e US-01.4 | parcialmente atendidos | o CLI local ja parseia comandos, invoca o jar e formata a saida |
| US-04.1 | nao iniciado | nao ha provisionamento automatico do JDK |
| US-01.5 a US-01.9 | parcialmente atendido | ha modo servidor HTTP para o assinador, start/status/stop e uso automatico pelo CLI; ainda falta parada programada por inatividade |
| US-03.1 a US-03.4 | nao iniciado | nao ha CLI `simulador` nem download dinamico do `simulador.jar` |
| US-05.1 a US-05.3 | nao iniciado | nao ha CI/CD, release, checksum ou Cosign |

## 4. Ambientacao adotada

### Decisoes tecnicas

| ID | Decisao | Justificativa |
|---|---|---|
| DT-01 | Manter o CLI `assinatura` em Python na proxima iteracao | ja existe fluxo local funcional; a prioridade imediata e consolidar valor entregue e testes |
| DT-02 | Manter o `assinador` em Java 17 com Maven | o modulo ja esta operacional e testado |
| DT-03 | Adotar processo iterativo e incremental | alinhado a documentacao de referencia |
| DT-04 | Versionar somente codigo-fonte e documentacao | reduz ruido de build e melhora revisao |
| DT-05 | Usar `main` como branch principal e branches curtas por funcionalidade | alinhado a um fluxo simples de integracao e revisao |

Inferencia local: o plano oficial de referencia usa Go para os CLIs por facilitar distribuicao multiplataforma. Neste repositorio, a decisao para a proxima iteracao e manter Python e reavaliar migracao apenas se a entrega de US-05 se mostrar inviavel com a base atual.

### Ambiente minimo de desenvolvimento

- Python 3.10+
- Java 17+
- Maven 3.8+

### Comandos de trabalho

```bash
cd assinador
mvn test
```

```bash
cd assinatura
python -m unittest test_assinatura.py -v
```

```bash
cd assinador
mvn package
```

## 5. Modelo de construcao adotado

Cada iteracao deve seguir este ciclo:

1. design detalhado do incremento;
2. implementacao;
3. criacao ou ajuste de testes de unidade/integracao;
4. revisao;
5. refatoracao.

## 6. Planejamento das proximas iteracoes

### Iteracao 1 - consolidacao do modo local

Objetivo: deixar o fluxo local ja existente coerente com a documentacao do projeto, reproduzivel e rastreavel.

Historias foco:

- US-02.1
- US-02.2
- US-02.3
- US-01.2
- US-01.3
- US-01.4

Tarefas operacionais:

- [x] registrar a ambientacao e o planejamento no repositorio;
- [x] parar de versionar artefatos gerados e registrar `.gitignore`;
- [ ] revisar o CLI para incluir teste de integracao `assinatura -> assinador.jar`;
- [ ] revisar mensagens de ajuda/erro para cobrir os cenarios principais do fluxo local;
- [ ] padronizar a execucao dos testes do CLI em ambiente controlado.

Definition of Done da iteracao:

- `mvn test` passa no modulo `assinador`;
- testes do CLI executam sem depender de arquivos gerados previamente;
- o `README` aponta para o planejamento;
- o repositorio nao depende de `target/` ou `__pycache__/` versionados.

### Iteracao 2 - automacao e distribuicao

Objetivo: preparar o projeto para validacao continua e distribuicao.

Historias foco:

- US-04.1
- US-05.1
- US-05.2
- US-05.3

Tarefas operacionais:

- [ ] definir estrategia de empacotamento multiplataforma do CLI atual;
- [ ] criar pipeline para testes de Java e Python;
- [ ] gerar artefatos versionados para release;
- [ ] publicar checksums SHA-256;
- [ ] investigar automacao de assinatura com Cosign.

Risco principal:

- se o empacotamento do CLI em Python introduzir friccao excessiva para releases multiplataforma, a migracao para Go deve ser reaberta antes do fim desta iteracao.

### Iteracao 3 - modo servidor do assinador

Objetivo: reduzir latencia e permitir gerenciamento do processo do `assinador.jar`.

Historias foco:

- US-02.4
- US-01.5
- US-01.6
- US-01.7
- US-01.8
- US-01.9

Tarefas operacionais:

- [x] definir endpoints HTTP do `assinador`;
- [x] implementar inicializacao em background;
- [ ] registrar PID/porta em area de trabalho do usuario;
- [x] criar comandos de start/stop/status ou equivalente no CLI;
- [x] criar selecao explicita entre modo servidor e modo local.
- [ ] implementar interrupcao programada por inatividade.

### Iteracao 4 - simulador do HubSaude

Objetivo: completar o escopo de gerenciamento do `simulador.jar`.

Historias foco:

- US-03.1
- US-03.2
- US-03.3
- US-03.4

Tarefas operacionais:

- [ ] definir a interface do CLI `simulador`;
- [ ] documentar portas, startup e encerramento do simulador;
- [ ] baixar dinamicamente a versao mais recente do `simulador.jar`;
- [ ] suportar `--source` para fonte alternativa;
- [ ] registrar status local e checagem de integridade do download.

## 7. Resultado desta atualizacao no repositorio

Esta atualizacao acrescenta o planejamento ao repositorio, registra a ambientacao adotada e deixa de tratar artefatos gerados como parte do codigo-fonte.
