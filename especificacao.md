# Sistema Runner - Especificacao

## 1. Visao geral

Este documento define o escopo, objetivos e requisitos do *Sistema Runner*, util para integradores da Plataforma HubSaude. O Runner facilita o acesso a aplicacoes Java envolvidas no fluxo de assinatura digital e no gerenciamento do Simulador do HubSaude, sem expor ao usuario detalhes de configuracao do ambiente Java.

A especificacao aqui apresentada e uma adaptacao do trabalho pratico da disciplina de Implementacao e Integracao (Bacharelado em Engenharia de Software, 2026-01).

## 2. Objetivo do Sistema Runner

Facilitar o acesso a funcionalidade de execucao de aplicacoes Java via linha de comandos.

## 3. Objetivos especificos

1. Permitir que os usuarios possam executar aplicacoes Java sem necessidade de conhecer detalhes de configuracao ou instalacao do ambiente Java. Em particular, as aplicacoes que fazem parte do Sistema Runner.

2. Fornecer uma interface de linha de comandos (CLI) simples e intuitiva para interacao com as aplicacoes Java, permitindo que os usuarios possam executar comandos especificos para cada aplicacao. Desta forma, ocultando a complexidade de configuracao e facilitando o acesso as funcionalidades sem necessidade de conhecimento tecnico aprofundado.

## 4. Escopo

### 4.1. O que ESTA no Escopo

- Desenvolvimento da aplicacao `assinatura` (CLI multiplataforma)
- Desenvolvimento da aplicacao `assinador.jar` (Java)
- Integracao entre as duas aplicacoes
- Validacao rigorosa de parametros pelo `assinador.jar`
- Simulacao de criacao de assinatura (`assinador.jar`)
- Simulacao de validacao de assinatura (`assinador.jar`)
- Tratamento de erros dos parametros e excecoes (`assinador.jar`)
- Desenvolvimento da aplicacao `simulador` (CLI multiplataforma)
- Testes
- Documentacao de uso

### 4.2. O que NAO ESTA no Escopo

- Implementacao real de assinatura digital criptografica
- Implementacao real de validacao de assinatura digital criptografica
- Integracao com autoridades certificadoras
- Armazenamento persistente de assinaturas
- Interface grafica (GUI)
- Autenticacao de usuarios
- Geracao de certificados digitais

## 5. Requisitos funcionais

Os requisitos sao expressos como historias de usuario (user stories).

### US-01: Invocar assinador.jar via CLI

**Como** usuario do Sistema Runner
**Quero** executar comandos de assinatura digital atraves da linha de comandos
**Para que** eu possa invocar a aplicacao `assinador.jar` (doravante, Assinador) sem conhecer os detalhes tecnicos de configuracao Java, tanto para assinar quanto para validar assinaturas digitais.

**Criterios de aceitacao:**
- [x] O CLI deve aceitar comandos para criacao e validacao de assinatura
- [x] O CLI deve invocar o `assinador.jar` com os parametros fornecidos
- [x] O CLI deve permitir a invocacao direta do `assinador.jar` (modo local/CLI)
- [x] O CLI deve permitir a invocacao do `assinador.jar` via HTTP (modo servidor)
- [x] O CLI deve exibir o resultado da operacao de forma legivel
- [x] O CLI deve iniciar o `assinador.jar` no modo servidor usando a porta padrao quando nao orientado de forma diferente
- [x] O CLI deve detectar se instancia do `assinador.jar` ja esta em execucao no modo servidor e usar essa instancia
- [x] O CLI deve fazer uso do `assinador.jar` no modo servidor quando nao orientado para usar o modo local
- [x] O CLI deve interromper a execucao do `assinador.jar` na porta padrao ou outra indicada
- [x] O CLI deve permitir a requisicao de interrupcao programada do `assinador.jar` apos N minutos sem interacao

**Modos de invocacao do Assinador:**

- **Invocacao direta (CLI)**: o CLI invoca o `assinador.jar` diretamente. Cada execucao realiza o ciclo completo de inicializacao da JVM e carga da aplicacao (*cold start*), adequado para execucoes esporadicas ou scripts de automacao.
- **Invocacao via HTTP (servidor)**: o `assinador.jar` permanece em execucao, aguardando requisicoes HTTP. O CLI envia requisicoes ao Assinador neste modo, eliminando o overhead de inicializacao (*warm start*) - menor latencia e maior throughput.

### US-02: Simular assinatura digital com validacao de parametros

