package com.example.axiomnizam.conductor;

import com.fasterxml.jackson.databind.JsonNode;
import java.time.Instant;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.HttpHeaders;
import org.springframework.http.MediaType;
import org.springframework.stereotype.Service;
import org.springframework.web.reactive.function.client.WebClient;
import org.springframework.web.reactive.function.client.WebClientResponseException;

@Service
public class AxiomConductorClient {

    private final WebClient webClient;
    private final String baseUrl;

    private final String username;
    private final String password;

    private String accessToken;
    private String refreshToken;
    private String tokenType = "Bearer";
    private Instant expiresAt = Instant.EPOCH;

    public AxiomConductorClient(
            @Value("${axiom.base-url:http://localhost:8000}") String baseUrl,
            @Value("${axiom.username:}") String username,
            @Value("${axiom.password:}") String password
    ) {
        this.baseUrl = baseUrl.replaceAll("/$", "");
        this.webClient = WebClient.builder()
                .baseUrl(this.baseUrl)
                .defaultHeader(HttpHeaders.CONTENT_TYPE, MediaType.APPLICATION_JSON_VALUE)
                .build();
        this.username = username;
        this.password = password;
    }

    public synchronized JsonNode login() {
        if (username == null || username.isBlank() || password == null || password.isBlank()) {
            throw new IllegalStateException("Set axiom.username and axiom.password");
        }

        Map<String, Object> body = new HashMap<>();
        body.put("username", username);
        body.put("password", password);

        JsonNode json = postWithoutAuth("/auth/login", body);
        updateTokenState(json);
        return json;
    }

    public synchronized JsonNode refresh() {
        if (refreshToken == null || refreshToken.isBlank()) {
            return login();
        }

        Map<String, Object> body = new HashMap<>();
        body.put("refresh_token", refreshToken);

        JsonNode json = postWithoutAuth("/auth/refresh", body);
        updateTokenState(json);
        return json;
    }

    public JsonNode getStats() {
        ensureToken();
        return getWithAuth("/api/v1/conductor/stats");
    }

    public JsonNode listProducers() {
        ensureToken();
        return getWithAuth("/api/v1/conductor/producers");
    }

    public JsonNode createProducer(Map<String, Object> payload) {
        ensureToken();
        return postWithAuth("/api/v1/conductor/producers", payload);
    }

    public JsonNode publish(Map<String, Object> payload) {
        ensureToken();
        return postWithAuth("/api/v1/conductor/publish", payload);
    }

    public JsonNode connectRabbitMQ(String url) {
        ensureToken();
        Map<String, Object> payload = new HashMap<>();
        payload.put("type", "rabbitmq");
        payload.put("url", url);
        return postWithAuth("/api/v1/conductor/connections", payload);
    }

    public JsonNode connectKafka(List<String> brokers) {
        ensureToken();
        Map<String, Object> payload = new HashMap<>();
        payload.put("type", "kafka");
        payload.put("brokers", brokers);
        return postWithAuth("/api/v1/conductor/connections", payload);
    }

    public String getStreamWebSocketUrl() {
        ensureToken();
        String wsBase;
        if (baseUrl.startsWith("https://")) {
            wsBase = "wss://" + baseUrl.substring("https://".length());
        } else if (baseUrl.startsWith("http://")) {
            wsBase = "ws://" + baseUrl.substring("http://".length());
        } else {
            wsBase = "ws://" + baseUrl;
        }
        return wsBase + "/ws/conductor?token=" + accessToken;
    }

    private synchronized void ensureToken() {
        if (accessToken == null || accessToken.isBlank()) {
            login();
            return;
        }
        if (Instant.now().isAfter(expiresAt)) {
            refresh();
        }
    }

    private void updateTokenState(JsonNode json) {
        accessToken = json.path("access_token").asText("");
        refreshToken = json.path("refresh_token").asText("");
        String maybeType = json.path("token_type").asText("");
        tokenType = maybeType.isBlank() ? "Bearer" : maybeType;

        int expiresIn = json.path("expires_in").asInt(0);
        int safeExpires = Math.max(expiresIn - 20, 20);
        expiresAt = Instant.now().plusSeconds(safeExpires);
    }

    private JsonNode getWithAuth(String path) {
        try {
            return webClient.get()
                    .uri(path)
                    .header(HttpHeaders.AUTHORIZATION, tokenType + " " + accessToken)
                    .retrieve()
                    .bodyToMono(JsonNode.class)
                    .block();
        } catch (WebClientResponseException.Unauthorized unauthorized) {
            int status = unauthorized.getStatusCode().value();
            if (status == 401) {
                refresh();
            }
            return webClient.get()
                    .uri(path)
                    .header(HttpHeaders.AUTHORIZATION, tokenType + " " + accessToken)
                    .retrieve()
                    .bodyToMono(JsonNode.class)
                    .block();
        }
    }

    private JsonNode postWithAuth(String path, Object payload) {
        try {
            return webClient.post()
                    .uri(path)
                    .header(HttpHeaders.AUTHORIZATION, tokenType + " " + accessToken)
                    .bodyValue(payload)
                    .retrieve()
                    .bodyToMono(JsonNode.class)
                    .block();
        } catch (WebClientResponseException.Unauthorized unauthorized) {
            int status = unauthorized.getStatusCode().value();
            if (status == 401) {
                refresh();
            }
            return webClient.post()
                    .uri(path)
                    .header(HttpHeaders.AUTHORIZATION, tokenType + " " + accessToken)
                    .bodyValue(payload)
                    .retrieve()
                    .bodyToMono(JsonNode.class)
                    .block();
        }
    }

    private JsonNode postWithoutAuth(String path, Object payload) {
        return webClient.post()
                .uri(path)
                .bodyValue(payload)
                .retrieve()
                .bodyToMono(JsonNode.class)
                .block();
    }
}
