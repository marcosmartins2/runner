# Sistema Runner

Sistema para facilitar o acesso a funcionalidade de execucao de aplicacoes Java via linha de comandos.

## Documentacao de Planejamento

- [Planejamento e ambientacao](./docs/planejamento.md)

## Status Atual

- Fluxo local implementado: `assinatura` (Python) invoca `assinador.jar` (Java)
- Validacao de parametros e simulacao deterministica no `assinador`
- Planejamento de construcao registrado no repositorio
- Pendencias principais: modo servidor, `simulador`, provisionamento de JDK, CI/CD e releases multiplataforma

## Estrutura do Projeto

```text
runner/
|-- assinador/              # Aplicacao Java (assinador.jar)
|   |-- pom.xml
|   `-- src/
|       |-- main/java/com/runner/assinador/
|       |   |-- AssinadorApp.java
|       |   |-- AssinaturaService.java
|       |   |-- ValidadorParametros.java
|       |   |-- ParametrosEntrada.java
|       |   `-- RespostaAssinatura.java
|       `-- test/java/com/runner/assinador/
|           `-- AssinadorAppTest.java
|-- assinatura/             # CLI Python
|   |-- assinatura.py
|   |-- test_assinatura.py
|   `-- requirements.txt
|-- docs/
|   `-- planejamento.md
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

### Testes Python (CLI)

```bash
cd assinatura
python -m unittest test_assinatura.py -v
```

## Arquitetura

O sistema segue o modelo C4 de documentacao arquitetural:

- CLI `assinatura` (Python): recebe comandos do usuario, invoca o `assinador.jar` via `subprocess` e formata a saida.
- `assinador.jar` (Java): valida parametros rigorosamente, simula criacao/validacao de assinaturas digitais e retorna JSON.

### Fluxo

```text
Usuario -> assinatura (Python CLI) -> assinador.jar (Java) -> assinatura -> Usuario
```
