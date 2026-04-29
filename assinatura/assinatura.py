#!/usr/bin/env python3
"""
CLI 'assinatura' - Interface de linha de comandos do Sistema Runner.

Permite ao usuário invocar o Assinador (assinador.jar) para criar
e validar assinaturas digitais sem conhecer os detalhes técnicos
de configuração Java.

Uso:
    python assinatura.py criar --documento <base64> --certificado <id>
    python assinatura.py validar --documento <base64> --assinatura <base64>
"""

import argparse
import json
import os
import subprocess
import sys
import time
import urllib.error
import urllib.request
from pathlib import Path

# Versão do CLI
__version__ = "1.0.0"

# Caminho padrão para o assinador.jar (relativo ao diretório do CLI)
ASSINADOR_JAR_PADRAO = os.path.join(
    os.path.dirname(os.path.abspath(__file__)),
    "..", "assinador", "target", "assinador-1.0.0.jar"
)
PORTA_ASSINADOR_PADRAO = 8080
TIMEOUT_HTTP_SEGUNDOS = 3


def encontrar_java() -> str:
    """
    Localiza o executável Java no sistema.

    Procura em ordem:
    1. Variável de ambiente JAVA_HOME
    2. Comando 'java' no PATH do sistema

    Returns:
        Caminho para o executável Java.

    Raises:
        FileNotFoundError: Se o Java não for encontrado.
    """
    # Tentar JAVA_HOME primeiro
    java_home = os.environ.get("JAVA_HOME")
    if java_home:
        java_path = os.path.join(java_home, "bin", "java")
        if sys.platform == "win32":
            java_path += ".exe"
        if os.path.isfile(java_path):
            return java_path

    # Tentar java no PATH
    java_cmd = "java.exe" if sys.platform == "win32" else "java"
    try:
        resultado = subprocess.run(
            [java_cmd, "-version"],
            capture_output=True, text=True, timeout=10
        )
        if resultado.returncode == 0:
            return java_cmd
    except (FileNotFoundError, subprocess.TimeoutExpired):
        pass

    raise FileNotFoundError(
        "Java não encontrado no sistema. "
        "Instale o JDK 17+ ou configure a variável JAVA_HOME."
    )


def invocar_assinador(args_java: list, jar_path: str = None) -> dict:
    """
    Invoca o assinador.jar via linha de comandos (modo local/CLI).

    Args:
        args_java: Lista de argumentos para o assinador.jar.
        jar_path: Caminho para o arquivo assinador.jar.

    Returns:
        Dicionário com a resposta JSON do assinador.

    Raises:
        RuntimeError: Se a invocação falhar.
    """
    if jar_path is None:
        jar_path = ASSINADOR_JAR_PADRAO

    # Resolver caminho absoluto
    jar_path = str(Path(jar_path).resolve())

    if not os.path.isfile(jar_path):
        raise FileNotFoundError(
            f"Assinador não encontrado em: {jar_path}\n"
            "Execute 'mvn package' no diretório do assinador primeiro."
        )

    java_cmd = encontrar_java()
    comando = [java_cmd, "-jar", jar_path] + args_java

    try:
        resultado = subprocess.run(
            comando,
            capture_output=True,
            text=True,
            timeout=30
        )

        if resultado.returncode != 0:
            erro = resultado.stderr.strip()
            raise RuntimeError(
                f"O Assinador retornou erro (código {resultado.returncode}):\n{erro}"
            )

        # Parsear resposta JSON
        saida = resultado.stdout.strip()
        if not saida:
            raise RuntimeError("O Assinador não retornou nenhuma resposta.")

        return json.loads(saida)

    except subprocess.TimeoutExpired:
        raise RuntimeError(
            "Tempo limite excedido ao invocar o Assinador. "
            "Verifique se o assinador.jar está funcionando corretamente."
        )
    except json.JSONDecodeError as e:
        raise RuntimeError(
            f"Resposta do Assinador não é JSON válido: {e}"
        )


def _resolver_jar(jar_path: str = None) -> str:
    """Resolve e valida o caminho do assinador.jar."""
    if jar_path is None:
        jar_path = ASSINADOR_JAR_PADRAO

    jar_path = str(Path(jar_path).resolve())
    if not os.path.isfile(jar_path):
        raise FileNotFoundError(
            f"Assinador não encontrado em: {jar_path}\n"
            "Execute 'mvn package' no diretório do assinador primeiro."
        )
    return jar_path


def _url(porta: int, caminho: str) -> str:
    return f"http://127.0.0.1:{porta}{caminho}"


