#!/usr/bin/env python3
"""
Provisionador automatico de JDK/JRE para o Sistema Runner.

Modulo responsavel por detectar a presenca de um JDK compativel
no ambiente do usuario e, quando ausente, baixar e instalar um
JRE compativel obtido do Eclipse Temurin (Adoptium).

Cobre a historia US-04 do plano oficial da disciplina:
    "Provisionar JDK Automaticamente".
"""

import os
import platform
import re
import subprocess
import sys
import urllib.request
from pathlib import Path

VERSAO_JDK_MINIMA = 17

DIRETORIO_HUBSAUDE = Path.home() / ".hubsaude"
DIRETORIO_JRE = DIRETORIO_HUBSAUDE / "jre"

URLS_ADOPTIUM = {
    ("windows", "x64"): (
        "https://api.adoptium.net/v3/binary/latest/21/ga/"
        "windows/x64/jre/hotspot/normal/eclipse"
    ),
    ("linux", "x64"): (
        "https://api.adoptium.net/v3/binary/latest/21/ga/"
        "linux/x64/jre/hotspot/normal/eclipse"
    ),
    ("mac", "x64"): (
        "https://api.adoptium.net/v3/binary/latest/21/ga/"
        "mac/x64/jre/hotspot/normal/eclipse"
    ),
    ("mac", "aarch64"): (
        "https://api.adoptium.net/v3/binary/latest/21/ga/"
        "mac/aarch64/jre/hotspot/normal/eclipse"
    ),
}


def detectar_plataforma() -> tuple:
    """Retorna a tupla (sistema, arquitetura) usada nas URLs do Adoptium."""
    sistema = sys.platform
    if sistema.startswith("win"):
        sistema_norm = "windows"
    elif sistema.startswith("darwin"):
        sistema_norm = "mac"
    elif sistema.startswith("linux"):
        sistema_norm = "linux"
    else:
        sistema_norm = sistema

    maquina = platform.machine().lower()
    if maquina in ("amd64", "x86_64"):
        arquitetura = "x64"
    elif maquina in ("arm64", "aarch64"):
        arquitetura = "aarch64"
    else:
        arquitetura = maquina

    return sistema_norm, arquitetura


def extrair_versao(saida_java_version: str) -> int:
    """Extrai o numero principal da versao a partir da saida do `java -version`."""
    correspondencia = re.search(r'version "(\d+)', saida_java_version)
    if not correspondencia:
        return 0
    return int(correspondencia.group(1))


def detectar_jdk_local(versao_minima: int = VERSAO_JDK_MINIMA) -> str:
    """
    Retorna o caminho do executavel `java` se a versao instalada atender
    o minimo exigido. Retorna string vazia caso contrario.

    A ordem de busca segue a mesma adotada por `encontrar_java`:
    1. Variavel JAVA_HOME.
    2. Comando `java` no PATH.
    3. JRE provisionado em `~/.hubsaude/jre`.
    """
    candidatos = []

    java_home = os.environ.get("JAVA_HOME")
    if java_home:
        candidato = Path(java_home) / "bin" / ("java.exe" if sys.platform == "win32" else "java")
        candidatos.append(candidato)

    candidato_path = "java.exe" if sys.platform == "win32" else "java"
    candidatos.append(Path(candidato_path))

    jre_local = DIRETORIO_JRE / "bin" / ("java.exe" if sys.platform == "win32" else "java")
    candidatos.append(jre_local)

    for candidato in candidatos:
        try:
            resultado = subprocess.run(
                [str(candidato), "-version"],
                capture_output=True, text=True, timeout=10,
            )
        except (FileNotFoundError, subprocess.TimeoutExpired):
            continue

        # `java -version` escreve no stderr
        saida = (resultado.stderr or "") + (resultado.stdout or "")
        if resultado.returncode == 0 and extrair_versao(saida) >= versao_minima:
            return str(candidato)

    return ""


def url_download_jre() -> str:
    """Retorna a URL do JRE compativel para a plataforma atual."""
    chave = detectar_plataforma()
    url = URLS_ADOPTIUM.get(chave)
    if not url:
        raise RuntimeError(
            f"Nao ha JRE pre-configurado para a plataforma {chave}. "
            "Defina JAVA_HOME manualmente."
        )
    return url


def baixar_jre(destino: Path = None) -> Path:
    """
    Baixa o arquivo do JRE compativel para o diretorio destino.

    Retorna o caminho do arquivo baixado (geralmente `.tar.gz` ou `.zip`).
    O destino padrao e `~/.hubsaude/jre/download/`.
    """
    if destino is None:
        destino = DIRETORIO_JRE / "download"
    destino.mkdir(parents=True, exist_ok=True)

    url = url_download_jre()
    nome_arquivo = "jre-windows.zip" if sys.platform == "win32" else "jre.tar.gz"
    caminho_destino = destino / nome_arquivo

    with urllib.request.urlopen(url, timeout=60) as resposta:
        caminho_destino.write_bytes(resposta.read())
    return caminho_destino


def provisionar_jdk() -> str:
    """
    Garante a existencia de um JDK/JRE compativel.

    Retorna o caminho absoluto do executavel `java`. Se nenhum JDK
    valido for encontrado, faz o download de um JRE 21 do Adoptium.
    """
    java_local = detectar_jdk_local()
    if java_local:
        return java_local

    baixar_jre()
    java_provisionado = DIRETORIO_JRE / "bin" / ("java.exe" if sys.platform == "win32" else "java")
    return str(java_provisionado)


if __name__ == "__main__":
    print(provisionar_jdk())
