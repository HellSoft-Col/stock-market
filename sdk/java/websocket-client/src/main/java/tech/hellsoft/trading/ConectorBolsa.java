package tech.hellsoft.trading;

import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.WebSocket;
import java.util.List;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.CopyOnWriteArrayList;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.Semaphore;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.TimeoutException;

import tech.hellsoft.trading.config.ConectorConfig;
import tech.hellsoft.trading.dto.client.AcceptOfferMessage;
import tech.hellsoft.trading.dto.client.CancelMessage;
import tech.hellsoft.trading.dto.client.LoginMessage;
import tech.hellsoft.trading.dto.client.OrderMessage;
import tech.hellsoft.trading.dto.client.ProductionUpdateMessage;
import tech.hellsoft.trading.dto.server.BalanceUpdateMessage;
import tech.hellsoft.trading.dto.server.BroadcastNotificationMessage;
import tech.hellsoft.trading.dto.server.ErrorMessage;
import tech.hellsoft.trading.dto.server.EventDeltaMessage;
import tech.hellsoft.trading.dto.server.FillMessage;
import tech.hellsoft.trading.dto.server.InventoryUpdateMessage;
import tech.hellsoft.trading.dto.server.LoginOKMessage;
import tech.hellsoft.trading.dto.server.OfferMessage;
import tech.hellsoft.trading.dto.server.OrderAckMessage;
import tech.hellsoft.trading.dto.server.PongMessage;
import tech.hellsoft.trading.dto.server.TickerMessage;
import tech.hellsoft.trading.enums.ConnectionState;
import tech.hellsoft.trading.enums.ErrorCode;
import tech.hellsoft.trading.enums.MessageType;
import tech.hellsoft.trading.exception.ConexionFallidaException;
import tech.hellsoft.trading.internal.connection.HeartbeatManager;
import tech.hellsoft.trading.internal.connection.WebSocketHandler;
import tech.hellsoft.trading.internal.routing.MessageRouter;
import tech.hellsoft.trading.internal.routing.MessageSequencer;
import tech.hellsoft.trading.internal.serialization.JsonSerializer;
import tech.hellsoft.trading.tasks.TareaAutomatica;
import tech.hellsoft.trading.tasks.TareaAutomaticaManager;

import lombok.Getter;
import lombok.extern.slf4j.Slf4j;

/**
 * Main client SDK for connecting to the Stock Market WebSocket server.
 *
 * <p>This class provides the primary interface for establishing connections, sending messages, and
 * receiving market data. All operations are thread-safe and use virtual threads for concurrent
 * processing.
 *
 * <p>Typical usage pattern:
 *
 * <pre>{@code
 * ConectorBolsa connector = new ConectorBolsa();
 * connector.addListener(new MyEventListener());
 * connector.conectar("wss://market.example.com/ws", "your-token");
 *
 * // Send orders - SDK automatically waits for authentication
 * OrderMessage order = OrderMessage.builder()
 *     .clOrdID("order-123")
 *     .side(OrderSide.BUY)
 *     .product(Product.GUACA)
 *     .qty(100)
 *     .mode(OrderMode.MARKET)
 *     .build();
 * connector.enviarOrden(order);
 *
 * // Send production updates - also waits automatically
 * ProductionUpdateMessage update = ProductionUpdateMessage.builder()
 *     .product(Product.GUACA)
 *     .quantity(5)
 *     .build();
 * connector.enviarActualizacionProduccion(update);
 *
 * // Clean shutdown
 * connector.shutdown();
 * }</pre>
 *
 * @see EventListener for receiving server messages
 * @see ConectorConfig for configuration options
 */
@Slf4j
public class ConectorBolsa {
  private final ConectorConfig config;
  private final List<EventListener> listeners = new CopyOnWriteArrayList<>();
  private final MessageSequencer sequencer = new MessageSequencer();
  private final MessageRouter router = new MessageRouter();
  private final ExecutorService callbackExecutor = Executors.newVirtualThreadPerTaskExecutor();
  private final Semaphore sendLock = new Semaphore(1);
  private final TareaAutomaticaManager tareaManager = new TareaAutomaticaManager();