def _requisicao_json(caminho: str, porta: int, metodo: str = "GET", dados: dict = None) -> dict:
    """Executa uma requisição HTTP JSON contra o assinador em modo servidor."""
    corpo = None
    headers = {"Accept": "application/json"}

    if dados is not None:
        corpo = json.dumps(dados).encode("utf-8")
        headers["Content-Type"] = "application/json"

    request = urllib.request.Request(
        _url(porta, caminho),
        data=corpo,
        headers=headers,
        method=metodo,
    )

    try:
        with urllib.request.urlopen(request, timeout=TIMEOUT_HTTP_SEGUNDOS) as resposta:
            return json.loads(resposta.read().decode("utf-8"))
    except urllib.error.HTTPError as e:
        mensagem = e.reason
        try:
            erro = json.loads(e.read().decode("utf-8"))
            mensagem = erro.get("mensagem", mensagem)
        except (json.JSONDecodeError, UnicodeDecodeError):
            pass
        raise RuntimeError(f"Assinador HTTP retornou erro {e.code}: {mensagem}")
    except urllib.error.URLError as e:
        raise RuntimeError(f"Não foi possível conectar ao Assinador HTTP: {e.reason}")
    except TimeoutError as e:
        raise RuntimeError(f"Tempo limite ao conectar ao Assinador HTTP: {e}")
    except json.JSONDecodeError as e:
        raise RuntimeError(f"Resposta HTTP do Assinador não é JSON válido: {e}")


def servidor_em_execucao(porta: int = PORTA_ASSINADOR_PADRAO) -> bool:
    """Verifica se o assinador HTTP responde na porta informada."""
    try:
        resposta = _requisicao_json("/api/info", porta)
        return resposta.get("status") == "sucesso"
    except RuntimeError:
        return False


def iniciar_servidor(jar_path: str = None, porta: int = PORTA_ASSINADOR_PADRAO) -> None:
    """Inicia o assinador.jar em modo servidor, se ele ainda não estiver ativo."""
    if servidor_em_execucao(porta):
        return

    jar_path = _resolver_jar(jar_path)
    java_cmd = encontrar_java()
    comando = [java_cmd, "-jar", jar_path, "--server", "--port", str(porta)]

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

    prazo = time.time() + 10
    while time.time() < prazo:
        if servidor_em_execucao(porta):
            return
        if processo.poll() is not None:
            raise RuntimeError(
                f"Assinador HTTP encerrou durante a inicialização (código {processo.returncode})."
            )
        time.sleep(0.2)

    raise RuntimeError(f"Assinador HTTP não respondeu na porta {porta}.")


def parar_servidor(porta: int = PORTA_ASSINADOR_PADRAO) -> dict:
    """Solicita a interrupção do assinador em modo servidor."""
    return _requisicao_json("/shutdown", porta, metodo="POST")


def invocar_assinador_http(operacao: str, dados: dict, porta: int = PORTA_ASSINADOR_PADRAO) -> dict:
    """Invoca o assinador.jar já iniciado em modo servidor HTTP."""
    caminho = "/api/sign" if operacao == "criar" else "/api/validate"
    return _requisicao_json(caminho, porta, metodo="POST", dados=dados)


def formatar_resposta(resposta: dict) -> str:
    """
    Formata a resposta do Assinador de forma legível para o usuário.

    Args:
        resposta: Dicionário com a resposta do assinador.

    Returns:
        String formatada para exibição no terminal.
    """
    linhas = []
    linhas.append("=" * 60)

    status = resposta.get("status", "desconhecido")
    icone = "[OK]" if status == "sucesso" else "[ERRO]"
    linhas.append(f"  {icone} Status: {status.upper()}")
    linhas.append("-" * 60)

    if "mensagem" in resposta:
        linhas.append(f"  Mensagem:    {resposta['mensagem']}")

    if "algoritmo" in resposta:
        linhas.append(f"  Algoritmo:   {resposta['algoritmo']}")

    if "certificado" in resposta:
        linhas.append(f"  Certificado: {resposta['certificado']}")

    if "assinatura" in resposta:
        assinatura = resposta["assinatura"]
        if len(assinatura) > 50:
            assinatura = assinatura[:50] + "..."
        linhas.append(f"  Assinatura:  {assinatura}")

    if "timestamp" in resposta:
        linhas.append(f"  Timestamp:   {resposta['timestamp']}")

    linhas.append("=" * 60)
    return "\n".join(linhas)


def comando_criar(args: argparse.Namespace) -> None:
    """Executa o comando de criação de assinatura."""
    args_java = [
        "--operacao", "criar",
        "--documento", args.documento,
        "--certificado", args.certificado,
    ]

    jar = getattr(args, "jar", None)

    try:
        if args.modo == "local":
            resposta = invocar_assinador(args_java, jar_path=jar)
        else:
            iniciar_servidor(jar_path=jar, porta=args.porta)
            resposta = invocar_assinador_http(
                "criar",
                {"documento": args.documento, "certificado": args.certificado},
                porta=args.porta,
            )
        print(formatar_resposta(resposta))
    except (FileNotFoundError, RuntimeError) as e:
        print(f"\n[ERRO]: {e}", file=sys.stderr)
        sys.exit(1)


