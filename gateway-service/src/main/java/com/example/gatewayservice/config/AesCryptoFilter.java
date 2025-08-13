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

    /** 判断是否为 multipart/form-data 上传请求 */
    private boolean isMultipart(ServerWebExchange exchange) {
        MediaType ct = exchange.getRequest().getHeaders().getContentType();
        return ct != null && "multipart".equalsIgnoreCase(ct.getType());
    }

    @Override
    public Mono<Void> filter(ServerWebExchange exchange,
                             org.springframework.cloud.gateway.filter.GatewayFilterChain chain) {

        String method = exchange.getRequest().getMethodValue();
        String path = exchange.getRequest().getURI().getPath();

        // 跳过 GET 请求、上传文件请求
        if ("GET".equalsIgnoreCase(method) || path.startsWith("/uploadFile/") || isMultipart(exchange)) {
            return chain.filter(exchange);
        }

        // 非 GET 请求，进入解密处理
        return DataBufferUtils.join(exchange.getRequest().getBody())
                .flatMap(buf -> {
                    byte[] originalBytes = new byte[buf.readableByteCount()];
                    buf.read(originalBytes);
                    DataBufferUtils.release(buf);

                    boolean needCrypto = false;
                    String cipherHex = null;

                    String bodyStr = new String(originalBytes, StandardCharsets.UTF_8);
                    if (bodyStr.contains(FIELD)) {
                        try {
                            JsonNode node = MAPPER.readTree(bodyStr);
                            cipherHex = node.get(FIELD).asText();
                            needCrypto = true;
                        } catch (Exception ignore) {}
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

                    // 不需要解密，仅重新包装 body
                    ServerHttpRequestDecorator newReq = wrapRequest(exchange, originalBytes);
                    return chain.filter(exchange.mutate().request(newReq).build());
                });
    }

    /** 用新的 body 包装请求，同时尽量保留原始 Content-Type */
    private ServerHttpRequestDecorator wrapRequest(ServerWebExchange exchange, byte[] bytes) {
        DataBufferFactory factory = exchange.getResponse().bufferFactory();
        MediaType originalCt = exchange.getRequest().getHeaders().getContentType();

        return new ServerHttpRequestDecorator(exchange.getRequest()) {
            @Override
            public Flux<DataBuffer> getBody() {
                return Flux.just(factory.wrap(bytes));
            }

            @Override
            public HttpHeaders getHeaders() {
                HttpHeaders h = new HttpHeaders();
                h.putAll(super.getHeaders());

                // 仅当原本是 JSON 或者没有 Content-Type 时才设置为 JSON
                if (originalCt == null || MediaType.APPLICATION_JSON.isCompatibleWith(originalCt)) {
                    h.setContentType(MediaType.APPLICATION_JSON);
                }
                h.setContentLength(bytes.length);
                return h;
            }
        };
    }

    /** 将响应内容 AES 加密后返回 */
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