**Como** usuario do Sistema Runner
**Quero** que o `assinador.jar` valide rigorosamente os parametros de entrada antes de simular uma operacao de assinatura digital
**Para que** eu receba feedback imediato sobre erros de parametros, garantindo que apenas requisicoes bem formadas sejam aceitas.

**Criterios de aceitacao:**
- [x] O `assinador.jar` deve validar todos os parametros conforme especificacoes
- [x] O `assinador.jar` deve simular a criacao de assinatura retornando resposta pre-construida quando parametros validos
- [x] O `assinador.jar` deve simular validacao de assinatura retornando resultado pre-determinado
- [ ] O `assinador.jar` deve suportar interacao com dispositivo criptografico (token/smart card) via interface PKCS#11
- [x] O `assinador.jar` deve retornar mensagens de erro claras quando parametros forem invalidos

### US-03: Gerenciar Ciclo de Vida do Simulador do HubSaude

**Como** usuario do Sistema Runner
**Quero** iniciar, parar e monitorar o Simulador do HubSaude (`simulador.jar`) atraves do CLI
**Para que** eu possa gerenciar o ciclo de vida do Simulador sem conhecer os comandos Java subjacentes.

**Criterios de aceitacao:**
- [ ] O CLI deve verificar se a porta padrao, 8443, esta disponivel antes de tentar iniciar o Simulador
- [ ] O CLI deve permitir iniciar o Simulador
- [ ] O CLI deve permitir parar o Simulador (endpoint `/shutdown`)
- [ ] O CLI deve exibir o status atual do Simulador (via `/api/info`)
- [ ] O `simulador.jar` deve ser obtido dinamicamente pelo CLI a partir do GitHub Releases
- [ ] O CLI deve baixar o JRE caso nao esteja disponivel no diretorio `.hubsaude` a partir do Eclipse Temurin (Adoptium)
- [ ] O CLI nao deve baixar o `simulador.jar` se a versao mais recente ja estiver disponivel localmente

Estrategia de descoberta de versao via `release.json`:

```json
{
  "jar": {
    "url": "https://github.com/<owner>/<repo>/releases/latest/download/simulador.jar",
    "version": "1.0.0"
  },
  "jre": {
    "windows_x64": "https://api.adoptium.net/v3/binary/latest/21/ga/windows/x64/jre/hotspot/normal/eclipse",
    "linux_x64":   "https://api.adoptium.net/v3/binary/latest/21/ga/linux/x64/jre/hotspot/normal/eclipse",
    "mac_x64":     "https://api.adoptium.net/v3/binary/latest/21/ga/mac/x64/jre/hotspot/normal/eclipse"
  }
}
```

### US-04: Provisionar JDK automaticamente

**Como** usuario do Sistema Runner
**Quero** que o sistema baixe e configure automaticamente o JDK necessario quando este nao estiver disponivel
**Para que** eu possa utilizar o Assinador e o Simulador sem precisar instalar ou configurar o Java manualmente.

**Criterios de aceitacao:**
- [ ] O sistema deve detectar se o JDK esta presente na maquina (na versao exigida)
- [ ] O sistema deve baixar o JDK compativel quando ausente
- [ ] O sistema deve disponibilizar o JDK baixado para uso pelo Assinador e Simulador
- [ ] O download deve funcionar nas tres plataformas (Windows, Linux, macOS)

### US-05: Disponibilizar binarios multiplataforma

**Como** usuario do Sistema Runner
**Quero** baixar uma versao pre-compilada do CLI para minha plataforma (Windows, Linux ou macOS)
**Para que** eu possa utilizar o sistema imediatamente sem necessidade de compilacao.

**Criterios de aceitacao:**
- [ ] Disponibilizar binario para Windows (amd64)
- [ ] Disponibilizar binario para Linux (amd64)
- [ ] Disponibilizar binario para macOS (amd64)
- [ ] Distribuir via GitHub Releases
- [ ] Incluir checksums SHA256 para verificacao de integridade
- [ ] Utilizar versionamento semantico (SemVer)

## 6. Integracao entre aplicacoes

A aplicacao `assinatura` (CLI) se comunica com o `assinador.jar` por dois mecanismos:

- **Invocacao direta**: `assinatura` executa `assinador.jar` via linha de comandos (`java -jar assinador.jar ...`)
- **Invocacao via HTTP**: `assinatura` envia requisicoes HTTP para o `assinador.jar` em execucao como servidor