  @Getter private volatile ConnectionState state = ConnectionState.DISCONNECTED;
  private volatile WebSocket webSocket;
  private HeartbeatManager heartbeatManager;
  private volatile CompletableFuture<LoginOKMessage> loginFuture;

  /**
   * Creates a new ConectorBolsa instance with custom configuration.
   *
   * @param config the configuration settings (must not be null)
   * @throws IllegalArgumentException if config is null
   * @throws IllegalArgumentException if configuration validation fails
   * @see ConectorConfig for available configuration options
   */
  public ConectorBolsa(ConectorConfig config) {
    if (config == null) {
      throw new IllegalArgumentException("config cannot be null");
    }

    config.validate();
    this.config = config;
  }

  /**
   * Creates a new ConectorBolsa instance with default configuration.
   *
   * <p>Uses {@link ConectorConfig#defaultConfig()} for all settings. This is suitable for most use
   * cases.
   */
  public ConectorBolsa() {
    this(ConectorConfig.defaultConfig());
  }

  /**
   * Adds an event listener to receive server messages.
   *
   * <p>Multiple listeners can be added. All callbacks are executed on virtual threads to avoid
   * blocking the WebSocket processing thread.
   *
   * @param listener the listener to add (must not be null)
   * @throws IllegalArgumentException if listener is null
   * @see EventListener for callback methods
   */
  public void addListener(EventListener listener) {
    if (listener == null) {
      throw new IllegalArgumentException("listener cannot be null");
    }

    listeners.add(listener);
    log.debug("Added listener: {}", listener.getClass().getSimpleName());
  }

  /**
   * Removes an event listener.
   *
   * <p>If the listener is not currently registered, this method does nothing. The method is
   * thread-safe and can be called from any thread.
   *
   * @param listener the listener to remove (may be null)
   */
  public void removeListener(EventListener listener) {
    if (listener == null) {
      return;
    }

    listeners.remove(listener);
    log.debug("Removed listener: {}", listener.getClass().getSimpleName());
  }

  /**
   * Registers an automatic task that will start execution immediately.
   *
   * <p>Tasks are automatically managed and will stop when the connection is closed via {@link
   * #shutdown()} or {@link #desconectar()}. Multiple tasks can be registered and they will execute
   * concurrently using virtual threads.
   *
   * <p>Example usage:
   *
   * <pre>{@code
   * // Create a price monitoring task
   * public class MonitorPreciosTarea extends TareaAutomatica {
   *     private final ConectorBolsa connector;
   *     private final String producto;
   *
   *     public MonitorPreciosTarea(ConectorBolsa connector, String producto) {
   *         super("monitor-" + producto);
   *         this.connector = connector;
   *         this.producto = producto;
   *     }
   *
   *     @Override
   *     protected void ejecutar() {
   *         // Check prices and send orders
   *         log.info("Monitoring prices for {}", producto);
   *     }
   *
   *     @Override
   *     protected Duration intervalo() {
   *         return Duration.ofMillis(100); // Fast execution
   *     }
   * }
   *
   * // Register the task
   * ConectorBolsa connector = new ConectorBolsa();
   * connector.conectar("wss://market.example.com", "token");
   * connector.registrarTarea(new MonitorPreciosTarea(connector, "GUACA"));
   * }</pre>
   *
   * <p>This method is thread-safe and can be called from any thread.
   *
   * @param tarea the task to register (must not be null)
   * @throws IllegalArgumentException if tarea is null
   * @see TareaAutomatica for creating custom tasks
   * @see #detenerTarea(String) to stop a specific task
   */
  public void registrarTarea(TareaAutomatica tarea) {
    if (tarea == null) {
      throw new IllegalArgumentException("tarea cannot be null");
    }
    tareaManager.registrar(tarea);
    log.debug("Registered automatic task: {}", tarea.getTaskKey());
  }

