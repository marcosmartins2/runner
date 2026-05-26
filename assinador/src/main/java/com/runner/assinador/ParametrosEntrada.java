package com.runner.assinador;

/**
 * Representa os parâmetros de entrada validados para uma operação do Assinador.
 */
public class ParametrosEntrada {

    private final String operacao;
    private final String documento;
    private final String certificado;
    private final String assinatura;
    private final String pkcs11Provider;

    public ParametrosEntrada(String operacao, String documento, String certificado, String assinatura) {
        this(operacao, documento, certificado, assinatura, null);
    }

    public ParametrosEntrada(String operacao, String documento, String certificado,
                              String assinatura, String pkcs11Provider) {
        this.operacao = operacao;
        this.documento = documento;
        this.certificado = certificado;
        this.assinatura = assinatura;
        this.pkcs11Provider = pkcs11Provider;
    }

    public String getOperacao() {
        return operacao;
    }

    public String getDocumento() {
        return documento;
    }

    public String getCertificado() {
        return certificado;
    }

    public String getAssinatura() {
        return assinatura;
    }

    /**
     * Identificador do provider PKCS#11 (token/smart card) usado para simulacao
     * de assinatura. Quando null, o servico nao usa dispositivo criptografico.
     */
    public String getPkcs11Provider() {
        return pkcs11Provider;
    }

    @Override
    public String toString() {
        return "ParametrosEntrada{" +
                "operacao='" + operacao + '\'' +
                ", documento='" + documento + '\'' +
                ", certificado='" + certificado + '\'' +
                ", assinatura='" + assinatura + '\'' +
                ", pkcs11Provider='" + pkcs11Provider + '\'' +
                '}';
    }
}