O fluxo logico e o mesmo em ambos os modos, diferindo apenas no mecanismo de comunicacao.

### 6.1. Fluxo de criacao de assinatura

```
Usuario -> assinatura -> assinador.jar -> assinatura -> Usuario

1. Usuario: executa comando para criar assinatura
2. assinatura: valida entrada do usuario
3. assinatura: invoca assinador.jar (diretamente ou via HTTP)
4. assinador.jar: valida parametros
5. assinador.jar: retorna assinatura simulada
6. assinatura: formata resultado
7. assinatura: apresenta ao usuario
```

### 6.2. Fluxo de validacao de assinatura

```
Usuario -> assinatura -> assinador.jar -> assinatura -> Usuario

1. Usuario: executa comando para validar assinatura
2. assinatura: valida entrada do usuario
3. assinatura: invoca assinador.jar (diretamente ou via HTTP)
4. assinador.jar: valida parametros
5. assinador.jar: retorna resultado simulado
6. assinatura: formata resultado
7. assinatura: apresenta ao usuario
```

### 6.3. Tratamento de erros

Em qualquer ponto do fluxo, erros devem ser:

- Capturados apropriadamente
- Propagados de forma estruturada
- Apresentados ao usuario de forma clara
- Incluir informacao suficiente para correcao

## 7. Entregaveis

1. **Codigo-fonte da aplicacao `assinatura`** - implementacao completa, multiplataforma, bem documentado
2. **Codigo-fonte da aplicacao `assinador.jar`** - implementacao em Java, validacao completa, simulacao das operacoes
3. **Testes** - unitarios, integracao, cenarios de erro, aceitacao
4. **Documentacao** - manual do usuario, documentacao tecnica da integracao, exemplos de uso, guia de instalacao
5. **Especificacao** (este documento) - contexto, escopo, diagramas C4, requisitos
6. **Artefatos executaveis** - binarios pre-compilados para as tres plataformas, distribuidos via GitHub Releases
7. **Codigo-fonte do Simulador do HubSaude** - implementacao completa, multiplataforma, bem documentado

## 8. Consideracoes de implementacao

### 8.1. Simulacao

Como o sistema *simula* operacoes de assinatura digital:

- **Para criacao**: assinaturas pre-construidas/derivadas que podem ser retornadas quando os parametros sao validos
- **Para validacao**: logica simples que retorna um resultado determinado (valido/invalido) baseado em criterios simples
- **Foco na validacao**: a maior parte do esforco deve estar em validar corretamente os parametros de entrada

### 8.2. Padroes de qualidade

- Codigo limpo e bem organizado
- Tratamento adequado de excecoes
- Testes com boa cobertura
- Documentacao clara
- Mensagens de erro uteis

## 9. Integridade e assinatura de artefatos

Para garantir a autenticidade e a integridade dos binarios distribuidos, os artefatos publicados nas releases do projeto devem ser assinados criptograficamente utilizando **Cosign**, parte do ecossistema **Sigstore**.

Esse mecanismo permite que qualquer usuario verifique de forma independente a origem e a integridade dos artefatos distribuidos, reduzindo riscos de ataques a cadeia de suprimentos de software.

### 9.1. Arquivos por artefato

Para cada artefato distribuido:

```
<artefato>
<artefato>.sig
<artefato>.pem
<artefato>.sha256
```

### 9.2. Verificacao

```bash
cosign verify-blob \
  --certificate assinatura-1.0.0-linux-amd64.pem \
  --signature assinatura-1.0.0-linux-amd64.sig \
  assinatura-1.0.0-linux-amd64
```

## 10. Referencias

1. **Especificacoes FHIR - Seguranca**
   - [Caso de Uso: Criar Assinatura](https://fhir.saude.go.gov.br/r4/seguranca/caso-de-uso-criar-assinatura.html)
   - [Caso de Uso: Validar Assinatura](https://fhir.saude.go.gov.br/r4/seguranca/caso-de-uso-validar-assinatura.html)

2. **Modelo C4 para Visualizacao de Arquitetura**
   - [C4 Model](https://c4model.com/)
   - Nivel 1: [Diagrama de Contexto](./diagramas/imagens/contexto.svg)
   - Nivel 2: [Diagrama de Conteiner](./diagramas/imagens/conteineres.svg)

3. **Boas praticas de CLI**
   - Mensagens claras e consistentes
   - Tratamento adequado de erros
   - Documentacao de help integrada
