package com.example.gatewayservice.config;

import com.example.common_api.util.AesUtil;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.reactivestreams.Publisher;
import org.springframework.cloud.gateway.filter.GlobalFilter;
import org.springframework.core.Ordered;
import org.springframework.core.io.buffer.DataBuffer;
import org.springframework.core.io.buffer.DataBufferFactory;
import org.springframework.core.io.buffer.DataBufferUtils;
import org.springframework.http.HttpHeaders;
import org.springframework.http.MediaType;
import org.springframework.http.server.reactive.ServerHttpRequestDecorator;
import org.springframework.http.server.reactive.ServerHttpResponseDecorator;
import org.springframework.stereotype.Component;
import org.springframework.web.server.ServerWebExchange;
import reactor.core.publisher.Flux;
import reactor.core.publisher.Mono;

import java.nio.charset.StandardCharsets;

@Component
public class AesCryptoFilter implements GlobalFilter, Ordered {

    private static final String FIELD = "cipherText";
    private static final ObjectMapper MAPPER = new ObjectMapper();

    @Override
    public Mono<Void> filter(ServerWebExchange exchange,
                             org.springframework.cloud.gateway.filter.GatewayFilterChain chain) {

        String path = exchange.getRequest().getURI().getPath();

        // ✅ 如果是上传图片或静态资源，则直接跳过加密解密逻辑
        if (path.startsWith("/uploadFile/")) {
            return chain.filter(exchange);
        }

        return DataBufferUtils.join(exchange.getRequest().getBody())
                .flatMap(buf -> {
                    byte[] originalBytes = new byte[buf.readableByteCount()];
                    buf.read(originalBytes);
                    DataBufferUtils.release(buf);

                    boolean needCrypto = false;
                    String cipherHex = null;

                    if ("GET".equalsIgnoreCase(exchange.getRequest().getMethodValue())) {
                        cipherHex = exchange.getRequest().getQueryParams().getFirst(FIELD);
                        needCrypto = cipherHex != null;
                    } else {
                        String bodyStr = new String(originalBytes, StandardCharsets.UTF_8);
                        if (bodyStr.contains(FIELD)) {
                            try {
                                JsonNode node = MAPPER.readTree(bodyStr);
                                cipherHex = node.get(FIELD).asText();
                                needCrypto = true;
                            } catch (Exception ignore) {}
                        }
                    }

                    if (needCrypto) {
                        String plain;
                        try {
                            plain = AesUtil.decrypt(cipherHex);
                        } catch (Exception e) {
                            return Mono.error(new IllegalStateException("AES decrypt failed", e));
                        }
                        byte[] plainBytes = plain.getBytes(StandardCharsets.UTF_8);

                        ServerHttpRequestDecorator newReq = wrapRequest(exchange, plainBytes);
                        ServerHttpResponseDecorator newResp = wrapResponseForEncrypt(exchange);

                        return chain.filter(exchange.mutate()
                                .request(newReq)
                                .response(newResp)
                                .build());
                    }

                    ServerHttpRequestDecorator newReq = wrapRequest(exchange, originalBytes);
                    return chain.filter(exchange.mutate().request(newReq).build());
                });
    }

    private ServerHttpRequestDecorator wrapRequest(ServerWebExchange exchange, byte[] bytes) {
        DataBufferFactory factory = exchange.getResponse().bufferFactory();
        return new ServerHttpRequestDecorator(exchange.getRequest()) {
            @Override
            public Flux<DataBuffer> getBody() {
                return Flux.just(factory.wrap(bytes));
            }

            @Override
            public HttpHeaders getHeaders() {
                HttpHeaders h = new HttpHeaders();
                h.putAll(super.getHeaders());
                h.setContentType(MediaType.APPLICATION_JSON);
                h.setContentLength(bytes.length);
                return h;
            }
        };
    }

    private ServerHttpResponseDecorator wrapResponseForEncrypt(ServerWebExchange exchange) {
        return new ServerHttpResponseDecorator(exchange.getResponse()) {
            @Override
            public Mono<Void> writeWith(Publisher<? extends DataBuffer> body) {
                if (body instanceof Flux) {
                    return Flux.from(body).collectList().flatMap(list -> {
                        int totalSize = list.stream().mapToInt(DataBuffer::readableByteCount).sum();
                        byte[] fullBytes = new byte[totalSize];
                        int offset = 0;
                        for (DataBuffer db : list) {
                            int count = db.readableByteCount();
                            db.read(fullBytes, offset, count);
                            offset += count;
                            DataBufferUtils.release(db);
                        }
                        String fullBody = new String(fullBytes, StandardCharsets.UTF_8);
                        String cipher = AesUtil.encrypt(fullBody);
                        DataBuffer encryptedBuf = bufferFactory().wrap(cipher.getBytes(StandardCharsets.UTF_8));
                        getHeaders().setContentType(MediaType.TEXT_PLAIN);
                        return super.writeWith(Mono.just(encryptedBuf));
                    });
                }
                return super.writeWith(body);
            }
        };
    }

    @Override
    public int getOrder() {
        return Ordered.HIGHEST_PRECEDENCE;
    }
}
