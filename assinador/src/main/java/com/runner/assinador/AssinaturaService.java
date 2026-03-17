package com.runner.assinador;

import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;
import java.time.Instant;
import java.util.Base64;
import java.util.UUID;

/**
 * Serviço responsável por simular operações de assinatura digital.
 *
 * <p>Não realiza criptografia real — retorna respostas pré-construídas
 * para fins de simulação, conforme definido na especificação.</p>
 */
public class AssinaturaService {

    /**
     * Simula a criação de uma assinatura digital.
     *
     * <p>Gera uma assinatura simulada baseada em um hash SHA-256 do documento,
     * retornando uma resposta pré-construída no formato esperado.</p>
     *
     * @param parametros parâmetros validados contendo documento e certificado
     * @return {@link RespostaAssinatura} com a assinatura simulada
     */
    public RespostaAssinatura criarAssinatura(ParametrosEntrada parametros) {
        // Gerar assinatura simulada baseada no documento
        String assinaturaSimulada = gerarAssinaturaSimulada(parametros.getDocumento());

        return new RespostaAssinatura(
                "sucesso",
                "Assinatura digital criada com sucesso (simulação).",
                assinaturaSimulada,
                Instant.now().toString(),
                parametros.getCertificado(),
                "SHA256withRSA"
        );
    }

    /**
     * Simula a validação de uma assinatura digital.
     *
     * <p>Retorna um resultado pré-determinado de validação. Na simulação,
     * qualquer assinatura em Base64 válido é considerada válida.</p>
     *
     * @param parametros parâmetros validados contendo documento e assinatura
     * @return {@link RespostaAssinatura} com o resultado da validação
     */
    public RespostaAssinatura validarAssinatura(ParametrosEntrada parametros) {
        // Na simulação, toda assinatura válida (Base64) é aceita
        boolean valida = parametros.getAssinatura() != null
                && !parametros.getAssinatura().isBlank();

        String status = valida ? "sucesso" : "falha";
        String mensagem = valida
                ? "Assinatura digital válida (simulação)."
                : "Assinatura digital inválida (simulação).";

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
     * Gera uma string Base64 simulando uma assinatura digital.
     * Usa SHA-256 do documento + UUID para produzir um valor determinístico
     * mas único por execução.
     */
    private String gerarAssinaturaSimulada(String documentoBase64) {
        try {
            MessageDigest digest = MessageDigest.getInstance("SHA-256");
            String conteudo = documentoBase64 + ":" + UUID.randomUUID().toString();
            byte[] hash = digest.digest(conteudo.getBytes(StandardCharsets.UTF_8));
            return Base64.getEncoder().encodeToString(hash);
        } catch (NoSuchAlgorithmException e) {
            // SHA-256 é garantido pela especificação Java, não deve ocorrer
            throw new RuntimeException("Algoritmo SHA-256 não disponível", e);
        }
    }
}
