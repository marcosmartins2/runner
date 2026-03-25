# Sistema Runner

Sistema para facilitar o acesso à funcionalidade de execução de aplicações Java via linha de comandos.

## Estrutura do Projeto

```
runner/
├── assinador/              # Aplicação Java (assinador.jar)
│   ├── pom.xml
│   └── src/
│       ├── main/java/com/runner/assinador/
│       │   ├── AssinadorApp.java         # Ponto de entrada
│       │   ├── AssinaturaService.java    # Lógica de negócio
│       │   ├── ValidadorParametros.java  # Validação de parâmetros
│       │   ├── ParametrosEntrada.java    # DTO de entrada
│       │   └── RespostaAssinatura.java   # DTO de resposta (JSON)
│       └── test/java/com/runner/assinador/
│           └── AssinadorAppTest.java     # Testes unitários
│
├── assinatura/             # CLI Python (interface do usuário)
│   ├── assinatura.py       # CLI principal
│   ├── test_assinatura.py  # Testes unitários
│   └── requirements.txt
│
└── README.md
```

## Pré-requisitos

- **Java**: JDK 17 ou superior
- **Maven**: 3.8 ou superior
- **Python**: 3.10 ou superior

## Compilação e Instalação

### 1. Compilar o Assinador (Java)

```bash
cd assinador
mvn clean package
```

Isso gera o arquivo `assinador/target/assinador-1.0.0.jar`.

### 2. Instalar dependências do CLI (Python)

```bash
cd assinatura
pip install -r requirements.txt
```

## Uso

### Criar Assinatura Digital (simulação)

```bash
python assinatura/assinatura.py criar --documento SGVsbG8= --certificado cert-001
```

### Validar Assinatura Digital (simulação)

```bash
python assinatura/assinatura.py validar --documento SGVsbG8= --assinatura dGVzdA==
```

### Verificar versão do CLI

```bash
python assinatura/assinatura.py --version
```

### Especificar caminho do JAR

```bash
python assinatura/assinatura.py --jar /caminho/para/assinador.jar criar --documento SGVsbG8= --certificado cert-001
```

## Exemplo de Saída

```
============================================================
  [OK] Status: SUCESSO
------------------------------------------------------------
  Mensagem:    Assinatura digital criada com sucesso (simulação).
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
python -m pytest test_assinatura.py -v
```

## Arquitetura

O sistema segue o modelo C4 de documentação arquitetural:

- **CLI `assinatura`** (Python): Recebe comandos do usuário, invoca o `assinador.jar` via `subprocess`, e formata a saída.
- **`assinador.jar`** (Java): Valida parâmetros rigorosamente, simula criação/validação de assinaturas digitais, e retorna JSON.

### Fluxo

```
Usuário → assinatura (Python CLI) → assinador.jar (Java) → assinatura → Usuário
```
