"""
Testes do provisionador automatico de JDK/JRE.

Cobre a deteccao da plataforma, extracao da versao do `java -version`
e a resolucao das URLs do Eclipse Temurin (Adoptium) por plataforma.
"""

import unittest
from unittest import mock

import provisionador_jdk


class DetectarPlataformaTest(unittest.TestCase):
    def test_linux_x64(self):
        with mock.patch.object(provisionador_jdk.sys, "platform", "linux"):
            with mock.patch.object(provisionador_jdk.platform, "machine", return_value="x86_64"):
                self.assertEqual(provisionador_jdk.detectar_plataforma(), ("linux", "x64"))

    def test_windows_x64(self):
        with mock.patch.object(provisionador_jdk.sys, "platform", "win32"):
            with mock.patch.object(provisionador_jdk.platform, "machine", return_value="AMD64"):
                self.assertEqual(provisionador_jdk.detectar_plataforma(), ("windows", "x64"))

    def test_mac_arm(self):
        with mock.patch.object(provisionador_jdk.sys, "platform", "darwin"):
            with mock.patch.object(provisionador_jdk.platform, "machine", return_value="arm64"):
                self.assertEqual(provisionador_jdk.detectar_plataforma(), ("mac", "aarch64"))


class ExtrairVersaoTest(unittest.TestCase):
    def test_java_17(self):
        saida = 'openjdk version "17.0.9" 2025-10-15\n'
        self.assertEqual(provisionador_jdk.extrair_versao(saida), 17)

    def test_java_21(self):
        saida = 'openjdk version "21" 2026-03-19\n'
        self.assertEqual(provisionador_jdk.extrair_versao(saida), 21)

    def test_saida_invalida(self):
        self.assertEqual(provisionador_jdk.extrair_versao("texto qualquer"), 0)


class UrlDownloadJreTest(unittest.TestCase):
    def test_url_linux(self):
        with mock.patch.object(
            provisionador_jdk, "detectar_plataforma", return_value=("linux", "x64")
        ):
            url = provisionador_jdk.url_download_jre()
            self.assertIn("linux/x64/jre/hotspot", url)
            self.assertIn("adoptium.net", url)

    def test_plataforma_nao_suportada(self):
        with mock.patch.object(
            provisionador_jdk, "detectar_plataforma", return_value=("freebsd", "x64")
        ):
            with self.assertRaises(RuntimeError):
                provisionador_jdk.url_download_jre()


if __name__ == "__main__":
    unittest.main()
