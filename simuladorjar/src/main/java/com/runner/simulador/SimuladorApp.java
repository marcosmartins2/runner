package com.runner.simulador;

import com.sun.net.httpserver.HttpExchange;
import com.sun.net.httpserver.HttpServer;

import java.io.IOException;
import java.io.OutputStream;
import java.net.InetSocketAddress;
import java.nio.charset.StandardCharsets;
import java.time.Instant;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

/**
 * Simulador HTTP minimo do HubSaude.
 *
 * <p>Aceita --server.port=N como Spring (compatibilidade com o CLI antigo) ou
 * "--port N". Expoe /api/info e /shutdown.</p>
 */
public class SimuladorApp {

    public static final String VERSAO = "1.0.0";
    private static final int PORTA_PADRAO = 8443;

    public static void main(String[] args) throws IOException {
        int porta = obterPorta(args);
        new SimuladorApp().iniciar(porta);
    }

    private HttpServer servidor;
    private ExecutorService executor;
    private final long iniciadoEm = System.currentTimeMillis();

    public void iniciar(int porta) throws IOException {
        servidor = HttpServer.create(new InetSocketAddress(porta), 0);
        servidor.createContext("/api/info", this::info);
        servidor.createContext("/shutdown", this::shutdown);
        executor = Executors.newCachedThreadPool();
        servidor.setExecutor(executor);
        servidor.start();
        System.out.println("Simulador HubSaude na porta " + porta);
    }

    private void info(HttpExchange exchange) throws IOException {
        if (!metodo(exchange, "GET")) return;
        long uptime = System.currentTimeMillis() - iniciadoEm;
        String corpo = "{"
                + "\"status\":\"sucesso\","
                + "\"mensagem\":\"Simulador HubSaude em execucao.\","
                + "\"versao\":\"" + VERSAO + "\","
                + "\"porta\":" + exchange.getLocalAddress().getPort() + ","
                + "\"uptimeMillis\":" + uptime + ","
                + "\"timestamp\":\"" + Instant.now() + "\""
                + "}";
        responder(exchange, 200, corpo);
    }

    private void shutdown(HttpExchange exchange) throws IOException {
        if (!metodo(exchange, "POST")) return;
        responder(exchange, 200, "{\"status\":\"sucesso\",\"mensagem\":\"Simulador sera interrompido.\"}");
        Thread t = new Thread(() -> {
            try { Thread.sleep(200); } catch (InterruptedException e) { Thread.currentThread().interrupt(); }
            encerrar();
        });
        t.setDaemon(false);
        t.start();
    }

    public void encerrar() {
        if (servidor != null) servidor.stop(0);
        if (executor != null) executor.shutdownNow();
    }

    private boolean metodo(HttpExchange exchange, String esperado) throws IOException {
        if (esperado.equals(exchange.getRequestMethod())) return true;
        exchange.getResponseHeaders().add("Allow", esperado);
        responder(exchange, 405, "{\"status\":\"falha\",\"mensagem\":\"metodo nao permitido\"}");
        return false;
    }

    private void responder(HttpExchange exchange, int status, String corpo) throws IOException {
        byte[] bytes = corpo.getBytes(StandardCharsets.UTF_8);
        exchange.getResponseHeaders().set("Content-Type", "application/json; charset=utf-8");
        exchange.sendResponseHeaders(status, bytes.length);
        try (OutputStream os = exchange.getResponseBody()) {
            os.write(bytes);
        }
    }

    static int obterPorta(String[] args) {
        for (int i = 0; i < args.length; i++) {
            String arg = args[i];
            if (arg.startsWith("--server.port=")) {
                return Integer.parseInt(arg.substring("--server.port=".length()));
            }
            if ((arg.equals("--port") || arg.equals("--porta")) && i + 1 < args.length) {
                return Integer.parseInt(args[i + 1]);
            }
        }
        return PORTA_PADRAO;
    }
}