  /**
   * Stops and removes a registered automatic task.
   *
   * <p>If the task is not currently registered, this method does nothing. The task will stop
   * gracefully after completing its current execution cycle.
   *
   * <p>This method is thread-safe and can be called from any thread.
   *
   * @param taskKey the key of the task to stop (may be null)
   * @see #registrarTarea(TareaAutomatica) to register tasks
   */
  public void detenerTarea(String taskKey) {
    if (taskKey == null) {
      return;
    }
    tareaManager.detener(taskKey);
    log.debug("Stopped automatic task: {}", taskKey);
  }

  /**
   * Connects to the Stock Market WebSocket server using a full WebSocket URL.
   *
   * <p>This method supports both WS (ws://) and WSS (wss://) protocols. Use WSS for secure
   * connections (equivalent to HTTPS). This is the recommended method for production environments.
   *
   * <p><strong>Note:</strong> Authentication happens asynchronously after connection. However, all
   * send methods ({@link #enviarOrden(OrderMessage)}, {@link
   * #enviarActualizacionProduccion(ProductionUpdateMessage)}, etc.) automatically wait for
   * authentication to complete before sending, so you can call them immediately after this method
   * without any additional waiting.
   *
   * <p>Example usage:
   *
   * <pre>{@code
   * connector.conectar("wss://trading.hellsoft.tech/ws", "your-token");
   *
   * // Send methods automatically wait for authentication
   * connector.enviarOrden(order);
   * connector.enviarActualizacionProduccion(update);
   * }</pre>
   *
   * @param websocketUrl the full WebSocket URL (must not be null or blank)
   * @param token the authentication token (must not be null or blank)
   * @throws ConexionFallidaException if the connection fails
   * @throws IllegalStateException if already connected or connecting
   * @throws IllegalArgumentException if any parameter is invalid
   * @see #desconectar() to disconnect
   * @see EventListener#onLoginOk(LoginOKMessage) for successful authentication
   */
  public void conectar(String websocketUrl, String token) throws ConexionFallidaException {
    if (state != ConnectionState.DISCONNECTED) {
      throw new IllegalStateException("Already connected or connecting");
    }

    if (websocketUrl == null || websocketUrl.isBlank()) {
      throw new IllegalArgumentException("websocketUrl cannot be null or blank");
    }

    if (token == null || token.isBlank()) {
      throw new IllegalArgumentException("token cannot be null or blank");
    }

    try {
      state = ConnectionState.CONNECTING;
      loginFuture = new CompletableFuture<>();

      URI uri = URI.create(websocketUrl);

      // Validate protocol
      String scheme = uri.getScheme();
      if (scheme == null || (!scheme.equals("ws") && !scheme.equals("wss"))) {
        throw new IllegalArgumentException("URL must start with ws:// or wss://");
      }

      log.info("Connecting to {} (secure: {})", uri, scheme.equals("wss"));

      HttpClient client = HttpClient.newHttpClient();

      WebSocketHandler handler =
          new WebSocketHandler(
              this::onMessageReceived, this::onWebSocketError, this::onWebSocketClosed);

      webSocket =
          client
              .newWebSocketBuilder()
              .connectTimeout(config.getConnectionTimeout())
              .buildAsync(uri, handler)
              .join();

      state = ConnectionState.CONNECTED;
      log.info("Connected to {}", uri);

      enviarLogin(token);
      startHeartbeat();

    } catch (Exception e) {
      state = ConnectionState.DISCONNECTED;
      loginFuture = null;
      throw new ConexionFallidaException(
          "Failed to connect to " + websocketUrl, websocketUrl, 0, e);
    }
  }

