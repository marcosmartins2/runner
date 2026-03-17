package com.runner.assinador;

/**
 * Representa os parâmetros de entrada validados para uma operação do Assinador.
 */
public class ParametrosEntrada {

    private final String operacao;
    private final String documento;
    private final String certificado;
    private final String assinatura;

    /**
     * Construtor para operação de criação de assinatura.
     */
    public ParametrosEntrada(String operacao, String documento, String certificado, String assinatura) {
        this.operacao = operacao;
        this.documento = documento;
        this.certificado = certificado;
        this.assinatura = assinatura;
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

    @Override
    public String toString() {
        return "ParametrosEntrada{" +
                "operacao='" + operacao + '\'' +
                ", documento='" + documento + '\'' +
                ", certificado='" + certificado + '\'' +
                ", assinatura='" + assinatura + '\'' +
                '}';
    }
}
