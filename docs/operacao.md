# Operacao do Assinador em modo servidor

Este documento descreve como operar o `assinador.jar` em modo
servidor HTTP usando o CLI `assinatura`. Cobre os ciclos de
inicializacao, consulta, encerramento manual e encerramento
programado por inatividade.

## 1. Pre-requisitos

- `assinador-1.0.0.jar` empacotado em `assinador/target/`
  (gerado por `mvn package` no diretorio `assinador/`).
- Java 17+ disponivel via `JAVA_HOME` ou no `PATH`.

## 2. Ciclo de vida pelo CLI

### 2.1. Iniciar o servidor

```bash
python assinatura/assinatura.py servidor iniciar
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
python assinatura/assinatura.py --porta 9090 servidor iniciar
```

### 2.2. Consultar o status

```bash
python assinatura/assinatura.py servidor status
```

A saida exibe o resultado de `/api/info` e, quando ha registro
local, mostra tambem PID e porta gravados em
`~/.hubsaude/assinador-server.json`.

### 2.3. Parar o servidor

```bash
python assinatura/assinatura.py servidor parar
```

O CLI envia `POST /shutdown` e, em seguida, remove o arquivo de
estado local.

## 3. Encerramento programado por inatividade

A partir desta iteracao, o servidor pode encerrar automaticamente
apos um intervalo sem requisicoes. Use `--parar-apos-minutos` ao
inicia-lo:

```bash
python assinatura/assinatura.py servidor iniciar --parar-apos-minutos 30
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
python assinatura/assinatura.py criar --documento SGVsbG8= --certificado cert-001
```

Se o servidor nao estiver em execucao, ele e iniciado sob demanda
antes da primeira chamada e mantido vivo para chamadas
subsequentes. Para forcar invocacao direta (cold start), use
`--modo local`:

```bash
python assinatura/assinatura.py --modo local criar --documento SGVsbG8= --certificado cert-001
```

## 5. Diagnostico

| Sintoma | Verificacao | Acao |
|---|---|---|
| `Assinador HTTP nao respondeu na porta N` | porta ocupada por outro processo, ou JVM lenta para subir | escolher outra porta via `--porta` ou aguardar e tentar novamente |
| `Java nao encontrado no sistema` | `JAVA_HOME` nao definido e `java` ausente do `PATH` | instalar JDK 17+ ou exportar `JAVA_HOME` |
| `Assinador nao encontrado em ...` | `assinador.jar` nao foi empacotado | rodar `mvn package` em `assinador/` |
| Status mostra dados antigos apos kill manual | arquivo `~/.hubsaude/assinador-server.json` ficou orfao | apagar o arquivo ou rodar `servidor iniciar` novamente para sobrescrever |

## 6. Arquivos de estado

O CLI grava em `~/.hubsaude/assinador-server.json`:

```json
{
  "pid": 12345,
  "porta": 8080,
  "jar": "/caminho/absoluto/para/assinador-1.0.0.jar",
  "iniciadoEm": "2026-05-12T22:00:00-0300"
}
```

Esse arquivo e gravado a cada execucao de `servidor iniciar` e
removido por `servidor parar`. Ele e estritamente informativo: o
estado autoritativo continua sendo o que o processo Java reporta
em `/api/info`.
