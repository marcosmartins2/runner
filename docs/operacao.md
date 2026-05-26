# Operacao do Assinador em modo servidor

Este documento descreve como operar o `assinador.jar` em modo
servidor HTTP usando o CLI nativo `assinatura`. Cobre os ciclos de
inicializacao, consulta, encerramento manual e encerramento
programado por inatividade.

## 1. Pre-requisitos

- `assinador.jar` baixado da release ou empacotado em
  `assinador/target/` (gerado por `mvn package` no diretorio
  `assinador/`).
- Java 17+ disponivel via `JAVA_HOME` ou no `PATH`. Quando ausente,
  o CLI `assinatura provisionar-jdk` baixa um JRE Temurin para
  `~/.hubsaude/jre/`.

## 2. Ciclo de vida pelo CLI

### 2.1. Iniciar o servidor

```bash
assinatura servidor iniciar
```

Ao iniciar, o CLI:

1. verifica se ja existe um servidor escutando em `127.0.0.1:8080`;
2. se nao houver, dispara o `assinador.jar` em background usando
   `--server --port 8080`;
3. aguarda ate o endpoint `/api/info` responder com sucesso;
4. registra os metadados do processo em
   `~/.hubsaude/assinador-server.json`.

Para usar uma porta diferente:

```bash
assinatura servidor iniciar --porta 9090
```

### 2.2. Consultar o status

```bash
assinatura servidor status
```

A saida exibe o resultado de `/api/info` e, quando ha registro
local, mostra tambem PID e porta gravados em
`~/.hubsaude/assinador-server.json`.

### 2.3. Parar o servidor

```bash
assinatura servidor parar
```

O CLI envia `POST /shutdown` e, em seguida, remove o arquivo de
estado local.

## 3. Encerramento programado por inatividade

O servidor pode encerrar automaticamente apos um intervalo sem
requisicoes. Use `--parar-apos-minutos` ao inicia-lo:

```bash
assinatura servidor iniciar --parar-apos-minutos 30
```

Com a configuracao acima, o `assinador.jar` encerra apos 30
minutos sem requisicoes nos endpoints `/api/sign`, `/api/validate`
ou `/api/info`. O CLI nao precisa estar em execucao para que o
encerramento aconteca: a contagem e feita dentro do processo Java.

Quando o `assinador.jar` finaliza por inatividade, o arquivo
`~/.hubsaude/assinador-server.json` permanece no disco com os
metadados da ultima execucao. Uma chamada subsequente a
`servidor iniciar` detecta a porta livre e reescreve o arquivo.

## 4. Uso transparente via `criar`/`validar`

Os comandos `criar` e `validar` selecionam automaticamente o modo
servidor:

```bash
assinatura criar --documento SGVsbG8= --certificado cert-001
```

Se o servidor nao estiver em execucao, ele e iniciado sob demanda
antes da primeira chamada e mantido vivo para chamadas
subsequentes. Para forcar invocacao direta (cold start), use
`--modo local`:

```bash
assinatura --modo local criar --documento SGVsbG8= --certificado cert-001
```

## 5. Diagnostico

| Sintoma | Verificacao | Acao |
|---|---|---|
| `Assinador HTTP nao respondeu na porta N` | porta ocupada por outro processo, ou JVM lenta para subir | escolher outra porta via `--porta` ou aguardar e tentar novamente |
| `java nao encontrado: configure JAVA_HOME ou execute 'assinatura provisionar-jdk'` | `JAVA_HOME` nao definido e `java` ausente do `PATH` | rodar `assinatura provisionar-jdk` ou instalar JDK 17+ manualmente |
| `assinador.jar nao encontrado em ...` | binario ausente | passar `--jar PATH` ou copiar o jar para `~/.hubsaude/assinador.jar` |
| Status mostra dados antigos apos kill manual | arquivo `~/.hubsaude/assinador-server.json` ficou orfao | apagar o arquivo ou rodar `servidor iniciar` novamente para sobrescrever |

## 6. Arquivos de estado

O CLI grava em `~/.hubsaude/assinador-server.json`:

```json
{
  "pid": 12345,
  "porta": 8080,
  "jar": "/caminho/absoluto/para/assinador.jar",
  "iniciadoEm": "2026-05-26T22:00:00Z"
}
```

Esse arquivo e gravado a cada execucao de `servidor iniciar` e
removido por `servidor parar`. Ele e estritamente informativo: o
estado autoritativo continua sendo o que o processo Java reporta
em `/api/info`.
