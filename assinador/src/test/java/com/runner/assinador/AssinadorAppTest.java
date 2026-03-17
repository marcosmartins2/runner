package com.runner.assinador;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;

import java.util.Base64;

import static org.junit.jupiter.api.Assertions.*;

/**
 * Testes unitários para o Assinador.
 *
 * <p>Cobre validação de parâmetros, criação simulada de assinatura,
 * validação simulada de assinatura e tratamento de erros.</p>
 */
class AssinadorAppTest {

    private static final String DOCUMENTO_VALIDO =
            Base64.getEncoder().encodeToString("documento de teste".getBytes());
    private static final String CERTIFICADO_VALIDO = "cert-001";
    private static final String ASSINATURA_VALIDA =
            Base64.getEncoder().encodeToString("assinatura simulada".getBytes());

    // ======================================================================
    // Testes de Validação de Parâmetros
    // ======================================================================

    @Nested
    @DisplayName("Validação de Parâmetros")
    class ValidacaoParametrosTest {

        @Test
        @DisplayName("Deve rejeitar quando nenhum parâmetro é fornecido")
        void deveRejeitarSemParametros() {
            IllegalArgumentException ex = assertThrows(
                    IllegalArgumentException.class,
                    () -> ValidadorParametros.validar(new String[]{})
            );
            assertTrue(ex.getMessage().contains("Nenhum parâmetro fornecido"));
        }

        @Test
        @DisplayName("Deve rejeitar quando --operacao não é informada")
        void deveRejeitarSemOperacao() {
            String[] args = {"--documento", DOCUMENTO_VALIDO};
            IllegalArgumentException ex = assertThrows(
                    IllegalArgumentException.class,
                    () -> ValidadorParametros.validar(args)
            );
            assertTrue(ex.getMessage().contains("--operacao"));
        }

        @Test
        @DisplayName("Deve rejeitar operação inválida")
        void deveRejeitarOperacaoInvalida() {
            String[] args = {"--operacao", "encriptar", "--documento", DOCUMENTO_VALIDO};
            IllegalArgumentException ex = assertThrows(
                    IllegalArgumentException.class,
                    () -> ValidadorParametros.validar(args)
            );
            assertTrue(ex.getMessage().contains("Operação inválida"));
            assertTrue(ex.getMessage().contains("encriptar"));
        }

        @Test
        @DisplayName("Deve rejeitar documento não Base64")
        void deveRejeitarDocumentoInvalido() {
            String[] args = {"--operacao", "criar", "--documento", "!!!inválido!!!", "--certificado", CERTIFICADO_VALIDO};
            IllegalArgumentException ex = assertThrows(
                    IllegalArgumentException.class,
                    () -> ValidadorParametros.validar(args)
            );
            assertTrue(ex.getMessage().contains("Base64"));
        }

        @Test
        @DisplayName("Deve rejeitar criação sem certificado")
        void deveRejeitarCriacaoSemCertificado() {
            String[] args = {"--operacao", "criar", "--documento", DOCUMENTO_VALIDO};
            IllegalArgumentException ex = assertThrows(
                    IllegalArgumentException.class,
                    () -> ValidadorParametros.validar(args)
            );
            assertTrue(ex.getMessage().contains("--certificado"));
        }

        @Test
        @DisplayName("Deve rejeitar validação sem assinatura")
        void deveRejeitarValidacaoSemAssinatura() {
            String[] args = {"--operacao", "validar", "--documento", DOCUMENTO_VALIDO};
            IllegalArgumentException ex = assertThrows(
                    IllegalArgumentException.class,
                    () -> ValidadorParametros.validar(args)
            );
            assertTrue(ex.getMessage().contains("--assinatura"));
        }

        @Test
        @DisplayName("Deve aceitar parâmetros válidos para criação")
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
        @DisplayName("Deve aceitar parâmetros válidos para validação")
        void deveAceitarParametrosValidosValidacao() {
            String[] args = {
                    "--operacao", "validar",
                    "--documento", DOCUMENTO_VALIDO,
                    "--assinatura", ASSINATURA_VALIDA
            };
            ParametrosEntrada p = ValidadorParametros.validar(args);
            assertEquals("validar", p.getOperacao());
            assertEquals(DOCUMENTO_VALIDO, p.getDocumento());
            assertEquals(ASSINATURA_VALIDA, p.getAssinatura());
        }
    }

    // ======================================================================
    // Testes do Serviço de Assinatura
    // ======================================================================

    @Nested
    @DisplayName("Serviço de Assinatura")
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
        @DisplayName("A assinatura simulada deve ser Base64 válido")
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
        @DisplayName("Deve validar assinatura com sucesso")
        void deveValidarAssinatura() {
            ParametrosEntrada params = new ParametrosEntrada(
                    "validar", DOCUMENTO_VALIDO, null, ASSINATURA_VALIDA
            );

            RespostaAssinatura resposta = service.validarAssinatura(params);

            assertEquals("sucesso", resposta.getStatus());
            assertTrue(resposta.getMensagem().contains("válida"));
            assertEquals("SHA256withRSA", resposta.getAlgoritmo());
        }
    }
}
