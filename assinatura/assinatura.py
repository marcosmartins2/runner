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
from pathlib import Path

# Versão do CLI
__version__ = "1.0.0"

# Caminho padrão para o assinador.jar (relativo ao diretório do CLI)
ASSINADOR_JAR_PADRAO = os.path.join(
    os.path.dirname(os.path.abspath(__file__)),
    "..", "assinador", "target", "assinador-1.0.0.jar"
)


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
        resposta = invocar_assinador(args_java, jar_path=jar)
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
        resposta = invocar_assinador(args_java, jar_path=jar)
        print(formatar_resposta(resposta))
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
               "  assinatura validar --documento SGVsbG8= --assinatura dGVzdA==\n",
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
