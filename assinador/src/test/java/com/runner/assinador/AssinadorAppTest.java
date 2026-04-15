package com.runner.assinador;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;

import java.util.Base64;

import static org.junit.jupiter.api.Assertions.assertDoesNotThrow;
import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;

/**
 * Unit tests for the signer simulation.
 */
class AssinadorAppTest {

    private static final String DOCUMENTO_VALIDO =
            Base64.getEncoder().encodeToString("documento de teste".getBytes());
    private static final String CERTIFICADO_VALIDO = "cert-001";
    private static final String ASSINATURA_DIFERENTE =
            Base64.getEncoder().encodeToString("assinatura simulada".getBytes());

    @Nested
    @DisplayName("Validacao de parametros")
    class ValidacaoParametrosTest {

        @Test
        @DisplayName("Deve rejeitar quando nenhum parametro e fornecido")
        void deveRejeitarSemParametros() {
            IllegalArgumentException ex = assertThrows(
                    IllegalArgumentException.class,
                    () -> ValidadorParametros.validar(new String[]{})
            );
            assertTrue(ex.getMessage().contains("Nenhum"));
        }

        @Test
        @DisplayName("Deve rejeitar quando --operacao nao e informada")
        void deveRejeitarSemOperacao() {
            String[] args = {"--documento", DOCUMENTO_VALIDO};
            IllegalArgumentException ex = assertThrows(
                    IllegalArgumentException.class,
                    () -> ValidadorParametros.validar(args)
            );
            assertTrue(ex.getMessage().contains("--operacao"));
        }

        @Test
        @DisplayName("Deve rejeitar operacao invalida")
        void deveRejeitarOperacaoInvalida() {
            String[] args = {"--operacao", "encriptar", "--documento", DOCUMENTO_VALIDO};
            IllegalArgumentException ex = assertThrows(
                    IllegalArgumentException.class,
                    () -> ValidadorParametros.validar(args)
            );
            assertTrue(ex.getMessage().contains("Opera"));
            assertTrue(ex.getMessage().contains("encriptar"));
        }

        @Test
        @DisplayName("Deve rejeitar documento nao Base64")
        void deveRejeitarDocumentoInvalido() {
            String[] args = {
                    "--operacao", "criar",
                    "--documento", "!!!invalido!!!",
                    "--certificado", CERTIFICADO_VALIDO
            };
            IllegalArgumentException ex = assertThrows(
                    IllegalArgumentException.class,
                    () -> ValidadorParametros.validar(args)
            );
            assertTrue(ex.getMessage().contains("Base64"));
        }

        @Test
        @DisplayName("Deve rejeitar criacao sem certificado")
        void deveRejeitarCriacaoSemCertificado() {
            String[] args = {"--operacao", "criar", "--documento", DOCUMENTO_VALIDO};
            IllegalArgumentException ex = assertThrows(
                    IllegalArgumentException.class,
                    () -> ValidadorParametros.validar(args)
            );
            assertTrue(ex.getMessage().contains("--certificado"));
        }

        @Test
        @DisplayName("Deve rejeitar validacao sem assinatura")
        void deveRejeitarValidacaoSemAssinatura() {
            String[] args = {"--operacao", "validar", "--documento", DOCUMENTO_VALIDO};
            IllegalArgumentException ex = assertThrows(
                    IllegalArgumentException.class,
                    () -> ValidadorParametros.validar(args)
            );
            assertTrue(ex.getMessage().contains("--assinatura"));
        }

        @Test
        @DisplayName("Deve aceitar parametros validos para criacao")
        void deveAceitarParametrosValidosCriacao() {
            String[] args = {
                    "--operacao", "criar",
                    "--documento", DOCUMENTO_VALIDO,
                    "--certificado", CERTIFICADO_VALIDO
            };
            ParametrosEntrada p = ValidadorParametros.validar(args);
            assertEquals("criar", p.getOperacao());
            assertEquals(DOCUMENTO_VALIDO, p.getDocumento());
            assertEquals(CERTIFICADO_VALIDO, p.getCertificado());
        }

        @Test
        @DisplayName("Deve aceitar parametros validos para validacao")
        void deveAceitarParametrosValidosValidacao() {
            String[] args = {
                    "--operacao", "validar",
                    "--documento", DOCUMENTO_VALIDO,
                    "--assinatura", ASSINATURA_DIFERENTE
            };
            ParametrosEntrada p = ValidadorParametros.validar(args);
            assertEquals("validar", p.getOperacao());
            assertEquals(DOCUMENTO_VALIDO, p.getDocumento());
            assertEquals(ASSINATURA_DIFERENTE, p.getAssinatura());
        }
    }

    @Nested
    @DisplayName("Servico de assinatura")
    class AssinaturaServiceTest {

        private final AssinaturaService service = new AssinaturaService();

        @Test
        @DisplayName("Deve criar assinatura simulada com sucesso")
        void deveCriarAssinatura() {
            ParametrosEntrada params = new ParametrosEntrada(
                    "criar", DOCUMENTO_VALIDO, CERTIFICADO_VALIDO, null
            );

            RespostaAssinatura resposta = service.criarAssinatura(params);

            assertEquals("sucesso", resposta.getStatus());
            assertNotNull(resposta.getAssinatura());
            assertNotNull(resposta.getTimestamp());
            assertEquals(CERTIFICADO_VALIDO, resposta.getCertificado());
            assertEquals("SHA256withRSA", resposta.getAlgoritmo());
            assertTrue(resposta.getMensagem().contains("criada com sucesso"));
        }

        @Test
        @DisplayName("A assinatura simulada deve ser Base64 valido")
        void assinaturaDeveSerBase64Valido() {
            ParametrosEntrada params = new ParametrosEntrada(
                    "criar", DOCUMENTO_VALIDO, CERTIFICADO_VALIDO, null
            );

            RespostaAssinatura resposta = service.criarAssinatura(params);

            assertDoesNotThrow(() ->
                    Base64.getDecoder().decode(resposta.getAssinatura())
            );
        }

        @Test
        @DisplayName("Deve validar assinatura gerada para o mesmo documento")
        void deveValidarAssinaturaGerada() {
            ParametrosEntrada criacao = new ParametrosEntrada(
                    "criar", DOCUMENTO_VALIDO, CERTIFICADO_VALIDO, null
            );
            String assinaturaGerada = service.criarAssinatura(criacao).getAssinatura();

            ParametrosEntrada validacao = new ParametrosEntrada(
                    "validar", DOCUMENTO_VALIDO, null, assinaturaGerada
            );

            RespostaAssinatura resposta = service.validarAssinatura(validacao);

            assertEquals("sucesso", resposta.getStatus());
            assertTrue(resposta.getMensagem().contains("valida"));
            assertEquals("SHA256withRSA", resposta.getAlgoritmo());
        }

        @Test
        @DisplayName("Deve rejeitar assinatura que nao corresponde ao documento")
        void deveRejeitarAssinaturaIncompativel() {
            ParametrosEntrada params = new ParametrosEntrada(
                    "validar", DOCUMENTO_VALIDO, null, ASSINATURA_DIFERENTE
            );

            RespostaAssinatura resposta = service.validarAssinatura(params);

            assertEquals("falha", resposta.getStatus());
            assertTrue(resposta.getMensagem().contains("invalida"));
            assertEquals("SHA256withRSA", resposta.getAlgoritmo());
        }
    }
}