  /**
   * Connects to the Stock Market WebSocket server using host and port (non-secure).
   *
   * <p>This method establishes a non-secure WebSocket connection (ws://). For production
   * environments, use {@link #conectar(String, String)} or {@link #conectarSeguro(String, int,
   * String)} for secure connections.
   *
   * <p>After successful connection and authentication, the {@link
   * EventListener#onLoginOk(LoginOKMessage)} callback will be invoked.
   *
   * @param host the server hostname or IP address (must not be null or blank)
   * @param port the server port (must be between 1 and 65535)
   * @param token the authentication token (must not be null or blank)
   * @throws ConexionFallidaException if the connection fails
   * @throws IllegalStateException if already connected or connecting
   * @throws IllegalArgumentException if any parameter is invalid
   * @see #conectarSeguro(String, int, String) for secure connections
   * @see #desconectar() to disconnect
   * @see EventListener#onLoginOk(LoginOKMessage) for successful authentication
   */
  public void conectar(String host, int port, String token) throws ConexionFallidaException {
    conectarInternal(host, port, token, false);
  }

  /**
   * Connects to the Stock Market WebSocket server using secure WebSocket (WSS).
   *
   * <p>This method establishes a secure WebSocket connection (wss://), which is the WebSocket
   * equivalent of HTTPS. This is the recommended method for production environments.
   *
   * <p>Example usage:
   *
   * <pre>{@code
   * connector.conectarSeguro("trading.hellsoft.tech", 443, "your-token");
   * }</pre>
   *
   * @param host the server hostname or IP address (must not be null or blank)
   * @param port the server port (must be between 1 and 65535)
   * @param token the authentication token (must not be null or blank)
   * @throws ConexionFallidaException if the connection fails
   * @throws IllegalStateException if already connected or connecting
   * @throws IllegalArgumentException if any parameter is invalid
   * @see #desconectar() to disconnect
   * @see EventListener#onLoginOk(LoginOKMessage) for successful authentication
   */
  public void conectarSeguro(String host, int port, String token) throws ConexionFallidaException {
    conectarInternal(host, port, token, true);
  }

