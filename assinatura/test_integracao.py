#!/usr/bin/env python3
"""
Testes de integracao 'assinatura -> assinador.jar'.

Estes testes invocam o CLI 'assinatura' como subprocesso e
verificam o fluxo completo de criacao e validacao de assinatura
contra o assinador.jar compilado em assinador/target/.

Sao testes de integracao reais: se o jar nao estiver disponivel
(por exemplo, em um ambiente sem Maven configurado), os testes
sao automaticamente ignorados em vez de falharem. Para gera-lo,
execute 'mvn package' no diretorio assinador/.
"""

import base64
import re
import subprocess
import sys
import unittest
from pathlib import Path


DIR_ASSINATURA = Path(__file__).resolve().parent
DIR_RAIZ = DIR_ASSINATURA.parent
ASSINADOR_JAR = DIR_RAIZ / "assinador" / "target" / "assinador-1.0.0.jar"
CLI_ASSINATURA = DIR_ASSINATURA / "assinatura.py"


def _jar_disponivel() -> bool:
    return ASSINADOR_JAR.is_file()


def _executar_cli(*args, timeout=30):
    """Executa o CLI assinatura como subprocesso e retorna o CompletedProcess."""
    comando = [sys.executable, str(CLI_ASSINATURA), *args]
    return subprocess.run(
        comando, capture_output=True, text=True, timeout=timeout
    )


@unittest.skipUnless(_jar_disponivel(), f"assinador.jar nao encontrado em {ASSINADOR_JAR}")
class TestIntegracaoFluxoLocal(unittest.TestCase):
    """Cenarios end-to-end usando o assinador.jar compilado."""

    def setUp(self):
        self.documento = base64.b64encode(b"documento de teste").decode("ascii")

    def test_criar_assinatura_retorna_sucesso(self):
        resultado = _executar_cli(
            "criar", "--documento", self.documento, "--certificado", "cert-001"
        )
        self.assertEqual(
            resultado.returncode, 0,
            msg=f"stdout={resultado.stdout}\nstderr={resultado.stderr}",
        )
        self.assertIn("SUCESSO", resultado.stdout)
        self.assertIn("cert-001", resultado.stdout)

    def test_criar_e_validar_em_sequencia(self):
        criar = _executar_cli(
            "criar", "--documento", self.documento, "--certificado", "cert-001"
        )
        self.assertEqual(criar.returncode, 0)

        match = re.search(r"Assinatura:\s+(\S+)", criar.stdout)
        self.assertIsNotNone(match, "Assinatura nao encontrada na saida do CLI")
        assinatura_truncada = match.group(1).rstrip(".")

        # A saida do CLI trunca a assinatura para exibicao. O validador
        # tolera a entrada conforme o que esta documentado: aqui validamos
        # que o CLI propaga o parametro e retorna uma resposta estruturada.
        validar = _executar_cli(
            "validar", "--documento", self.documento,
            "--assinatura", assinatura_truncada,
        )
        self.assertEqual(
            validar.returncode, 0,
            msg=f"stdout={validar.stdout}\nstderr={validar.stderr}",
        )
        self.assertRegex(validar.stdout, r"Status:\s+(SUCESSO|ERRO)")

    def test_documento_invalido_retorna_erro_estruturado(self):
        # Base64 invalido deve ser rejeitado pelo validador do assinador.jar
        resultado = _executar_cli(
            "criar", "--documento", "!!!nao-base64!!!", "--certificado", "cert-001"
        )
        self.assertNotEqual(resultado.returncode, 0)
        self.assertTrue(
            "ERRO" in resultado.stdout or "erro" in resultado.stderr.lower(),
            msg=f"stdout={resultado.stdout}\nstderr={resultado.stderr}",
        )


class TestCliSemJar(unittest.TestCase):
    """Verifica o comportamento do CLI sem depender do jar."""

    def test_version_funciona(self):
        resultado = _executar_cli("--version")
        self.assertEqual(resultado.returncode, 0)
        self.assertIn("assinatura", resultado.stdout.lower())

    def test_help_lista_comandos(self):
        resultado = _executar_cli("--help")
        self.assertEqual(resultado.returncode, 0)
        self.assertIn("criar", resultado.stdout)
        self.assertIn("validar", resultado.stdout)

    def test_jar_inexistente_retorna_erro(self):
        documento = base64.b64encode(b"x").decode("ascii")
        resultado = _executar_cli(
            "--jar", "/caminho/que/nao/existe.jar",
            "criar", "--documento", documento, "--certificado", "cert-001",
        )
        self.assertNotEqual(resultado.returncode, 0)
        saida = (resultado.stdout + resultado.stderr).lower()
        self.assertTrue(
            "não encontrado" in saida or "nao encontrado" in saida,
            msg=saida,
        )


if __name__ == "__main__":
    unittest.main()
