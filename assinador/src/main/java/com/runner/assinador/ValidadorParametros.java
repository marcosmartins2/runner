package com.runner.assinador;

import java.util.Base64;
import java.util.HashMap;
import java.util.Map;

/**
 * Responsável pela validação rigorosa dos parâmetros de entrada do Assinador.
 *
 * <p>Verifica a presença, o formato e a consistência dos parâmetros
 * antes que qualquer operação de assinatura seja executada.</p>
 */
public class ValidadorParametros {

    /**
     * Valida os argumentos de linha de comando e retorna um objeto
     * {@link ParametrosEntrada} com os valores validados.
     *
     * @param args argumentos de linha de comando
     * @return {@link ParametrosEntrada} com os parâmetros validados
     * @throws IllegalArgumentException se algum parâmetro for inválido ou ausente
     */
    public static ParametrosEntrada validar(String[] args) {
        if (args == null || args.length == 0) {
            throw new IllegalArgumentException(
                    "Nenhum parâmetro fornecido. É necessário informar ao menos --operacao.");
        }

        Map<String, String> parametros = parsearArgumentos(args);

        // Validar operação (obrigatória)
        String operacao = parametros.get("operacao");
        if (operacao == null || operacao.isBlank()) {
            throw new IllegalArgumentException(
                    "Parâmetro obrigatório '--operacao' não foi informado.");
        }

        operacao = operacao.trim().toLowerCase();
        if (!operacao.equals("criar") && !operacao.equals("validar")) {
            throw new IllegalArgumentException(
                    "Operação inválida: '" + operacao + "'. Valores aceitos: 'criar', 'validar'.");
        }

        // Validar documento (obrigatório para ambas as operações)
        String documento = parametros.get("documento");
        if (documento == null || documento.isBlank()) {
            throw new IllegalArgumentException(
                    "Parâmetro obrigatório '--documento' não foi informado. "
                            + "Forneça o documento codificado em Base64.");
        }
        validarBase64(documento, "documento");

        String certificado = parametros.get("certificado");
        String assinatura = parametros.get("assinatura");

        // Validações específicas por operação
        if (operacao.equals("criar")) {
            if (certificado == null || certificado.isBlank()) {
                throw new IllegalArgumentException(
                        "Parâmetro '--certificado' é obrigatório para a operação 'criar'. "
                                + "Forneça o identificador do certificado digital.");
            }
        } else if (operacao.equals("validar")) {
            if (assinatura == null || assinatura.isBlank()) {
                throw new IllegalArgumentException(
                        "Parâmetro '--assinatura' é obrigatório para a operação 'validar'. "
                                + "Forneça a assinatura codificada em Base64.");
            }
            validarBase64(assinatura, "assinatura");
        }

        // Parametro opcional: provider PKCS#11 (token/smart card).
        String pkcs11 = parametros.get("pkcs11");
        if (pkcs11 != null && pkcs11.isBlank()) {
            pkcs11 = null;
        }

        return new ParametrosEntrada(operacao, documento, certificado, assinatura, pkcs11);
    }

    /**
     * Parseia argumentos no formato --chave valor para um mapa.
     */
    private static Map<String, String> parsearArgumentos(String[] args) {
        Map<String, String> mapa = new HashMap<>();

        for (int i = 0; i < args.length; i++) {
            if (args[i].startsWith("--")) {
                String chave = args[i].substring(2);
                if (i + 1 < args.length && !args[i + 1].startsWith("--")) {
                    mapa.put(chave, args[i + 1]);
                    i++; // pular o valor
                } else {
                    mapa.put(chave, "");
                }
            } else {
                throw new IllegalArgumentException(
                        "Argumento inesperado: '" + args[i]
                                + "'. Os parâmetros devem começar com '--'.");
            }
        }

        return mapa;
    }

    /**
     * Valida se uma string é uma codificação Base64 válida.
     *
     * @param valor     o valor a ser validado
     * @param nomeParam nome do parâmetro (para mensagem de erro)
     * @throws IllegalArgumentException se o valor não for Base64 válido
     */
    private static void validarBase64(String valor, String nomeParam) {
        try {
            Base64.getDecoder().decode(valor);
        } catch (IllegalArgumentException e) {
            throw new IllegalArgumentException(
                    "O parâmetro '--" + nomeParam + "' não contém um valor Base64 válido: '"
                            + valor + "'. Verifique a codificação do dado.");
        }
    }
}
