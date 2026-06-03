package com.runner.assinador;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertTrue;

@DisplayName("MonitorInatividade (auto-shutdown - criterio E2)")
class MonitorInatividadeTest {

    @Test
    @DisplayName("timeout <= 0 desativa o monitor")
    void timeoutZeroDesativa() {
        MonitorInatividade m = new MonitorInatividade(0, 0L);
        assertFalse(m.ativo());
        assertFalse(m.deveEncerrar(Long.MAX_VALUE));
    }

    @Test
    @DisplayName("encerra somente quando a inatividade atinge o timeout")
    void encerraAposTimeout() {
        // timeout de 1 minuto (60_000 ms), referencia inicial t=1000.
        MonitorInatividade m = new MonitorInatividade(1, 1_000L);
        assertTrue(m.ativo());
        assertFalse(m.deveEncerrar(1_000L + 59_999L)); // ainda dentro da janela
        assertTrue(m.deveEncerrar(1_000L + 60_000L));  // atingiu o timeout
    }

    @Test
    @DisplayName("cada requisicao reinicia o temporizador de inatividade")
    void requisicaoReiniciaTemporizador() {
        MonitorInatividade m = new MonitorInatividade(1, 0L);
        // Quase no limite sem nenhuma interacao nova.
        assertFalse(m.deveEncerrar(59_000L));
        // Uma requisicao chega em t=59_000 e reinicia a contagem.
        m.registrarInteracao(59_000L);
        assertEquals(59_000L, m.ultimaInteracaoMillis());
        // 59s apos a requisicao (t=118_000) ainda NAO deve encerrar.
        assertFalse(m.deveEncerrar(118_000L));
        // Somente 60s apos a ultima interacao deve encerrar.
        assertTrue(m.deveEncerrar(59_000L + 60_000L));
    }
}
