package com.runner.assinador;

import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.annotation.JsonProperty;

/**
 * DTO (Data Transfer Object) que representa a resposta do Assinador.
 *
 * <p>Serializado como JSON para comunicação com o CLI.</p>
 */
@JsonInclude(JsonInclude.Include.NON_NULL)
public class RespostaAssinatura {

    @JsonProperty("status")
    private final String status;

    @JsonProperty("mensagem")
    private final String mensagem;

    @JsonProperty("assinatura")
    private final String assinatura;

    @JsonProperty("timestamp")
    private final String timestamp;

    @JsonProperty("certificado")
    private final String certificado;

    @JsonProperty("algoritmo")
    private final String algoritmo;

    /**
     * Construtor completo para criação da resposta.
     *
     * @param status      status da operação ("sucesso" ou "falha")
     * @param mensagem    mensagem descritiva do resultado
     * @param assinatura  assinatura digital em Base64 (pode ser null para validação)
     * @param timestamp   momento da operação em formato ISO-8601
     * @param certificado identificador do certificado utilizado (pode ser null)
     * @param algoritmo   algoritmo de assinatura utilizado
     */
    public RespostaAssinatura(String status, String mensagem, String assinatura,
                               String timestamp, String certificado, String algoritmo) {
        this.status = status;
        this.mensagem = mensagem;
        this.assinatura = assinatura;
        this.timestamp = timestamp;
        this.certificado = certificado;
        this.algoritmo = algoritmo;
    }

    public String getStatus() {
        return status;
    }

    public String getMensagem() {
        return mensagem;
    }

    public String getAssinatura() {
        return assinatura;
    }

    public String getTimestamp() {
        return timestamp;
    }

    public String getCertificado() {
        return certificado;
    }

    public String getAlgoritmo() {
        return algoritmo;
    }
}