  private void conectarInternal(String host, int port, String token, boolean secure)
      throws ConexionFallidaException {
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
      loginFuture = new CompletableFuture<>();

      String protocol = secure ? "wss" : "ws";
      URI uri = URI.create(String.format("%s://%s:%d", protocol, host, port));
      log.info("Connecting to {} (secure: {})", uri, secure);

      HttpClient client = HttpClient.newHttpClient();

      WebSocketHandler handler =
          new WebSocketHandler(
              this::onMessageReceived, this::onWebSocketError, this::onWebSocketClosed);

      webSocket =
          client
              .newWebSocketBuilder()
              .connectTimeout(config.getConnectionTimeout())
              .buildAsync(uri, handler)
              .join();

      state = ConnectionState.CONNECTED;
      log.info("Connected to {}", uri);

      enviarLogin(token);
      startHeartbeat();

    } catch (Exception e) {
      state = ConnectionState.DISCONNECTED;
      loginFuture = null;
      throw new ConexionFallidaException(
          "Failed to connect to " + host + ":" + port, host, port, e);
    }
  }

  /**
   * Disconnects from the server gracefully.
   *
   * <p>This method sends a WebSocket close frame and stops the heartbeat mechanism. If already
   * disconnected, this method does nothing. The connection state is set to DISCONNECTED.
   *
   * <p>After disconnection, the {@link EventListener#onConnectionLost(Throwable)} callback will NOT
   * be invoked since this is an intentional disconnection.
   *
   * @see #conectar(String, int, String) to reconnect
   * @see #shutdown() for complete cleanup
   */
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
    loginFuture = null;
    log.info("Disconnected");
  }

  /**
   * Waits internally for authentication to complete.
   *
   * <p>This is called automatically by send methods to ensure authentication has completed before
   * sending messages.
   */
  private void waitForAuthentication() {
    if (state == ConnectionState.AUTHENTICATED) {
      return;
    }

    if (loginFuture == null) {
      throw new IllegalStateException("Not connected");
    }

    try {
      loginFuture.get(config.getConnectionTimeout().toMillis(), TimeUnit.MILLISECONDS);
    } catch (TimeoutException e) {
      throw new IllegalStateException("Authentication timed out", e);
    } catch (Exception e) {
      throw new IllegalStateException("Authentication failed", e);
    }
  }

  /**
   * Sends a login message to authenticate with the server.
   *
   * <p>This method is typically called automatically by {@link #conectar(String, int, String)}, but
   * can be called manually for re-authentication scenarios.
   *
   * <p>Upon successful authentication, the {@link EventListener#onLoginOk(LoginOKMessage)} callback
   * will be invoked with the session details.
   *
   * @param token the authentication token (must not be null or blank)
   * @throws IllegalStateException if not connected to the server
   * @throws IllegalArgumentException if token is null or blank
   * @see EventListener#onLoginOk(LoginOKMessage) for successful authentication
   * @see EventListener#onError(ErrorMessage) for authentication failures
   */
  public void enviarLogin(String token) {
    if (state != ConnectionState.CONNECTED) {
      throw new IllegalStateException("Not connected");
    }

    if (token == null || token.isBlank()) {
      throw new IllegalArgumentException("token cannot be null or blank");
    }

    LoginMessage login =
        LoginMessage.builder().type(MessageType.LOGIN).token(token).tz("UTC").build();

    sendMessage(login);
  }

  /**
   * Sends a buy or sell order to the market.
   *
   * <p>The order must be properly constructed with all required fields. Order responses arrive
   * asynchronously via {@link EventListener#onFill(FillMessage)} for executions and {@link
   * EventListener#onOrderAck(OrderAckMessage)} for acknowledgments.
   *
   * <p><strong>Note:</strong> This method automatically waits for authentication to complete if
   * called immediately after {@link #conectar(String, String)}. The SDK handles the timing
   * internally, so you can call this method right after connecting.
   *
   * <p>Example usage:
   *
   * <pre>{@code
   * connector.conectar("wss://server/ws", "token");
   *
   * // SDK automatically waits for authentication to complete
   * OrderMessage order = OrderMessage.builder()
   *     .clOrdID("unique-order-id")
   *     .side(OrderSide.BUY)
   *     .product(Product.GUACA)
   *     .qty(100)
   *     .mode(OrderMode.MARKET)
   *     .limitPrice(50.0) // Only for LIMIT orders
   *     .build();
   * connector.enviarOrden(order);
   * }</pre>
   *
   * @param order the order to send (must not be null)
   * @throws IllegalStateException if not connected or authentication fails
   * @throws IllegalArgumentException if order is null
   * @see EventListener#onFill(FillMessage) for order executions
   * @see EventListener#onOrderAck(OrderAckMessage) for order acknowledgments
   * @see EventListener#onError(ErrorMessage) for order rejections
   */
  public void enviarOrden(OrderMessage order) {
    if (order == null) {
      throw new IllegalArgumentException("order cannot be null");
    }

    waitForAuthentication();

    order.setType(MessageType.ORDER);
    sendMessage(order);
  }

  /**
   * Cancels a previously submitted order.
   *
   * <p>Cancels the order with the specified client order ID (clOrdID). The cancellation result
   * arrives asynchronously via {@link EventListener#onOrderAck(OrderAckMessage)}.
   *
   * <p><strong>Note:</strong> This method automatically waits for authentication to complete.
   *
   * @param clOrdID the client order ID of the order to cancel (must not be null or blank)
   * @throws IllegalStateException if not connected or authentication fails
   * @throws IllegalArgumentException if clOrdID is null or blank
   * @see EventListener#onOrderAck(OrderAckMessage) for cancellation confirmation
   * @see EventListener#onError(ErrorMessage) for cancellation failures
   */
  public void enviarCancelacion(String clOrdID) {
    if (clOrdID == null || clOrdID.isBlank()) {
      throw new IllegalArgumentException("clOrdID cannot be null or blank");
    }

    waitForAuthentication();

    CancelMessage cancel =
        CancelMessage.builder().type(MessageType.CANCEL).clOrdID(clOrdID).build();

    sendMessage(cancel);
  }

  /**
   * Sends a production update to the server.
   *
   * <p>Production updates are used to report manufacturing or production activities that may affect
   * market conditions and inventory levels.
   *
   * <p><strong>Note:</strong> This method automatically waits for authentication to complete.
   *
   * @param update the production update message (must not be null)
   * @throws IllegalStateException if not connected or authentication fails
   * @throws IllegalArgumentException if update is null
   * @see EventListener#onInventoryUpdate(InventoryUpdateMessage) for inventory changes
   * @see EventListener#onError(ErrorMessage) for update failures
   */
  public void enviarActualizacionProduccion(ProductionUpdateMessage update) {
    if (update == null) {
      throw new IllegalArgumentException("update cannot be null");
    }

    waitForAuthentication();

    update.setType(MessageType.PRODUCTION_UPDATE);
    sendMessage(update);
  }

  /**
   * Sends a response to a received offer.
   *
   * <p>Used to accept or reject offers received via {@link EventListener#onOffer(OfferMessage)}.
   * The response should reference the original offer ID.
   *
   * <p><strong>Note:</strong> This method automatically waits for authentication to complete.
   *
   * @param response the offer response message (must not be null)
   * @throws IllegalStateException if not connected or authentication fails
   * @throws IllegalArgumentException if response is null
   * @see EventListener#onOffer(OfferMessage) for receiving offers
   * @see EventListener#onFill(FillMessage) for offer executions
   * @see EventListener#onError(ErrorMessage) for response failures
   */
  public void enviarRespuestaOferta(AcceptOfferMessage response) {
    if (response == null) {
      throw new IllegalArgumentException("response cannot be null");
    }

    waitForAuthentication();

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

        if (loginFuture != null) {
          loginFuture.complete(message);
        }

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
        if (message.getCode() == ErrorCode.AUTH_FAILED
            && loginFuture != null
            && !loginFuture.isDone()) {
          loginFuture.completeExceptionally(
              new RuntimeException("Authentication failed: " + message.getReason()));
        }

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
    listeners.forEach(
        listener ->
            callbackExecutor.execute(
                () -> {
                  try {
                    action.accept(listener);
                  } catch (Exception e) {
                    log.error("Listener error", e);
                  }
                }));
  }

  private void notifyConnectionLost(Throwable error) {
    listeners.forEach(
        listener ->
            callbackExecutor.execute(
                () -> {
                  try {
                    listener.onConnectionLost(error);
                  } catch (Exception e) {
                    log.error("Listener error in onConnectionLost", e);
                  }
                }));
  }

  private void startHeartbeat() {
    if (heartbeatManager != null) {
      heartbeatManager.stop();
    }

    heartbeatManager =
        new HeartbeatManager(
            config.getHeartbeatInterval(),
            config.getHeartbeatInterval().multipliedBy(3),
            this::sendMessage,
            this::onPongTimeout);

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

  /**
   * Performs a complete shutdown of the SDK.
   *
   * <p>This method disconnects from the server, stops all registered automatic tasks, and releases
   * all resources including thread pools and message sequencers. After calling this method, the
   * ConectorBolsa instance cannot be reused.
   *
   * <p>This method should be called when the application is shutting down or when the SDK instance
   * is no longer needed.
   *
   * @see #desconectar() for graceful disconnect without resource cleanup
   * @see #detenerTarea(String) to stop individual tasks
   */
  public void shutdown() {
    desconectar();
    tareaManager.shutdown();
    sequencer.shutdown();
    callbackExecutor.shutdown();
    log.info("SDK shutdown complete");
  }
}
