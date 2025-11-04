package tech.hellsoft.trading.internal.connection;

import java.net.http.WebSocket;
import java.nio.ByteBuffer;
import java.util.concurrent.CompletionStage;
import java.util.function.Consumer;

import lombok.extern.slf4j.Slf4j;

@Slf4j
public class WebSocketHandler implements WebSocket.Listener {
  private final StringBuilder messageBuffer = new StringBuilder();
  private final Consumer<String> onMessage;
  private final Consumer<Throwable> onError;
  private final Runnable onClose;

  public WebSocketHandler(
      Consumer<String> onMessage, Consumer<Throwable> onError, Runnable onClose) {
    if (onMessage == null) {
      throw new IllegalArgumentException("onMessage cannot be null");
    }

    if (onError == null) {
      throw new IllegalArgumentException("onError cannot be null");
    }

    if (onClose == null) {
      throw new IllegalArgumentException("onClose cannot be null");
    }

    this.onMessage = onMessage;
    this.onError = onError;
    this.onClose = onClose;
  }

  @Override
  public void onOpen(WebSocket webSocket) {
    log.debug("WebSocket opened");
    webSocket.request(1);
  }

  @Override
  public CompletionStage<?> onText(WebSocket webSocket, CharSequence data, boolean last) {
    messageBuffer.append(data);

    if (last) {
      String json = messageBuffer.toString();
      messageBuffer.setLength(0);

      try {
        onMessage.accept(json);
      } catch (Exception e) {
        log.error("Error processing message: {}", json, e);
      }
    }

    webSocket.request(1);
    return null;
  }

  @Override
  public CompletionStage<?> onBinary(WebSocket webSocket, ByteBuffer data, boolean last) {
    log.warn("Received unexpected binary data");
    webSocket.request(1);
    return null;
  }

  @Override
  public CompletionStage<?> onPing(WebSocket webSocket, ByteBuffer message) {
    log.trace("Received ping");
    webSocket.request(1);
    return null;
  }

  @Override
  public CompletionStage<?> onPong(WebSocket webSocket, ByteBuffer message) {
    log.trace("Received pong");
    webSocket.request(1);
    return null;
  }

  @Override
  public CompletionStage<?> onClose(WebSocket webSocket, int statusCode, String reason) {
    log.info("WebSocket closed: {} - {}", statusCode, reason);

    try {
      onClose.run();
    } catch (Exception e) {
      log.error("Error in close handler", e);
    }

    return null;
  }

  @Override
  public void onError(WebSocket webSocket, Throwable error) {
    log.error("WebSocket error", error);

    try {
      onError.accept(error);
    } catch (Exception e) {
      log.error("Error in error handler", e);
    }
  }
}
