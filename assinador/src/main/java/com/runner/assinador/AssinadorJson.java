package com.runner.assinador;

import com.fasterxml.jackson.databind.ObjectMapper;

final class AssinadorJson {

    private static final ObjectMapper MAPPER = new ObjectMapper();

    private AssinadorJson() {
    }

    static ObjectMapper mapper() {
        return MAPPER;
    }
}
