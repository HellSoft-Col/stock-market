package tech.hellsoft.trading;

import lombok.Getter;
import lombok.extern.slf4j.Slf4j;
import tech.hellsoft.trading.config.ConectorConfig;
import tech.hellsoft.trading.dto.client.*;
import tech.hellsoft.trading.dto.server.*;
import tech.hellsoft.trading.enums.ConnectionState;
import tech.hellsoft.trading.enums.MessageType;
import tech.hellsoft.trading.exception.ConexionFallidaException;
import tech.hellsoft.trading.internal.connection.HeartbeatManager;
import tech.hellsoft.trading.internal.connection.WebSocketHandler;
import tech.hellsoft.trading.internal.routing.MessageRouter;
import tech.hellsoft.trading.internal.routing.MessageSequencer;
import tech.hellsoft.trading.internal.serialization.JsonSerializer;

import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.WebSocket;
import java.util.List;
import java.util.concurrent.CopyOnWriteArrayList;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.Semaphore;

@Slf4j
public class ConectorBolsa {
    private final ConectorConfig config;
    private final List<EventListener> listeners = new CopyOnWriteArrayList<>();
    private final MessageSequencer sequencer = new MessageSequencer();
    private final MessageRouter router = new MessageRouter();
    private final ExecutorService callbackExecutor = Executors.newVirtualThreadPerTaskExecutor();
    private final Semaphore sendLock = new Semaphore(1);

    @Getter
    private volatile ConnectionState state = ConnectionState.DISCONNECTED;
    private volatile WebSocket webSocket;
    private HeartbeatManager heartbeatManager;

    public ConectorBolsa(ConectorConfig config) {
        if (config == null) {
            throw new IllegalArgumentException("config cannot be null");
        }
        
        config.validate();
        this.config = config;
    }

    public ConectorBolsa() {
        this(ConectorConfig.defaultConfig());
    }

    public void addListener(EventListener listener) {
        if (listener == null) {
            throw new IllegalArgumentException("listener cannot be null");
        }
        
        listeners.add(listener);
        log.debug("Added listener: {}", listener.getClass().getSimpleName());
    }

    public void removeListener(EventListener listener) {
        if (listener == null) {
            return;
        }
        
        listeners.remove(listener);
        log.debug("Removed listener: {}", listener.getClass().getSimpleName());
    }

    public void conectar(String host, int port, String token) throws ConexionFallidaException {
        if (state != ConnectionState.DISCONNECTED) {
            throw new IllegalStateException("Already connected or connecting");
        }

        if (host == null || host.isBlank()) {
            throw new IllegalArgumentException("host cannot be null or blank");
        }

        if (port <= 0 || port > 65535) {
            throw new IllegalArgumentException("port must be between 1 and 65535");
        }

        if (token == null || token.isBlank()) {
            throw new IllegalArgumentException("token cannot be null or blank");
        }

        try {
            state = ConnectionState.CONNECTING;
            
            URI uri = URI.create(String.format("ws://%s:%d", host, port));
            log.info("Connecting to {}", uri);

            try (HttpClient client = HttpClient.newHttpClient()) {

                WebSocketHandler handler = new WebSocketHandler(
                        this::onMessageReceived,
                        this::onWebSocketError,
                        this::onWebSocketClosed
                );

                webSocket = client.newWebSocketBuilder()
                        .connectTimeout(config.getConnectionTimeout())
                        .buildAsync(uri, handler)
                        .join();
            }

            state = ConnectionState.CONNECTED;
            log.info("Connected to {}", uri);

            enviarLogin(token);
            startHeartbeat();

        } catch (Exception e) {
            state = ConnectionState.DISCONNECTED;
            throw new ConexionFallidaException("Failed to connect to " + host + ":" + port, host, port, e);
        }
    }

    public void desconectar() {
        if (state == ConnectionState.DISCONNECTED) {
            return;
        }

        log.info("Disconnecting...");
        
        stopHeartbeat();
        
        if (webSocket != null) {
            webSocket.sendClose(WebSocket.NORMAL_CLOSURE, "Client disconnect");
            webSocket = null;
        }
        
        state = ConnectionState.DISCONNECTED;
        log.info("Disconnected");
    }

    public void enviarLogin(String token) {
        if (state != ConnectionState.CONNECTED) {
            throw new IllegalStateException("Not connected");
        }

        if (token == null || token.isBlank()) {
            throw new IllegalArgumentException("token cannot be null or blank");
        }

        LoginMessage login = LoginMessage.builder()
            .type(MessageType.LOGIN)
            .token(token)
            .tz("UTC")
            .build();

        sendMessage(login);
    }

    public void enviarOrden(OrderMessage order) {
        if (state != ConnectionState.AUTHENTICATED) {
            throw new IllegalStateException("Not authenticated");
        }

        if (order == null) {
            throw new IllegalArgumentException("order cannot be null");
        }

        order.setType(MessageType.ORDER);
        sendMessage(order);
    }

