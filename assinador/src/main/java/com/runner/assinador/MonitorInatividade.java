package com.runner.assinador;

/**
 * Encapsula a politica de auto-encerramento por inatividade do servidor HTTP.
 *
 * <p>A logica e isolada e deterministica: recebe o instante atual como
 * parametro em vez de consultar o relogio internamente. Isso permite testes
 * sem dependencia de tempo real, comprovando que o temporizador reinicia a
 * cada requisicao (criterio E2 de docs/criterios.md).
 */
public final class MonitorInatividade {

    private final long timeoutMillis;
    private volatile long ultimaInteracaoMillis;

    /**
     * @param timeoutMinutos janela de inatividade; valores &lt;= 0 desativam o monitor.
     * @param agoraMillis instante de referencia inicial.
     */
    public MonitorInatividade(int timeoutMinutos, long agoraMillis) {
        this.timeoutMillis = timeoutMinutos > 0 ? timeoutMinutos * 60_000L : 0L;
        this.ultimaInteracaoMillis = agoraMillis;
    }

    /** Indica se o monitor esta ativo (timeout configurado). */
    public boolean ativo() {
        return timeoutMillis > 0;
    }

    /** Registra uma interacao, reiniciando a contagem de inatividade. */
    public void registrarInteracao(long agoraMillis) {
        this.ultimaInteracaoMillis = agoraMillis;
    }

    /** Retorna true quando a inatividade atingiu ou superou o timeout. */
    public boolean deveEncerrar(long agoraMillis) {
        return ativo() && (agoraMillis - ultimaInteracaoMillis) >= timeoutMillis;
    }

    long ultimaInteracaoMillis() {
        return ultimaInteracaoMillis;
    }
}
