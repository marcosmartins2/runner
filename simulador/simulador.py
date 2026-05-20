#!/usr/bin/env python3
"""
CLI 'simulador' - Gerencia o ciclo de vida do Simulador do HubSaude.

Permite iniciar, parar e consultar o status do simulador.jar sem que o
usuario precise conhecer detalhes do ambiente Java.

Uso:
    python simulador.py iniciar
    python simulador.py status
    python simulador.py parar
"""

import argparse
import json
import os
import socket
import subprocess
import sys
import time
import urllib.error
import urllib.request
from pathlib import Path

__version__ = "0.1.0"

PORTA_SIMULADOR_PADRAO = 8443
TIMEOUT_HTTP_SEGUNDOS = 5
DIRETORIO_HUBSAUDE = Path.home() / ".hubsaude"
ARQUIVO_ESTADO_SIMULADOR = DIRETORIO_HUBSAUDE / "simulador-server.json"
ARQUIVO_RELEASE_LOCAL = DIRETORIO_HUBSAUDE / "release.json"
SIMULADOR_JAR_LOCAL = DIRETORIO_HUBSAUDE / "simulador.jar"


def encontrar_java() -> str:
    """Localiza o executavel java no sistema (JAVA_HOME ou PATH)."""
    java_home = os.environ.get("JAVA_HOME")
    if java_home:
        java_path = os.path.join(java_home, "bin", "java")
        if sys.platform == "win32":
            java_path += ".exe"
        if os.path.isfile(java_path):
            return java_path

    java_cmd = "java.exe" if sys.platform == "win32" else "java"
    try:
        resultado = subprocess.run(
            [java_cmd, "-version"], capture_output=True, text=True, timeout=10
        )
        if resultado.returncode == 0:
            return java_cmd
    except (FileNotFoundError, subprocess.TimeoutExpired):
        pass

    raise FileNotFoundError(
        "Java nao encontrado. Instale o JRE 17+ ou configure JAVA_HOME."
    )


def porta_disponivel(porta: int, host: str = "127.0.0.1") -> bool:
    """Verifica se a porta esta livre para escutar."""
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
        s.settimeout(1.0)
        try:
            s.bind((host, porta))
        except OSError:
            return False
    return True


def _url(porta: int, caminho: str) -> str:
    return f"https://127.0.0.1:{porta}{caminho}"


def _requisicao_json(caminho: str, porta: int, metodo: str = "GET") -> dict:
    """Requisicao HTTP simples (sem verificacao de TLS para localhost)."""
    import ssl

    contexto = ssl.create_default_context()
    contexto.check_hostname = False
    contexto.verify_mode = ssl.CERT_NONE

    request = urllib.request.Request(
        _url(porta, caminho),
        headers={"Accept": "application/json"},
        method=metodo,
    )

    try:
        with urllib.request.urlopen(request, timeout=TIMEOUT_HTTP_SEGUNDOS, context=contexto) as resposta:
            corpo = resposta.read().decode("utf-8")
            try:
                return json.loads(corpo)
            except json.JSONDecodeError:
                return {"status": "sucesso", "mensagem": corpo}
    except urllib.error.HTTPError as e:
        raise RuntimeError(f"Simulador respondeu com erro HTTP {e.code}: {e.reason}")
    except urllib.error.URLError as e:
        raise RuntimeError(f"Nao foi possivel conectar ao Simulador: {e.reason}")
    except TimeoutError as e:
        raise RuntimeError(f"Tempo limite ao consultar o Simulador: {e}")


def simulador_em_execucao(porta: int = PORTA_SIMULADOR_PADRAO) -> bool:
    """Verifica se o Simulador responde em /api/info."""
    try:
        resposta = _requisicao_json("/api/info", porta)
        return resposta.get("status", "").lower() in ("sucesso", "ok", "up", "running")
    except RuntimeError:
        return False


def registrar_estado(pid: int, porta: int, jar_path: str) -> None:
    """Persiste PID/porta/jar em ~/.hubsaude/simulador-server.json."""
    DIRETORIO_HUBSAUDE.mkdir(parents=True, exist_ok=True)
    estado = {
        "pid": pid,
        "porta": porta,
        "jar": str(Path(jar_path).resolve()),
        "iniciadoEm": time.strftime("%Y-%m-%dT%H:%M:%S%z"),
    }
    ARQUIVO_ESTADO_SIMULADOR.write_text(
        json.dumps(estado, indent=2, ensure_ascii=False), encoding="utf-8"
    )


def ler_estado() -> dict:
    """Le o arquivo de estado local, se existir."""
    if not ARQUIVO_ESTADO_SIMULADOR.is_file():
        return {}
    try:
        return json.loads(ARQUIVO_ESTADO_SIMULADOR.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError):
        return {}


def remover_estado() -> None:
    try:
        ARQUIVO_ESTADO_SIMULADOR.unlink()
    except FileNotFoundError:
        pass


def _resolver_jar(jar_path: str = None) -> str:
    """Resolve o caminho do simulador.jar; usa o cache local se nada for passado."""
    if jar_path is None:
        jar_path = str(SIMULADOR_JAR_LOCAL)

    jar_path = str(Path(jar_path).resolve())
    if not os.path.isfile(jar_path):
        raise FileNotFoundError(
            f"simulador.jar nao encontrado em: {jar_path}\n"
            "Forneca o caminho via --jar ou rode 'simulador atualizar' para baixa-lo."
        )
    return jar_path


