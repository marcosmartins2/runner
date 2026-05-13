package com.runner.assinador;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.SerializationFeature;

/**
 * Ponto de entrada da aplicacao Assinador.
 */
public class AssinadorApp {

    private static final ObjectMapper mapper = new ObjectMapper()
            .enable(SerializationFeature.INDENT_OUTPUT);
    private static final int PORTA_PADRAO = 8080;
    private static final int SEM_TIMEOUT_INATIVIDADE = 0;

    public static void main(String[] args) {
        try {
            if (deveIniciarServidor(args)) {
                int porta = obterPorta(args);
                int timeoutInatividadeMinutos = obterTimeoutInatividadeMinutos(args);
                new AssinadorHttpServer(porta, timeoutInatividadeMinutos).iniciar();
                return;
            }

            ParametrosEntrada parametros = ValidadorParametros.validar(args);
            RespostaAssinatura resposta = executar(parametros);

            System.out.println(serializarResposta(resposta));
            System.exit(0);

        } catch (IllegalArgumentException e) {
            System.err.println("ERRO DE PARAMETRO: " + e.getMessage());
            imprimirUso();
            System.exit(1);

        } catch (Exception e) {
            System.err.println("ERRO INESPERADO: " + e.getMessage());
            System.exit(2);
        }
    }

    public static RespostaAssinatura executar(ParametrosEntrada parametros) {
        AssinaturaService service = new AssinaturaService();

        switch (parametros.getOperacao()) {
            case "criar":
                return service.criarAssinatura(parametros);
            case "validar":
                return service.validarAssinatura(parametros);
            default:
                throw new IllegalArgumentException(
                        "Operacao desconhecida: " + parametros.getOperacao()
                                + ". Operacoes validas: criar, validar");
        }
    }

    public static String serializarResposta(Object resposta) throws Exception {
        return mapper.writeValueAsString(resposta);
    }

    private static boolean deveIniciarServidor(String[] args) {
        for (String arg : args) {
            if ("--server".equals(arg) || "--servidor".equals(arg)) {
                return true;
            }
        }
        return false;
    }

    private static int obterPorta(String[] args) {
        for (int i = 0; i < args.length; i++) {
            if (("--port".equals(args[i]) || "--porta".equals(args[i])) && i + 1 < args.length) {
                try {
                    int porta = Integer.parseInt(args[i + 1]);
                    if (porta < 1 || porta > 65535) {
                        throw new IllegalArgumentException("Porta fora do intervalo permitido: " + porta);
                    }
                    return porta;
                } catch (NumberFormatException e) {
                    throw new IllegalArgumentException("Porta invalida: " + args[i + 1]);
                }
            }
        }
        return PORTA_PADRAO;
    }

    private static int obterTimeoutInatividadeMinutos(String[] args) {
        for (int i = 0; i < args.length; i++) {
            if (("--idle-timeout-minutes".equals(args[i]) || "--parar-apos-minutos".equals(args[i]))
                    && i + 1 < args.length) {
                try {
                    int minutos = Integer.parseInt(args[i + 1]);
                    if (minutos < 0) {
                        throw new IllegalArgumentException(
                                "Timeout de inatividade nao pode ser negativo: " + minutos);
                    }
                    return minutos;
                } catch (NumberFormatException e) {
                    throw new IllegalArgumentException(
                            "Timeout de inatividade invalido: " + args[i + 1]);
                }
            }
        }
        return SEM_TIMEOUT_INATIVIDADE;
    }

    private static void imprimirUso() {
        System.err.println();
        System.err.println("Uso:");
        System.err.println("  Criar assinatura:");
        System.err.println("    java -jar assinador.jar --operacao criar --documento <base64> --certificado <id>");
        System.err.println();
        System.err.println("  Validar assinatura:");
        System.err.println("    java -jar assinador.jar --operacao validar --documento <base64> --assinatura <base64>");
        System.err.println();
        System.err.println("  Modo servidor:");
        System.err.println("    java -jar assinador.jar --server --port 8080");
        System.err.println("    java -jar assinador.jar --server --parar-apos-minutos 30");
        System.err.println();
        System.err.println("Parametros:");
        System.err.println("  --operacao     Operacao a realizar: 'criar' ou 'validar'");
        System.err.println("  --documento    Documento em Base64 para assinar ou validar");
        System.err.println("  --certificado  Identificador do certificado digital (para criacao)");
        System.err.println("  --assinatura   Assinatura em Base64 para validacao");
        System.err.println("  --server       Inicia o Assinador em modo servidor HTTP");
        System.err.println("  --port         Porta do servidor HTTP (padrao: 8080)");
        System.err.println("  --parar-apos-minutos  Encerra o servidor apos N minutos sem interacao (0 desativa)");
    }
}