    public void enviarCancelacion(String clOrdID) {
        if (state != ConnectionState.AUTHENTICATED) {
            throw new IllegalStateException("Not authenticated");
        }

        if (clOrdID == null || clOrdID.isBlank()) {
            throw new IllegalArgumentException("clOrdID cannot be null or blank");
        }

        CancelMessage cancel = CancelMessage.builder()
            .type(MessageType.CANCEL)
            .clOrdID(clOrdID)
            .build();

        sendMessage(cancel);
    }

    public void enviarActualizacionProduccion(ProductionUpdateMessage update) {
        if (state != ConnectionState.AUTHENTICATED) {
            throw new IllegalStateException("Not authenticated");
        }

        if (update == null) {
            throw new IllegalArgumentException("update cannot be null");
        }

        update.setType(MessageType.PRODUCTION_UPDATE);
        sendMessage(update);
    }

    public void enviarRespuestaOferta(AcceptOfferMessage response) {
        if (state != ConnectionState.AUTHENTICATED) {
            throw new IllegalStateException("Not authenticated");
        }

        if (response == null) {
            throw new IllegalArgumentException("response cannot be null");
        }

        response.setType(MessageType.ACCEPT_OFFER);
        sendMessage(response);
    }

    private void sendMessage(Object message) {
        if (webSocket == null || webSocket.isOutputClosed()) {
            throw new IllegalStateException("WebSocket not connected");
        }

        try {
            sendLock.acquire();
            try {
                String json = JsonSerializer.toJson(message);
                log.debug("Sending: {}", json);
                webSocket.sendText(json, true).join();
            } finally {
                sendLock.release();
            }
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            throw new RuntimeException("Send interrupted", e);
        }
    }

    private void onMessageReceived(String json) {
        log.trace("Received: {}", json);
        sequencer.submit(() -> router.routeMessage(json, createHandlers()));
    }

    private void onWebSocketError(Throwable error) {
        log.error("WebSocket error", error);
        state = ConnectionState.DISCONNECTED;
        stopHeartbeat();
        notifyConnectionLost(error);
    }

    private void onWebSocketClosed() {
        log.info("WebSocket closed");
        state = ConnectionState.DISCONNECTED;
        stopHeartbeat();
    }

    private MessageRouter.MessageHandlers createHandlers() {
        return new MessageRouter.MessageHandlers() {
            @Override
            public void onLoginOk(LoginOKMessage message) {
                state = ConnectionState.AUTHENTICATED;
                log.info("Authenticated as team: {}", message.getTeam());
                notifyListeners(l -> l.onLoginOk(message));
            }

            @Override
            public void onFill(FillMessage message) {
                notifyListeners(l -> l.onFill(message));
            }

            @Override
            public void onTicker(TickerMessage message) {
                notifyListeners(l -> l.onTicker(message));
            }

            @Override
            public void onOffer(OfferMessage message) {
                notifyListeners(l -> l.onOffer(message));
            }

            @Override
            public void onError(ErrorMessage message) {
                notifyListeners(l -> l.onError(message));
            }

            @Override
            public void onOrderAck(OrderAckMessage message) {
                notifyListeners(l -> l.onOrderAck(message));
            }

            @Override
            public void onInventoryUpdate(InventoryUpdateMessage message) {
                notifyListeners(l -> l.onInventoryUpdate(message));
            }

            @Override
            public void onBalanceUpdate(BalanceUpdateMessage message) {
                notifyListeners(l -> l.onBalanceUpdate(message));
            }

            @Override
            public void onEventDelta(EventDeltaMessage message) {
                notifyListeners(l -> l.onEventDelta(message));
            }

            @Override
            public void onBroadcast(BroadcastNotificationMessage message) {
                notifyListeners(l -> l.onBroadcast(message));
            }

            @Override
            public void onPong(PongMessage message) {
                if (heartbeatManager != null) {
                    heartbeatManager.onPongReceived();
                }
            }
        };
    }

    private void notifyListeners(java.util.function.Consumer<EventListener> action) {
        listeners.forEach(listener ->
            callbackExecutor.execute(() -> {
                try {
                    action.accept(listener);
                } catch (Exception e) {
                    log.error("Listener error", e);
                }
            })
        );
    }

    private void notifyConnectionLost(Throwable error) {
        listeners.forEach(listener ->
            callbackExecutor.execute(() -> {
                try {
                    listener.onConnectionLost(error);
                } catch (Exception e) {
                    log.error("Listener error in onConnectionLost", e);
                }
            })
        );
    }

    private void startHeartbeat() {
        if (heartbeatManager != null) {
            heartbeatManager.stop();
        }

        heartbeatManager = new HeartbeatManager(
            config.getHeartbeatInterval(),
            config.getHeartbeatInterval().multipliedBy(3),
            this::sendMessage,
            this::onPongTimeout
        );

        heartbeatManager.start();
        log.debug("Heartbeat started");
    }

    private void stopHeartbeat() {
        if (heartbeatManager != null) {
            heartbeatManager.stop();
            heartbeatManager = null;
            log.debug("Heartbeat stopped");
        }
    }

    private void onPongTimeout() {
        log.warn("Pong timeout - disconnecting");
        desconectar();
        notifyConnectionLost(new RuntimeException("Pong timeout"));
    }

    public void shutdown() {
        desconectar();
        sequencer.shutdown();
        callbackExecutor.shutdown();
        log.info("SDK shutdown complete");
    }
}
