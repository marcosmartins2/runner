package com.runner.assinador;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.SerializationFeature;

/**
 * Ponto de entrada da aplicação Assinador.
 *
 * <p>Recebe parâmetros via linha de comandos, valida-os e despacha
 * para o serviço de assinatura apropriado.</p>
 *
 * <h3>Uso:</h3>
 * <pre>
 *   java -jar assinador.jar --operacao criar --documento &lt;doc&gt; --certificado &lt;cert&gt;
 *   java -jar assinador.jar --operacao validar --documento &lt;doc&gt; --assinatura &lt;sig&gt;
 * </pre>
 */
public class AssinadorApp {

    private static final ObjectMapper mapper = new ObjectMapper()
            .enable(SerializationFeature.INDENT_OUTPUT);

    public static void main(String[] args) {
        try {
            // Validar parâmetros de entrada
            ParametrosEntrada parametros = ValidadorParametros.validar(args);

            // Executar operação
            AssinaturaService service = new AssinaturaService();
            RespostaAssinatura resposta;

            switch (parametros.getOperacao()) {
                case "criar":
                    resposta = service.criarAssinatura(parametros);
                    break;
                case "validar":
                    resposta = service.validarAssinatura(parametros);
                    break;
                default:
                    throw new IllegalArgumentException(
                            "Operação desconhecida: " + parametros.getOperacao()
                                    + ". Operações válidas: criar, validar");
            }

            // Serializar resposta como JSON
            String json = mapper.writeValueAsString(resposta);
            System.out.println(json);
            System.exit(0);

        } catch (IllegalArgumentException e) {
            System.err.println("ERRO DE PARÂMETRO: " + e.getMessage());
            imprimirUso();
            System.exit(1);

        } catch (Exception e) {
            System.err.println("ERRO INESPERADO: " + e.getMessage());
            System.exit(2);
        }
    }

    /**
     * Exibe instruções de uso da aplicação.
     */
    private static void imprimirUso() {
        System.err.println();
        System.err.println("Uso:");
        System.err.println("  Criar assinatura:");
        System.err.println("    java -jar assinador.jar --operacao criar --documento <base64> --certificado <id>");
        System.err.println();
        System.err.println("  Validar assinatura:");
        System.err.println("    java -jar assinador.jar --operacao validar --documento <base64> --assinatura <base64>");
        System.err.println();
        System.err.println("Parâmetros:");
        System.err.println("  --operacao     Operação a realizar: 'criar' ou 'validar'");
        System.err.println("  --documento    Documento em Base64 para assinar ou validar");
        System.err.println("  --certificado  Identificador do certificado digital (para criação)");
        System.err.println("  --assinatura   Assinatura em Base64 para validação");
    }
}
