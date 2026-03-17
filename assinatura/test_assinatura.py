#!/usr/bin/env python3
"""
Testes unitários para o CLI 'assinatura'.

Cobre parsing de argumentos, formatação de saída e invocação simulada
do assinador.jar.
"""

import json
import os
import sys
import unittest
from unittest.mock import patch, MagicMock

# Adicionar diretório do módulo ao path
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from assinatura import (
    criar_parser,
    formatar_resposta,
    invocar_assinador,
    encontrar_java,
)


class TestParserArgumentos(unittest.TestCase):
    """Testes para o parser de argumentos do CLI."""

    def setUp(self):
        self.parser = criar_parser()

    def test_comando_criar_com_parametros_validos(self):
        """Deve parsear corretamente o comando 'criar' com todos os parâmetros."""
        args = self.parser.parse_args([
            "criar", "--documento", "SGVsbG8=", "--certificado", "cert-001"
        ])
        self.assertEqual(args.comando, "criar")
        self.assertEqual(args.documento, "SGVsbG8=")
        self.assertEqual(args.certificado, "cert-001")

    def test_comando_validar_com_parametros_validos(self):
        """Deve parsear corretamente o comando 'validar' com todos os parâmetros."""
        args = self.parser.parse_args([
            "validar", "--documento", "SGVsbG8=", "--assinatura", "dGVzdA=="
        ])
        self.assertEqual(args.comando, "validar")
        self.assertEqual(args.documento, "SGVsbG8=")
        self.assertEqual(args.assinatura, "dGVzdA==")

    def test_criar_sem_documento_gera_erro(self):
        """Deve falhar quando --documento não é fornecido para 'criar'."""
        with self.assertRaises(SystemExit):
            self.parser.parse_args(["criar", "--certificado", "cert-001"])

    def test_criar_sem_certificado_gera_erro(self):
        """Deve falhar quando --certificado não é fornecido para 'criar'."""
        with self.assertRaises(SystemExit):
            self.parser.parse_args(["criar", "--documento", "SGVsbG8="])

    def test_validar_sem_assinatura_gera_erro(self):
        """Deve falhar quando --assinatura não é fornecido para 'validar'."""
        with self.assertRaises(SystemExit):
            self.parser.parse_args(["validar", "--documento", "SGVsbG8="])

    def test_sem_comando_retorna_none(self):
        """Deve retornar comando=None quando nenhum subcomando é fornecido."""
        args = self.parser.parse_args([])
        self.assertIsNone(args.comando)

    def test_parametro_jar_customizado(self):
        """Deve aceitar caminho personalizado para o assinador.jar."""
        args = self.parser.parse_args([
            "--jar", "/caminho/para/assinador.jar",
            "criar", "--documento", "SGVsbG8=", "--certificado", "cert-001"
        ])
        self.assertEqual(args.jar, "/caminho/para/assinador.jar")


class TestFormatarResposta(unittest.TestCase):
    """Testes para a formatação da resposta do Assinador."""

    def test_formatar_resposta_sucesso(self):
        """Deve exibir ícone de sucesso e campos corretamente."""
        resposta = {
            "status": "sucesso",
            "mensagem": "Assinatura criada com sucesso.",
            "algoritmo": "SHA256withRSA",
            "certificado": "cert-001",
            "assinatura": "YXNzaW5hdHVyYQ==",
            "timestamp": "2026-03-17T20:00:00Z"
        }
        saida = formatar_resposta(resposta)
        self.assertIn("[OK]", saida)
        self.assertIn("SUCESSO", saida)
        self.assertIn("Assinatura criada", saida)
        self.assertIn("SHA256withRSA", saida)
        self.assertIn("cert-001", saida)

    def test_formatar_resposta_erro(self):
        """Deve exibir ícone de falha para status de erro."""
        resposta = {
            "status": "falha",
            "mensagem": "Parâmetros inválidos.",
        }
        saida = formatar_resposta(resposta)
        self.assertIn("[ERRO]", saida)
        self.assertIn("FALHA", saida)

    def test_formatar_assinatura_longa_truncada(self):
        """Deve truncar assinaturas muito longas com '...'."""
        resposta = {
            "status": "sucesso",
            "mensagem": "OK",
            "assinatura": "A" * 100,
        }
        saida = formatar_resposta(resposta)
        self.assertIn("...", saida)


class TestInvocarAssinador(unittest.TestCase):
    """Testes para a invocação do assinador.jar."""

    @patch("assinatura.subprocess.run")
    @patch("assinatura.encontrar_java", return_value="java")
    @patch("os.path.isfile", return_value=True)
    def test_invocacao_com_sucesso(self, mock_isfile, mock_java, mock_run):
        """Deve retornar o JSON do assinador quando a execução é bem-sucedida."""
        resposta_esperada = {
            "status": "sucesso",
            "mensagem": "OK",
            "assinatura": "dGVzdA=="
        }
        mock_run.return_value = MagicMock(
            returncode=0,
            stdout=json.dumps(resposta_esperada),
            stderr=""
        )

        resultado = invocar_assinador(
            ["--operacao", "criar", "--documento", "SGVsbG8=", "--certificado", "cert-001"],
            jar_path="/fake/assinador.jar"
        )
        self.assertEqual(resultado["status"], "sucesso")

    @patch("assinatura.subprocess.run")
    @patch("assinatura.encontrar_java", return_value="java")
    @patch("os.path.isfile", return_value=True)
    def test_invocacao_com_erro(self, mock_isfile, mock_java, mock_run):
        """Deve lançar RuntimeError quando o assinador retorna código de erro."""
        mock_run.return_value = MagicMock(
            returncode=1,
            stdout="",
            stderr="ERRO DE PARÂMETRO: Operação inválida"
        )

        with self.assertRaises(RuntimeError) as ctx:
            invocar_assinador(
                ["--operacao", "invalida"],
                jar_path="/fake/assinador.jar"
            )
        self.assertIn("erro", str(ctx.exception).lower())

    def test_jar_nao_encontrado(self):
        """Deve lançar FileNotFoundError quando o jar não existe."""
        with self.assertRaises(FileNotFoundError):
            invocar_assinador(
                ["--operacao", "criar"],
                jar_path="/caminho/inexistente/assinador.jar"
            )


if __name__ == "__main__":
    unittest.main()
