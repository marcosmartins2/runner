package com.runner.simulador;

import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import java.io.IOException;
import java.net.HttpURLConnection;
import java.net.ServerSocket;
import java.net.URI;
import java.net.URL;
import java.nio.charset.StandardCharsets;

import static org.junit.jupiter.api.Assertions.*;

class SimuladorAppTest {

    private SimuladorApp app;
    private int porta;

    @BeforeEach
    void inicializar() throws IOException {
        try (ServerSocket ss = new ServerSocket(0)) {
            porta = ss.getLocalPort();
        }
        app = new SimuladorApp();
        app.iniciar(porta);
    }

    @AfterEach
    void finalizar() {
        app.encerrar();
    }

    @Test
    void infoRetornaJsonComStatusSucesso() throws Exception {
        String corpo = chamar("GET", "/api/info");
        assertTrue(corpo.contains("\"status\":\"sucesso\""), corpo);
        assertTrue(corpo.contains("\"versao\":\"" + SimuladorApp.VERSAO + "\""), corpo);
    }

    @Test
    void shutdownRetornaSucessoEEncerra() throws Exception {
        String corpo = chamar("POST", "/shutdown");
        assertTrue(corpo.contains("\"status\":\"sucesso\""), corpo);
    }

    @Test
    void obterPortaAceitaServerPort() {
        assertEquals(9999, SimuladorApp.obterPorta(new String[]{"--server.port=9999"}));
    }

    @Test
    void obterPortaAceitaPort() {
        assertEquals(8080, SimuladorApp.obterPorta(new String[]{"--port", "8080"}));
    }

    @Test
    void obterPortaSemArgumentoUsaPadrao() {
        assertEquals(8443, SimuladorApp.obterPorta(new String[]{}));
    }

    private String chamar(String metodo, String path) throws Exception {
        URL url = URI.create("http://127.0.0.1:" + porta + path).toURL();
        HttpURLConnection conn = (HttpURLConnection) url.openConnection();
        conn.setRequestMethod(metodo);
        conn.setConnectTimeout(2000);
        conn.setReadTimeout(2000);
        int sc = conn.getResponseCode();
        assertTrue(sc >= 200 && sc < 300, "status HTTP inesperado: " + sc);
        byte[] bytes = conn.getInputStream().readAllBytes();
        return new String(bytes, StandardCharsets.UTF_8);
    }
}
