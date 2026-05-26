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

    @JsonProperty("pkcs11Provider")
    private final String pkcs11Provider;

    public RespostaAssinatura(String status, String mensagem, String assinatura,
                               String timestamp, String certificado, String algoritmo) {
        this(status, mensagem, assinatura, timestamp, certificado, algoritmo, null);
    }

    public RespostaAssinatura(String status, String mensagem, String assinatura,
                               String timestamp, String certificado, String algoritmo,
                               String pkcs11Provider) {
        this.status = status;
        this.mensagem = mensagem;
        this.assinatura = assinatura;
        this.timestamp = timestamp;
        this.certificado = certificado;
        this.algoritmo = algoritmo;
        this.pkcs11Provider = pkcs11Provider;
    }

    public String getPkcs11Provider() {
        return pkcs11Provider;
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