def iniciar(jar_path: str = None, porta: int = PORTA_SIMULADOR_PADRAO) -> None:
    """Inicia o simulador.jar em background, se nao estiver em execucao."""
    if simulador_em_execucao(porta):
        return

    if not porta_disponivel(porta):
        raise RuntimeError(
            f"A porta {porta} ja esta em uso por outro processo. "
            f"Pare o processo ou escolha outra porta via --porta."
        )

    jar_path = _resolver_jar(jar_path)
    java_cmd = encontrar_java()
    comando = [java_cmd, "-jar", jar_path, "--server.port=" + str(porta)]

    kwargs = {
        "stdin": subprocess.DEVNULL,
        "stdout": subprocess.DEVNULL,
        "stderr": subprocess.DEVNULL,
    }
    if sys.platform == "win32":
        kwargs["creationflags"] = (
            subprocess.CREATE_NEW_PROCESS_GROUP | subprocess.DETACHED_PROCESS
        )
    else:
        kwargs["start_new_session"] = True

    processo = subprocess.Popen(comando, **kwargs)

    prazo = time.time() + 30
    while time.time() < prazo:
        if simulador_em_execucao(porta):
            registrar_estado(processo.pid, porta, jar_path)
            return
        if processo.poll() is not None:
            raise RuntimeError(
                f"Simulador encerrou durante a inicializacao (codigo {processo.returncode})."
            )
        time.sleep(0.5)

    raise RuntimeError(f"Simulador nao respondeu na porta {porta} em 30s.")


def parar(porta: int = PORTA_SIMULADOR_PADRAO) -> dict:
    """Solicita /shutdown ao Simulador."""
    resposta = _requisicao_json("/shutdown", porta, metodo="POST")
    remover_estado()
    return resposta


def status(porta: int = PORTA_SIMULADOR_PADRAO) -> dict:
    """Retorna o status do Simulador via /api/info."""
    if not simulador_em_execucao(porta):
        return {
            "status": "parado",
            "mensagem": f"Simulador nao esta em execucao na porta {porta}.",
        }
    info = _requisicao_json("/api/info", porta)
    return info


def formatar(resposta: dict) -> str:
    """Formata a resposta para exibir no terminal."""
    linhas = ["=" * 60]

    estado = resposta.get("status", "desconhecido")
    icone = "[OK]" if estado.lower() in ("sucesso", "up", "ok", "running") else "[INFO]"
    linhas.append(f"  {icone} Status: {estado.upper()}")
    linhas.append("-" * 60)

    for chave in ("mensagem", "versao", "porta", "pid", "iniciadoEm"):
        if chave in resposta:
            linhas.append(f"  {chave.capitalize():<12} {resposta[chave]}")

    linhas.append("=" * 60)
    return "\n".join(linhas)


def comando_iniciar(args: argparse.Namespace) -> None:
    try:
        iniciar(jar_path=args.jar, porta=args.porta)
        print(f"Simulador em execucao na porta {args.porta}.")
    except (FileNotFoundError, RuntimeError) as e:
        print(f"\n[ERRO]: {e}", file=sys.stderr)
        sys.exit(1)


def comando_parar(args: argparse.Namespace) -> None:
    try:
        if not simulador_em_execucao(args.porta):
            print(f"Simulador nao esta em execucao na porta {args.porta}.")
            return
        resposta = parar(porta=args.porta)
        print(formatar(resposta))
    except (FileNotFoundError, RuntimeError) as e:
        print(f"\n[ERRO]: {e}", file=sys.stderr)
        sys.exit(1)


def comando_status(args: argparse.Namespace) -> None:
    try:
        resposta = status(porta=args.porta)
        print(formatar(resposta))
        estado = ler_estado()
        if estado and estado.get("porta") == args.porta:
            print(f"PID local: {estado.get('pid')} | JAR: {estado.get('jar')}")
    except (FileNotFoundError, RuntimeError) as e:
        print(f"\n[ERRO]: {e}", file=sys.stderr)
        sys.exit(1)


def criar_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        prog="simulador",
        description="Sistema Runner - CLI para gerenciar o Simulador do HubSaude",
        epilog=(
            "Exemplos:\n"
            "  simulador iniciar\n"
            "  simulador status\n"
            "  simulador parar\n"
            "  simulador --porta 9443 iniciar --jar /caminho/simulador.jar\n"
        ),
        formatter_class=argparse.RawDescriptionHelpFormatter,
    )

    parser.add_argument(
        "--version", action="version", version=f"%(prog)s {__version__}"
    )
    parser.add_argument(
        "--porta",
        type=int,
        default=PORTA_SIMULADOR_PADRAO,
        help=f"Porta do Simulador (padrao: {PORTA_SIMULADOR_PADRAO})",
    )

    subparsers = parser.add_subparsers(
        dest="comando",
        title="Comandos disponiveis",
        description="Use 'simulador <comando> --help' para mais detalhes.",
    )

    parser_iniciar = subparsers.add_parser(
        "iniciar", help="Inicia o Simulador do HubSaude"
    )
    parser_iniciar.add_argument(
        "--jar",
        type=str,
        default=None,
        help="Caminho para o simulador.jar (padrao: ~/.hubsaude/simulador.jar)",
    )
    parser_iniciar.set_defaults(func=comando_iniciar)

    parser_parar = subparsers.add_parser(
        "parar", help="Para o Simulador via /shutdown"
    )
    parser_parar.set_defaults(func=comando_parar)

    parser_status = subparsers.add_parser(
        "status", help="Exibe o status atual do Simulador"
    )
    parser_status.set_defaults(func=comando_status)

    return parser


def main():
    parser = criar_parser()
    args = parser.parse_args()

    if not args.comando:
        parser.print_help()
        sys.exit(1)

    args.func(args)


if __name__ == "__main__":
    main()