def comando_validar(args: argparse.Namespace) -> None:
    """Executa o comando de validação de assinatura."""
    args_java = [
        "--operacao", "validar",
        "--documento", args.documento,
        "--assinatura", args.assinatura,
    ]

    jar = getattr(args, "jar", None)

    try:
        if args.modo == "local":
            resposta = invocar_assinador(args_java, jar_path=jar)
        else:
            iniciar_servidor(jar_path=jar, porta=args.porta)
            resposta = invocar_assinador_http(
                "validar",
                {"documento": args.documento, "assinatura": args.assinatura},
                porta=args.porta,
            )
        print(formatar_resposta(resposta))
    except (FileNotFoundError, RuntimeError) as e:
        print(f"\n[ERRO]: {e}", file=sys.stderr)
        sys.exit(1)


def comando_servidor(args: argparse.Namespace) -> None:
    """Gerencia o ciclo de vida do assinador em modo servidor."""
    try:
        if args.acao == "iniciar":
            iniciar_servidor(jar_path=args.jar, porta=args.porta)
            print(f"Assinador em execução na porta {args.porta}.")
        elif args.acao == "parar":
            if not servidor_em_execucao(args.porta):
                print(f"Assinador não está em execução na porta {args.porta}.")
                return
            resposta = parar_servidor(porta=args.porta)
            print(formatar_resposta(resposta))
        elif args.acao == "status":
            if servidor_em_execucao(args.porta):
                resposta = _requisicao_json("/api/info", args.porta)
                print(formatar_resposta(resposta))
            else:
                print(f"Assinador não está em execução na porta {args.porta}.")
    except (FileNotFoundError, RuntimeError) as e:
        print(f"\n[ERRO]: {e}", file=sys.stderr)
        sys.exit(1)


def criar_parser() -> argparse.ArgumentParser:
    """
    Cria e configura o parser de argumentos do CLI.

    Returns:
        ArgumentParser configurado com subcomandos 'criar' e 'validar'.
    """
    parser = argparse.ArgumentParser(
        prog="assinatura",
        description="Sistema Runner - CLI para assinatura digital",
        epilog="Exemplos:\n"
               "  assinatura criar --documento SGVsbG8= --certificado cert-001\n"
               "  assinatura validar --documento SGVsbG8= --assinatura dGVzdA==\n"
               "  assinatura --modo local criar --documento SGVsbG8= --certificado cert-001\n"
               "  assinatura servidor status\n",
        formatter_class=argparse.RawDescriptionHelpFormatter,
    )

    parser.add_argument(
        "--version", action="version",
        version=f"%(prog)s {__version__}"
    )

    parser.add_argument(
        "--jar", type=str, default=None,
        help="Caminho para o assinador.jar (padrão: auto-detectar)"
    )

    parser.add_argument(
        "--modo",
        choices=["servidor", "local"],
        default="servidor",
        help="Modo de invocação do assinador.jar (padrão: servidor)"
    )
    parser.add_argument(
        "--porta",
        type=int,
        default=PORTA_ASSINADOR_PADRAO,
        help=f"Porta do assinador em modo servidor (padrão: {PORTA_ASSINADOR_PADRAO})"
    )

    subparsers = parser.add_subparsers(
        dest="comando",
        title="Comandos disponíveis",
        description="Use 'assinatura <comando> --help' para mais detalhes.",
    )

    # Subcomando: criar
    parser_criar = subparsers.add_parser(
        "criar",
        help="Criar uma assinatura digital (simulação)",
        description="Invoca o Assinador para criar uma assinatura digital simulada.",
    )
    parser_criar.add_argument(
        "--documento", required=True,
        help="Documento codificado em Base64 para assinar"
    )
    parser_criar.add_argument(
        "--certificado", required=True,
        help="Identificador do certificado digital"
    )
    parser_criar.set_defaults(func=comando_criar)

    # Subcomando: validar
    parser_validar = subparsers.add_parser(
        "validar",
        help="Validar uma assinatura digital (simulação)",
        description="Invoca o Assinador para validar uma assinatura digital simulada.",
    )
    parser_validar.add_argument(
        "--documento", required=True,
        help="Documento codificado em Base64 que foi assinado"
    )
    parser_validar.add_argument(
        "--assinatura", required=True,
        help="Assinatura codificada em Base64 para validar"
    )
    parser_validar.set_defaults(func=comando_validar)

    # Subcomando: servidor
    parser_servidor = subparsers.add_parser(
        "servidor",
        help="Gerenciar o assinador.jar em modo servidor",
        description="Inicia, consulta ou interrompe o Assinador HTTP.",
    )
    parser_servidor.add_argument(
        "acao",
        choices=["iniciar", "parar", "status"],
        help="Ação de gerenciamento do servidor"
    )
    parser_servidor.set_defaults(func=comando_servidor)

    return parser


def main():
    """Ponto de entrada principal do CLI."""
    parser = criar_parser()
    args = parser.parse_args()

    if not args.comando:
        parser.print_help()
        sys.exit(1)

    args.func(args)


if __name__ == "__main__":
    main()
