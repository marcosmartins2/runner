package com.runner.assinador;

import com.fasterxml.jackson.core.type.TypeReference;
import com.sun.net.httpserver.HttpExchange;
import com.sun.net.httpserver.HttpServer;

import java.io.IOException;
import java.io.OutputStream;
import java.net.InetSocketAddress;
import java.nio.charset.StandardCharsets;
import java.util.LinkedHashMap;
import java.util.Map;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

/**
 * Servidor HTTP minimo para permitir chamadas "warm start" ao Assinador.
 */
public class AssinadorHttpServer {

    private final int porta;
    private HttpServer servidor;
    private ExecutorService executor;

    public AssinadorHttpServer(int porta) {
        this.porta = porta;
    }

    public void iniciar() throws IOException {
        servidor = HttpServer.create(new InetSocketAddress(porta), 0);
        servidor.createContext("/api/info", this::info);
        servidor.createContext("/api/sign", this::sign);
        servidor.createContext("/api/validate", this::validate);
        servidor.createContext("/shutdown", this::shutdown);
        executor = Executors.newCachedThreadPool();
        servidor.setExecutor(executor);
        servidor.start();

        System.out.println("Assinador HTTP em execucao na porta " + porta);
    }

    private void info(HttpExchange exchange) throws IOException {
        if (!permitirMetodo(exchange, "GET")) {
            return;
        }

        Map<String, Object> resposta = new LinkedHashMap<>();
        resposta.put("status", "sucesso");
        resposta.put("mensagem", "Assinador em execucao.");
        resposta.put("porta", porta);
        responderJson(exchange, 200, resposta);
    }

    private void sign(HttpExchange exchange) throws IOException {
        if (!permitirMetodo(exchange, "POST")) {
            return;
        }

        try {
            Map<String, String> corpo = lerCorpoJson(exchange);
            String[] args = {
                    "--operacao", "criar",
                    "--documento", valor(corpo, "documento"),
                    "--certificado", valor(corpo, "certificado")
            };
            RespostaAssinatura resposta = AssinadorApp.executar(ValidadorParametros.validar(args));
            responderJson(exchange, 200, resposta);
        } catch (IllegalArgumentException e) {
            responderErro(exchange, 400, e.getMessage());
        } catch (Exception e) {
            responderErro(exchange, 500, "Erro inesperado: " + e.getMessage());
        }
    }

    private void validate(HttpExchange exchange) throws IOException {
        if (!permitirMetodo(exchange, "POST")) {
            return;
        }

        try {
            Map<String, String> corpo = lerCorpoJson(exchange);
            String[] args = {
                    "--operacao", "validar",
                    "--documento", valor(corpo, "documento"),
                    "--assinatura", valor(corpo, "assinatura")
            };
            RespostaAssinatura resposta = AssinadorApp.executar(ValidadorParametros.validar(args));
            responderJson(exchange, 200, resposta);
        } catch (IllegalArgumentException e) {
            responderErro(exchange, 400, e.getMessage());
        } catch (Exception e) {
            responderErro(exchange, 500, "Erro inesperado: " + e.getMessage());
        }
    }

    private void shutdown(HttpExchange exchange) throws IOException {
        if (!permitirMetodo(exchange, "POST")) {
            return;
        }

        Map<String, Object> resposta = new LinkedHashMap<>();
        resposta.put("status", "sucesso");
        resposta.put("mensagem", "Assinador sera interrompido.");
        responderJson(exchange, 200, resposta);

        Thread interrupcao = new Thread(() -> {
            try {
                Thread.sleep(200);
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
            }
            servidor.stop(0);
            executor.shutdownNow();
        });
        interrupcao.setDaemon(false);
        interrupcao.start();
    }

    private boolean permitirMetodo(HttpExchange exchange, String metodo) throws IOException {
        if (metodo.equals(exchange.getRequestMethod())) {
            return true;
        }
        exchange.getResponseHeaders().add("Allow", metodo);
        responderErro(exchange, 405, "Metodo nao permitido. Use " + metodo + ".");
        return false;
    }

    private Map<String, String> lerCorpoJson(HttpExchange exchange) throws IOException {
        String body = new String(exchange.getRequestBody().readAllBytes(), StandardCharsets.UTF_8);
        try {
            return AssinadorJson.mapper().readValue(body, new TypeReference<Map<String, String>>() {});
        } catch (IOException e) {
            throw new IllegalArgumentException("Corpo JSON invalido.");
        }
    }

    private String valor(Map<String, String> mapa, String chave) {
        return mapa.getOrDefault(chave, "");
    }

    private void responderErro(HttpExchange exchange, int statusCode, String mensagem) throws IOException {
        Map<String, Object> resposta = new LinkedHashMap<>();
        resposta.put("status", "falha");
        resposta.put("mensagem", mensagem);
        responderJson(exchange, statusCode, resposta);
    }

    private void responderJson(HttpExchange exchange, int statusCode, Object resposta) throws IOException {
        try {
            byte[] bytes = AssinadorApp.serializarResposta(resposta).getBytes(StandardCharsets.UTF_8);
            exchange.getResponseHeaders().set("Content-Type", "application/json; charset=utf-8");
            exchange.sendResponseHeaders(statusCode, bytes.length);
            try (OutputStream os = exchange.getResponseBody()) {
                os.write(bytes);
            }
        } catch (Exception e) {
            byte[] bytes = ("{\"status\":\"falha\",\"mensagem\":\"Falha ao serializar resposta\"}")
                    .getBytes(StandardCharsets.UTF_8);
            exchange.sendResponseHeaders(500, bytes.length);
            try (OutputStream os = exchange.getResponseBody()) {
                os.write(bytes);
            }
        }
    }
}
