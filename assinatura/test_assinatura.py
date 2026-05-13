#!/usr/bin/env python3
"""
Testes unitários para o CLI 'assinatura'.

Cobre parsing de argumentos, formatação de saída e invocação simulada
do assinador.jar.
"""

import json
import os
import sys
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch, MagicMock

# Adicionar diretório do módulo ao path
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from assinatura import (
    criar_parser,
    formatar_resposta,
    encontrar_java,
    invocar_assinador,
    invocar_assinador_http,
    iniciar_servidor,
    ler_estado_servidor,
    parar_servidor,
    registrar_estado_servidor,
    servidor_em_execucao,
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

    def test_modo_servidor_e_padrao(self):
        """Deve usar o modo servidor quando o usuario nao escolher modo local."""
        args = self.parser.parse_args([
            "criar", "--documento", "SGVsbG8=", "--certificado", "cert-001"
        ])
        self.assertEqual(args.modo, "servidor")

    def test_modo_local_explicito(self):
        """Deve aceitar invocacao local explicita do assinador.jar."""
        args = self.parser.parse_args([
            "--modo", "local",
            "validar", "--documento", "SGVsbG8=", "--assinatura", "dGVzdA=="
        ])
        self.assertEqual(args.modo, "local")

    def test_comando_servidor_status(self):
        """Deve parsear o comando de status do servidor."""
        args = self.parser.parse_args(["servidor", "status"])
        self.assertEqual(args.comando, "servidor")
        self.assertEqual(args.acao, "status")
        self.assertEqual(args.porta, 8080)

    def test_comando_servidor_aceita_parada_programada(self):
        """Deve aceitar timeout de inatividade no comando iniciar."""
        args = self.parser.parse_args(["servidor", "iniciar", "--parar-apos-minutos", "30"])
        self.assertEqual(args.acao, "iniciar")
        self.assertEqual(args.parar_apos_minutos, 30)


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

        args_fornecidos = ["--operacao", "criar", "--documento", "SGVsbG8=", "--certificado", "cert-001"]
        resultado = invocar_assinador(
            args_fornecidos,
            jar_path="/fake/assinador.jar"
        )
        self.assertEqual(resultado["status"], "sucesso")

        caminho_esperado = str(Path("/fake/assinador.jar").resolve())
        comando_esperado = ["java", "-jar", caminho_esperado] + args_fornecidos
        mock_run.assert_called_once_with(
            comando_esperado,
            capture_output=True,
            text=True,
            timeout=30
        )

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


class TestModoServidor(unittest.TestCase):
    """Testes para a integraÃ§Ã£o HTTP do CLI."""

    @patch("assinatura._requisicao_json", return_value={"status": "sucesso"})
    def test_servidor_em_execucao_quando_info_responde_sucesso(self, mock_request):
        """Deve considerar o servidor ativo quando /api/info retorna sucesso."""
        self.assertTrue(servidor_em_execucao(9090))
        mock_request.assert_called_once_with("/api/info", 9090)

    @patch("assinatura._requisicao_json", side_effect=RuntimeError("offline"))
    def test_servidor_em_execucao_retorna_falso_quando_offline(self, mock_request):
        """Deve retornar falso quando nÃ£o conseguir conectar ao servidor."""
        self.assertFalse(servidor_em_execucao(9090))
        mock_request.assert_called_once_with("/api/info", 9090)

    @patch("assinatura._requisicao_json", return_value={"status": "sucesso"})
    def test_invocar_assinador_http_criar_usa_endpoint_sign(self, mock_request):
        """Deve enviar criaÃ§Ã£o para o endpoint HTTP de assinatura."""
        dados = {"documento": "SGVsbG8=", "certificado": "cert-001"}
        invocar_assinador_http("criar", dados, porta=9090)
        mock_request.assert_called_once_with(
            "/api/sign", 9090, metodo="POST", dados=dados
        )

    @patch("assinatura._requisicao_json", return_value={"status": "sucesso"})
    def test_invocar_assinador_http_validar_usa_endpoint_validate(self, mock_request):
        """Deve enviar validaÃ§Ã£o para o endpoint HTTP de validaÃ§Ã£o."""
        dados = {"documento": "SGVsbG8=", "assinatura": "dGVzdA=="}
        invocar_assinador_http("validar", dados, porta=9090)
        mock_request.assert_called_once_with(
            "/api/validate", 9090, metodo="POST", dados=dados
        )


    @patch("assinatura.registrar_estado_servidor")
    @patch("assinatura.subprocess.Popen")
    @patch("assinatura.encontrar_java", return_value="java")
    @patch("assinatura._resolver_jar", return_value="/fake/assinador.jar")
    @patch("assinatura.servidor_em_execucao", side_effect=[False, True])
    def test_iniciar_servidor_envia_timeout_e_registra_estado(
            self, mock_online, mock_resolver, mock_java, mock_popen, mock_registrar):
        """Deve iniciar o servidor com parada programada e registrar PID/porta."""
        processo = MagicMock()
        processo.pid = 1234
        processo.poll.return_value = None
        mock_popen.return_value = processo

        iniciar_servidor(
            jar_path="/fake/assinador.jar",
            porta=9090,
            parar_apos_minutos=15,
        )

        comando = mock_popen.call_args.args[0]
        self.assertEqual(
            comando,
            [
                "java", "-jar", "/fake/assinador.jar",
                "--server", "--port", "9090",
                "--parar-apos-minutos", "15",
            ],
        )
        mock_registrar.assert_called_once_with(1234, 9090, "/fake/assinador.jar")

    def test_registrar_e_ler_estado_servidor(self):
        """Deve persistir PID, porta e caminho do JAR em arquivo local."""
        with tempfile.TemporaryDirectory() as tmp:
            arquivo_estado = Path(tmp) / "assinador-server.json"
            with patch("assinatura.DIRETORIO_ESTADO", Path(tmp)), \
                    patch("assinatura.ARQUIVO_ESTADO_ASSINADOR", arquivo_estado):
                registrar_estado_servidor(1234, 9090, "/fake/assinador.jar")
                estado = ler_estado_servidor()

        self.assertEqual(estado["pid"], 1234)
        self.assertEqual(estado["porta"], 9090)
        self.assertIn("assinador.jar", estado["jar"])

    @patch("assinatura._requisicao_json", return_value={"status": "sucesso"})
    def test_parar_servidor_remove_estado_local(self, mock_request):
        """Deve remover arquivo de estado quando o servidor for interrompido."""
        with tempfile.TemporaryDirectory() as tmp:
            arquivo_estado = Path(tmp) / "assinador-server.json"
            arquivo_estado.write_text("{}", encoding="utf-8")
            with patch("assinatura.ARQUIVO_ESTADO_ASSINADOR", arquivo_estado):
                resposta = parar_servidor(9090)
                existe = arquivo_estado.exists()

        self.assertEqual(resposta["status"], "sucesso")
        self.assertFalse(existe)


if __name__ == "__main__":
    unittest.main()
