#!/usr/bin/env python3
"""Testes unitarios para o CLI 'simulador'."""

import json
import os
import socket
import sys
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch, MagicMock

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from simulador import (
    criar_parser,
    formatar,
    porta_disponivel,
    registrar_estado,
    ler_estado,
    simulador_em_execucao,
    parar,
    iniciar,
    PORTA_SIMULADOR_PADRAO,
)


class TestParser(unittest.TestCase):
    def setUp(self):
        self.parser = criar_parser()

    def test_comando_iniciar_default(self):
        args = self.parser.parse_args(["iniciar"])
        self.assertEqual(args.comando, "iniciar")
        self.assertEqual(args.porta, PORTA_SIMULADOR_PADRAO)
        self.assertIsNone(args.jar)

    def test_comando_status(self):
        args = self.parser.parse_args(["status"])
        self.assertEqual(args.comando, "status")

    def test_comando_parar(self):
        args = self.parser.parse_args(["parar"])
        self.assertEqual(args.comando, "parar")

    def test_porta_customizada(self):
        args = self.parser.parse_args(["--porta", "9443", "iniciar"])
        self.assertEqual(args.porta, 9443)

    def test_jar_customizado(self):
        args = self.parser.parse_args(["iniciar", "--jar", "/tmp/simulador.jar"])
        self.assertEqual(args.jar, "/tmp/simulador.jar")

    def test_sem_comando_retorna_none(self):
        args = self.parser.parse_args([])
        self.assertIsNone(args.comando)


class TestPortaDisponivel(unittest.TestCase):
    def test_porta_livre_retorna_true(self):
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
            s.bind(("127.0.0.1", 0))
            porta_alvo = s.getsockname()[1]
        # Apos fechar o socket, a porta deve voltar a estar livre
        self.assertTrue(porta_disponivel(porta_alvo))

    def test_porta_ocupada_retorna_false(self):
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
            s.bind(("127.0.0.1", 0))
            s.listen(1)
            porta_em_uso = s.getsockname()[1]
            self.assertFalse(porta_disponivel(porta_em_uso))


class TestEstadoServidor(unittest.TestCase):
    def test_registrar_e_ler_estado(self):
        with tempfile.TemporaryDirectory() as tmp:
            arquivo = Path(tmp) / "simulador-server.json"
            with patch("simulador.DIRETORIO_HUBSAUDE", Path(tmp)), \
                    patch("simulador.ARQUIVO_ESTADO_SIMULADOR", arquivo):
                registrar_estado(4321, 8443, "/fake/simulador.jar")
                estado = ler_estado()
        self.assertEqual(estado["pid"], 4321)
        self.assertEqual(estado["porta"], 8443)
        self.assertIn("simulador.jar", estado["jar"])

    def test_ler_estado_sem_arquivo_retorna_vazio(self):
        with tempfile.TemporaryDirectory() as tmp:
            arquivo = Path(tmp) / "nao-existe.json"
            with patch("simulador.ARQUIVO_ESTADO_SIMULADOR", arquivo):
                self.assertEqual(ler_estado(), {})


class TestStatusEParada(unittest.TestCase):
    @patch("simulador._requisicao_json", return_value={"status": "sucesso", "versao": "1.0.0"})
    def test_em_execucao_quando_info_responde_sucesso(self, mock_request):
        self.assertTrue(simulador_em_execucao(8443))
        mock_request.assert_called_once_with("/api/info", 8443)

    @patch("simulador._requisicao_json", side_effect=RuntimeError("offline"))
    def test_em_execucao_falso_quando_offline(self, mock_request):
        self.assertFalse(simulador_em_execucao(8443))

    @patch("simulador._requisicao_json", return_value={"status": "sucesso"})
    def test_parar_remove_estado(self, mock_request):
        with tempfile.TemporaryDirectory() as tmp:
            arquivo = Path(tmp) / "simulador-server.json"
            arquivo.write_text("{}", encoding="utf-8")
            with patch("simulador.ARQUIVO_ESTADO_SIMULADOR", arquivo):
                resposta = parar(8443)
                ainda_existe = arquivo.exists()
        self.assertEqual(resposta["status"], "sucesso")
        self.assertFalse(ainda_existe)


class TestIniciarServidor(unittest.TestCase):
    @patch("simulador.registrar_estado")
    @patch("simulador.subprocess.Popen")
    @patch("simulador.encontrar_java", return_value="java")
    @patch("simulador._resolver_jar", return_value="/fake/simulador.jar")
    @patch("simulador.porta_disponivel", return_value=True)
    @patch("simulador.simulador_em_execucao", side_effect=[False, True])
    def test_iniciar_chama_popen_e_registra_estado(
        self, mock_online, mock_porta, mock_resolver, mock_java, mock_popen, mock_registrar
    ):
        processo = MagicMock()
        processo.pid = 9999
        processo.poll.return_value = None
        mock_popen.return_value = processo

        iniciar(jar_path="/fake/simulador.jar", porta=8443)

        comando = mock_popen.call_args.args[0]
        self.assertEqual(
            comando,
            ["java", "-jar", "/fake/simulador.jar", "--server.port=8443"],
        )
        mock_registrar.assert_called_once_with(9999, 8443, "/fake/simulador.jar")

    @patch("simulador.porta_disponivel", return_value=False)
    @patch("simulador.simulador_em_execucao", return_value=False)
    def test_iniciar_falha_quando_porta_em_uso(self, mock_online, mock_porta):
        with self.assertRaises(RuntimeError) as ctx:
            iniciar(jar_path="/fake/simulador.jar", porta=8443)
        self.assertIn("8443", str(ctx.exception))


class TestFormatar(unittest.TestCase):
    def test_formatar_sucesso_inclui_ok(self):
        saida = formatar({"status": "sucesso", "mensagem": "Tudo certo", "versao": "1.0.0"})
        self.assertIn("[OK]", saida)
        self.assertIn("SUCESSO", saida)
        self.assertIn("Tudo certo", saida)
        self.assertIn("1.0.0", saida)

    def test_formatar_parado_usa_info(self):
        saida = formatar({"status": "parado", "mensagem": "Simulador nao esta em execucao"})
        self.assertIn("[INFO]", saida)
        self.assertIn("PARADO", saida)


if __name__ == "__main__":
    unittest.main()
