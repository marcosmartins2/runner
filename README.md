# Sistema Runner

Sistema para facilitar o acesso a funcionalidade de execucao de aplicacoes Java via linha de comandos.

## Documentacao

- [Especificacao](./especificacao.md) - escopo, requisitos (US-01 a US-05), entregaveis
- [Design (C4)](./design.md) - diagrama de contexto, de conteineres e decisoes de design
- [Planejamento e ambientacao](./docs/planejamento.md)
- [Diagramas PlantUML](./diagramas/)

## Status Atual

- Fluxo local implementado: `assinatura` (Python) invoca `assinador.jar` (Java)
- Modo servidor HTTP implementado para o `assinador.jar`, com start/status/stop pelo CLI
- Registro local de PID/porta do servidor em `.hubsaude/assinador-server.json`
- Parada programada por inatividade para o servidor do Assinador
- Validacao de parametros e simulacao deterministica no `assinador`
- CLI `simulador` (Python) com iniciar/parar/status do `simulador.jar`
- Pipeline CI configurado (testes Java e Python multi-OS)
- Especificacao, design C4 e diagramas registrados no repositorio
- Pendencias principais: download dinamico do `simulador.jar`, provisionamento de JDK, releases multiplataforma com Cosign, PKCS#11

## Estrutura do Projeto

```text
runner/
|-- .github/workflows/ci.yml  # Pipeline de testes (Java + Python)
|-- assinador/                # Aplicacao Java (assinador.jar)
|   |-- pom.xml
|   `-- src/
|       |-- main/java/com/runner/assinador/
|       |   |-- AssinadorApp.java
|       |   |-- AssinadorHttpServer.java
|       |   |-- AssinadorJson.java
|       |   |-- AssinaturaService.java
|       |   |-- ValidadorParametros.java
|       |   |-- ParametrosEntrada.java
|       |   `-- RespostaAssinatura.java
|       `-- test/java/com/runner/assinador/
|           `-- AssinadorAppTest.java
|-- assinatura/               # CLI Python (assinatura)
|   |-- assinatura.py
|   |-- test_assinatura.py
|   `-- requirements.txt
|-- simulador/                # CLI Python (simulador)
|   |-- simulador.py
|   |-- test_simulador.py
|   `-- requirements.txt
|-- diagramas/                # Diagramas C4 em PlantUML
|   |-- contexto.puml
|   `-- conteineres.puml
|-- docs/
|   `-- planejamento.md
|-- especificacao.md
|-- design.md
|-- geraimagens.bat           # Gera SVGs dos diagramas (Windows)
|-- geraimagens.sh            # Gera SVGs dos diagramas (Linux/Mac)
|-- LICENSE
`-- README.md
```

## Pre-requisitos

- Java: JDK 17 ou superior
- Maven: 3.8 ou superior
- Python: 3.10 ou superior

## Compilacao e Instalacao

### 1. Compilar o Assinador (Java)

```bash
cd assinador
mvn clean package
```

Isso gera o arquivo `assinador/target/assinador-1.0.0.jar`.

### 2. Dependencias do CLI (Python)

O CLI usa a biblioteca padrao do Python. Para ferramentas de teste opcionais:

```bash
cd assinatura
pip install -r requirements.txt
```

## Uso

### Criar assinatura digital (simulacao)

```bash
python assinatura/assinatura.py criar --documento SGVsbG8= --certificado cert-001
```

### Validar assinatura digital (simulacao)

```bash
python assinatura/assinatura.py validar --documento SGVsbG8= --assinatura <assinatura-gerada-para-este-documento>
```

Por padrao, o CLI usa o modo servidor: se o `assinador.jar` nao estiver em execucao na porta 8080, ele tenta inicia-lo automaticamente.

### Usar invocacao local direta

```bash
python assinatura/assinatura.py --modo local criar --documento SGVsbG8= --certificado cert-001
```

### Gerenciar o servidor do Assinador

```bash
python assinatura/assinatura.py servidor iniciar
python assinatura/assinatura.py servidor status
python assinatura/assinatura.py servidor parar
```

Para usar outra porta:

```bash
python assinatura/assinatura.py --porta 9090 servidor iniciar
```

Para encerrar automaticamente o servidor apos um periodo sem requisicoes:

```bash
python assinatura/assinatura.py servidor iniciar --parar-apos-minutos 30
```

Ao iniciar o servidor pela CLI, o Runner registra PID, porta, caminho do JAR e horario de inicio em:

```text
~/.hubsaude/assinador-server.json
```

### Verificar versao do CLI

```bash
python assinatura/assinatura.py --version
```

### Especificar caminho do JAR

```bash
python assinatura/assinatura.py --jar /caminho/para/assinador.jar criar --documento SGVsbG8= --certificado cert-001
```

## Exemplo de Saida

```text
============================================================
  [OK] Status: SUCESSO
------------------------------------------------------------
  Mensagem:    Assinatura digital criada com sucesso (simulacao).
  Algoritmo:   SHA256withRSA
  Certificado: cert-001
  Assinatura:  YWJjZGVmZzEyMzQ1Njc4OTBhYmNkZWZn...
  Timestamp:   2026-03-17T20:45:00Z
============================================================
```

## Testes

### Testes Java (assinador)

```bash
cd assinador
mvn test
```

### Testes Python (CLI assinatura)

```bash
cd assinatura
python -m unittest test_assinatura.py -v
```

### Testes Python (CLI simulador)

```bash
cd simulador
python -m unittest test_simulador.py -v
```

## CLI simulador

Comandos basicos para gerenciar o ciclo de vida do `simulador.jar`:

```bash
python simulador/simulador.py iniciar --jar /caminho/para/simulador.jar
python simulador/simulador.py status
python simulador/simulador.py parar
```

A porta padrao e 8443. Estado local (PID/porta/jar) e gravado em `~/.hubsaude/simulador-server.json`.

## Diagramas (C4)

Os diagramas estao em [diagramas/](./diagramas/) (PlantUML). Para gerar os SVGs (requer Java):

```bash
# Linux / macOS
./geraimagens.sh

# Windows
geraimagens.bat
```

Os SVGs sao gravados em `diagramas/imagens/` e referenciados em [design.md](./design.md).

## Arquitetura

O sistema segue o modelo C4 de documentacao arquitetural:

- CLI `assinatura` (Python): recebe comandos do usuario, invoca o `assinador.jar` via `subprocess` e formata a saida.
- `assinatura` tambem pode usar o `assinador.jar` em modo servidor HTTP para evitar reinicializacao da JVM a cada chamada.
- `assinador.jar` (Java): valida parametros rigorosamente, simula criacao/validacao de assinaturas digitais e retorna JSON.

### Fluxo

```text
Usuario -> assinatura (Python CLI) -> assinador.jar (Java) -> assinatura -> Usuario
```
