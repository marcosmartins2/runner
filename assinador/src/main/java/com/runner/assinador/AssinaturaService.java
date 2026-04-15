package com.runner.assinador;

import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;
import java.time.Instant;
import java.util.Base64;

/**
 * Service responsible for simulating digital signature operations.
 *
 * <p>The implementation is intentionally simple: it does not perform real
 * cryptography, but returns deterministic values that can be validated later.</p>
 */
public class AssinaturaService {

    /**
     * Simulates creation of a digital signature for the provided document.
     *
     * @param parametros validated parameters containing document and certificate
     * @return response with the simulated signature
     */
    public RespostaAssinatura criarAssinatura(ParametrosEntrada parametros) {
        String assinaturaSimulada = gerarAssinaturaSimulada(parametros.getDocumento());

        return new RespostaAssinatura(
                "sucesso",
                "Assinatura digital criada com sucesso (simulacao).",
                assinaturaSimulada,
                Instant.now().toString(),
                parametros.getCertificado(),
                "SHA256withRSA"
        );
    }

    /**
     * Simulates validation of a digital signature.
     *
     * <p>The signature is considered valid only when it matches the deterministic
     * value expected for the provided document.</p>
     *
     * @param parametros validated parameters containing document and signature
     * @return response with the validation result
     */
    public RespostaAssinatura validarAssinatura(ParametrosEntrada parametros) {
        String assinaturaEsperada = gerarAssinaturaSimulada(parametros.getDocumento());
        boolean valida = assinaturaEsperada.equals(parametros.getAssinatura());

        String status = valida ? "sucesso" : "falha";
        String mensagem = valida
                ? "Assinatura digital valida (simulacao)."
                : "Assinatura digital invalida para o documento informado (simulacao).";

        return new RespostaAssinatura(
                status,
                mensagem,
                parametros.getAssinatura(),
                Instant.now().toString(),
                null,
                "SHA256withRSA"
        );
    }

    /**
     * Generates a deterministic Base64 value that represents a simulated
     * signature for the provided document.
     */
    private String gerarAssinaturaSimulada(String documentoBase64) {
        try {
            MessageDigest digest = MessageDigest.getInstance("SHA-256");
            String conteudo = "runner-assinatura:" + documentoBase64;
            byte[] hash = digest.digest(conteudo.getBytes(StandardCharsets.UTF_8));
            return Base64.getEncoder().encodeToString(hash);
        } catch (NoSuchAlgorithmException e) {
            throw new RuntimeException("Algoritmo SHA-256 nao disponivel", e);
        }
    }
}
