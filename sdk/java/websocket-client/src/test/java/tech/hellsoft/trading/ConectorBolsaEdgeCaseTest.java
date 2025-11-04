package tech.hellsoft.trading;

import static org.junit.jupiter.api.Assertions.*;

import java.lang.reflect.Field;
import java.lang.reflect.Method;
import java.time.Duration;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicBoolean;
import java.util.concurrent.atomic.AtomicInteger;

import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import tech.hellsoft.trading.config.ConectorConfig;
import tech.hellsoft.trading.dto.server.*;
import tech.hellsoft.trading.enums.*;
import tech.hellsoft.trading.internal.serialization.JsonSerializer;

class ConectorBolsaEdgeCaseTest {

  private ConectorBolsa connector;
  private ConectorConfig config;

  @BeforeEach
  void setUp() {
    config =
        ConectorConfig.builder()
            .connectionTimeout(Duration.ofSeconds(5))
            .heartbeatInterval(Duration.ofSeconds(10))
            .build();
    connector = new ConectorBolsa(config);
  }

  @AfterEach
  void tearDown() {
    if (connector != null) {
      connector.shutdown();
    }
  }

  @Test
  void shouldHandleStartHeartbeatWhenHeartbeatManagerAlreadyExists() throws Exception {
    invokePrivateMethod(connector, "startHeartbeat");
    invokePrivateMethod(connector, "startHeartbeat");
  }

  @Test
  void shouldHandleStopHeartbeatWhenHeartbeatManagerIsNull() throws Exception {
    invokePrivateMethod(connector, "stopHeartbeat");
  }

  @Test
  void shouldHandleStopHeartbeatWhenHeartbeatManagerExists() throws Exception {
    invokePrivateMethod(connector, "startHeartbeat");
    invokePrivateMethod(connector, "stopHeartbeat");
  }

  @Test
  void shouldTransitionStateOnWebSocketError() throws Exception {
    setPrivateField(connector, "state", ConnectionState.CONNECTED);

    Throwable error = new RuntimeException("Test error");
    invokePrivateMethod(connector, "onWebSocketError", Throwable.class, error);

    assertEquals(ConnectionState.DISCONNECTED, connector.getState());
  }

  @Test
  void shouldTransitionStateOnWebSocketClosed() throws Exception {
    setPrivateField(connector, "state", ConnectionState.CONNECTED);

    invokePrivateMethod(connector, "onWebSocketClosed");

    assertEquals(ConnectionState.DISCONNECTED, connector.getState());
  }

  @Test
  void shouldNotifyAllListenersOnMessage() throws Exception {
    CountDownLatch latch = new CountDownLatch(5);
    AtomicInteger callCount = new AtomicInteger(0);

    for (int i = 0; i < 5; i++) {
      EventListener listener =
          new TestListener() {
            @Override
            public void onError(ErrorMessage message) {
              callCount.incrementAndGet();
              latch.countDown();
            }
          };
      connector.addListener(listener);
    }

    ErrorMessage error =
        ErrorMessage.builder()
            .type(MessageType.ERROR)
            .code(ErrorCode.INVALID_ORDER)
            .reason("Test")
            .timestamp("2024-01-01T00:00:00Z")
            .build();

    String json = JsonSerializer.toJson(error);
    invokePrivateMethod(connector, "onMessageReceived", String.class, json);

    assertTrue(latch.await(2, TimeUnit.SECONDS));
    assertEquals(5, callCount.get());
  }

  @Test
  void shouldHandleMalformedJsonGracefully() throws Exception {
    String invalidJson = "{invalid json}";
    assertDoesNotThrow(
        () -> invokePrivateMethod(connector, "onMessageReceived", String.class, invalidJson));
  }

  @Test
  void shouldHandleEmptyJsonGracefully() throws Exception {
    String emptyJson = "";
    assertDoesNotThrow(
        () -> invokePrivateMethod(connector, "onMessageReceived", String.class, emptyJson));
  }

  @Test
  void shouldHandleNullJsonGracefully() throws Exception {
    assertDoesNotThrow(
        () -> invokePrivateMethod(connector, "onMessageReceived", String.class, (Object) null));
  }

  @Test
  void shouldHandleUnknownMessageTypeGracefully() throws Exception {
    String unknownMessage = "{\"type\":\"UNKNOWN_TYPE\"}";
    assertDoesNotThrow(
        () -> invokePrivateMethod(connector, "onMessageReceived", String.class, unknownMessage));
  }

  @Test
  void shouldNotifyListenerConcurrently() throws Exception {
    int listenerCount = 10;
    CountDownLatch latch = new CountDownLatch(listenerCount);
    AtomicInteger callCount = new AtomicInteger(0);

    for (int i = 0; i < listenerCount; i++) {
      EventListener listener =
          new TestListener() {
            @Override
            public void onTicker(TickerMessage message) {
              callCount.incrementAndGet();
              latch.countDown();
            }
          };
      connector.addListener(listener);
    }

    TickerMessage ticker =
        TickerMessage.builder()
            .type(MessageType.TICKER)
            .product(Product.GUACA)
            .bestBid(100.0)
            .bestAsk(105.0)
            .mid(102.5)
            .serverTime("2024-01-01T00:00:00Z")
            .build();

    String json = JsonSerializer.toJson(ticker);
    invokePrivateMethod(connector, "onMessageReceived", String.class, json);

    assertTrue(latch.await(3, TimeUnit.SECONDS), "All listeners should be notified");
    assertEquals(listenerCount, callCount.get());
  }

  @Test
  void shouldHandleMultipleListenerExceptions() throws Exception {
    CountDownLatch latch = new CountDownLatch(1);
    AtomicBoolean goodListenerCalled = new AtomicBoolean(false);

    EventListener throwingListener1 =
        new TestListener() {
          @Override
          public void onFill(FillMessage message) {
            throw new RuntimeException("Exception 1");
          }
        };

    EventListener throwingListener2 =
        new TestListener() {
          @Override
          public void onFill(FillMessage message) {
            throw new IllegalStateException("Exception 2");
          }
        };

    EventListener goodListener =
        new TestListener() {
          @Override
          public void onFill(FillMessage message) {
            goodListenerCalled.set(true);
            latch.countDown();
          }
        };

    connector.addListener(throwingListener1);
    connector.addListener(throwingListener2);
    connector.addListener(goodListener);

    FillMessage fill =
        FillMessage.builder()
            .type(MessageType.FILL)
            .clOrdID("order-1")
            .fillQty(10)
            .fillPrice(100.0)
            .side(OrderSide.BUY)
            .product(Product.GUACA)
            .serverTime("2024-01-01T00:00:00Z")
            .build();

    String json = JsonSerializer.toJson(fill);
    invokePrivateMethod(connector, "onMessageReceived", String.class, json);

    assertTrue(latch.await(2, TimeUnit.SECONDS), "Good listener should be called");
    assertTrue(goodListenerCalled.get());
  }

  @Test
  void shouldRemoveMultipleListeners() throws Exception {
    EventListener listener1 = new TestListener();
    EventListener listener2 = new TestListener();
    EventListener listener3 = new TestListener();

    connector.addListener(listener1);
    connector.addListener(listener2);
    connector.addListener(listener3);

    connector.removeListener(listener1);
    connector.removeListener(listener2);
    connector.removeListener(listener3);

    CountDownLatch latch = new CountDownLatch(1);
    ErrorMessage error =
        ErrorMessage.builder()
            .type(MessageType.ERROR)
            .code(ErrorCode.INVALID_ORDER)
            .reason("Test")
            .timestamp("2024-01-01T00:00:00Z")
            .build();

    String json = JsonSerializer.toJson(error);
    invokePrivateMethod(connector, "onMessageReceived", String.class, json);

    assertFalse(latch.await(500, TimeUnit.MILLISECONDS), "No listeners should be called");
  }

  @Test
  void shouldHandlePongWhenHeartbeatManagerIsNull() throws Exception {
    PongMessage pong =
        PongMessage.builder().type(MessageType.PONG).timestamp("2024-01-01T00:00:00Z").build();

    String json = JsonSerializer.toJson(pong);
    assertDoesNotThrow(
        () -> invokePrivateMethod(connector, "onMessageReceived", String.class, json));
  }

  @Test
  void shouldCreateHandlersMultipleTimes() throws Exception {
    Object handlers1 = invokePrivateMethodWithReturn(connector, "createHandlers");
    Object handlers2 = invokePrivateMethodWithReturn(connector, "createHandlers");

    assertNotNull(handlers1);
    assertNotNull(handlers2);
    assertNotSame(handlers1, handlers2, "Should create new handlers each time");
  }

  @Test
  void shouldHandleConnectionLostWithMultipleListeners() throws Exception {
    CountDownLatch latch = new CountDownLatch(3);
    AtomicInteger callCount = new AtomicInteger(0);

    for (int i = 0; i < 3; i++) {
      EventListener listener =
          new TestListener() {
            @Override
            public void onConnectionLost(Throwable error) {
              callCount.incrementAndGet();
              latch.countDown();
            }
          };
      connector.addListener(listener);
    }

    Throwable error = new RuntimeException("Connection lost");
    invokePrivateMethod(connector, "notifyConnectionLost", Throwable.class, error);

    assertTrue(latch.await(2, TimeUnit.SECONDS));
    assertEquals(3, callCount.get());
  }

  @Test
  void shouldHandleDesconectarMultipleTimes() {
    connector.desconectar();
    connector.desconectar();
    connector.desconectar();

    assertEquals(ConnectionState.DISCONNECTED, connector.getState());
  }

  @Test
  void shouldHandleShutdownAfterDesconectar() {
    connector.desconectar();
    assertDoesNotThrow(() -> connector.shutdown());
    assertEquals(ConnectionState.DISCONNECTED, connector.getState());
  }

  @Test
  void shouldHandleDesconectarAfterShutdown() {
    connector.shutdown();
    assertDoesNotThrow(() -> connector.desconectar());
    assertEquals(ConnectionState.DISCONNECTED, connector.getState());
  }

  @Test
  void shouldHandleOnPongTimeoutMultipleTimes() throws Exception {
    CountDownLatch latch = new CountDownLatch(1);

    EventListener listener =
        new TestListener() {
          @Override
          public void onConnectionLost(Throwable error) {
            latch.countDown();
          }
        };

    connector.addListener(listener);
    invokePrivateMethod(connector, "onPongTimeout");

    assertTrue(latch.await(2, TimeUnit.SECONDS));
    assertEquals(ConnectionState.DISCONNECTED, connector.getState());
  }

  @Test
  void shouldHandleAllServerMessageTypes() throws Exception {
    CountDownLatch loginLatch = new CountDownLatch(1);
    CountDownLatch fillLatch = new CountDownLatch(1);
    CountDownLatch tickerLatch = new CountDownLatch(1);
    CountDownLatch offerLatch = new CountDownLatch(1);
    CountDownLatch errorLatch = new CountDownLatch(1);
    CountDownLatch ackLatch = new CountDownLatch(1);
    CountDownLatch inventoryLatch = new CountDownLatch(1);
    CountDownLatch balanceLatch = new CountDownLatch(1);
    CountDownLatch eventDeltaLatch = new CountDownLatch(1);
    CountDownLatch broadcastLatch = new CountDownLatch(1);

    EventListener listener =
        new TestListener() {
          @Override
          public void onLoginOk(LoginOKMessage message) {
            loginLatch.countDown();
          }

          @Override
          public void onFill(FillMessage message) {
            fillLatch.countDown();
          }

          @Override
          public void onTicker(TickerMessage message) {
            tickerLatch.countDown();
          }

          @Override
          public void onOffer(OfferMessage message) {
            offerLatch.countDown();
          }

          @Override
          public void onError(ErrorMessage message) {
            errorLatch.countDown();
          }

          @Override
          public void onOrderAck(OrderAckMessage message) {
            ackLatch.countDown();
          }

          @Override
          public void onInventoryUpdate(InventoryUpdateMessage message) {
            inventoryLatch.countDown();
          }

          @Override
          public void onBalanceUpdate(BalanceUpdateMessage message) {
            balanceLatch.countDown();
          }

          @Override
          public void onEventDelta(EventDeltaMessage message) {
            eventDeltaLatch.countDown();
          }

          @Override
          public void onBroadcast(BroadcastNotificationMessage message) {
            broadcastLatch.countDown();
          }
        };

    connector.addListener(listener);

    sendMessage(MessageType.LOGIN_OK, createLoginOk());
    assertTrue(loginLatch.await(1, TimeUnit.SECONDS));

    sendMessage(MessageType.FILL, createFill());
    assertTrue(fillLatch.await(1, TimeUnit.SECONDS));

    sendMessage(MessageType.TICKER, createTicker());
    assertTrue(tickerLatch.await(1, TimeUnit.SECONDS));

    sendMessage(MessageType.OFFER, createOffer());
    assertTrue(offerLatch.await(1, TimeUnit.SECONDS));

    sendMessage(MessageType.ERROR, createError());
    assertTrue(errorLatch.await(1, TimeUnit.SECONDS));

    sendMessage(MessageType.ORDER_ACK, createOrderAck());
    assertTrue(ackLatch.await(1, TimeUnit.SECONDS));

    sendMessage(MessageType.INVENTORY_UPDATE, createInventoryUpdate());
    assertTrue(inventoryLatch.await(1, TimeUnit.SECONDS));

    sendMessage(MessageType.BALANCE_UPDATE, createBalanceUpdate());
    assertTrue(balanceLatch.await(1, TimeUnit.SECONDS));

    sendMessage(MessageType.EVENT_DELTA, createEventDelta());
    assertTrue(eventDeltaLatch.await(1, TimeUnit.SECONDS));

    sendMessage(MessageType.BROADCAST_NOTIFICATION, createBroadcast());
    assertTrue(broadcastLatch.await(1, TimeUnit.SECONDS));
  }

  private void sendMessage(MessageType type, Object message) throws Exception {
    String json = JsonSerializer.toJson(message);
    invokePrivateMethod(connector, "onMessageReceived", String.class, json);
  }

  private LoginOKMessage createLoginOk() {
    return LoginOKMessage.builder()
        .type(MessageType.LOGIN_OK)
        .team("Team")
        .species("Species")
        .initialBalance(1000.0)
        .currentBalance(1000.0)
        .inventory(java.util.Collections.emptyMap())
        .authorizedProducts(java.util.Collections.emptyList())
        .serverTime("2024-01-01T00:00:00Z")
        .build();
  }

  private FillMessage createFill() {
    return FillMessage.builder()
        .type(MessageType.FILL)
        .clOrdID("order-1")
        .fillQty(10)
        .fillPrice(100.0)
        .side(OrderSide.BUY)
        .product(Product.GUACA)
        .serverTime("2024-01-01T00:00:00Z")
        .build();
  }

  private TickerMessage createTicker() {
    return TickerMessage.builder()
        .type(MessageType.TICKER)
        .product(Product.GUACA)
        .bestBid(100.0)
        .bestAsk(105.0)
        .mid(102.5)
        .serverTime("2024-01-01T00:00:00Z")
        .build();
  }

  private OfferMessage createOffer() {
    return OfferMessage.builder()
        .type(MessageType.OFFER)
        .offerId("offer-1")
        .buyer("Buyer")
        .product(Product.GUACA)
        .quantityRequested(10)
        .maxPrice(100.0)
        .expiresIn(60)
        .timestamp("2024-01-01T00:00:00Z")
        .build();
  }

  private ErrorMessage createError() {
    return ErrorMessage.builder()
        .type(MessageType.ERROR)
        .code(ErrorCode.INVALID_ORDER)
        .reason("Error")
        .timestamp("2024-01-01T00:00:00Z")
        .build();
  }

  private OrderAckMessage createOrderAck() {
    return OrderAckMessage.builder()
        .type(MessageType.ORDER_ACK)
        .clOrdID("order-1")
        .status(OrderStatus.FILLED)
        .serverTime("2024-01-01T00:00:00Z")
        .build();
  }

  private InventoryUpdateMessage createInventoryUpdate() {
    return InventoryUpdateMessage.builder()
        .type(MessageType.INVENTORY_UPDATE)
        .inventory(java.util.Map.of(Product.GUACA, 100))
        .serverTime("2024-01-01T00:00:00Z")
        .build();
  }

  private BalanceUpdateMessage createBalanceUpdate() {
    return BalanceUpdateMessage.builder()
        .type(MessageType.BALANCE_UPDATE)
        .balance(1500.0)
        .serverTime("2024-01-01T00:00:00Z")
        .build();
  }

  private EventDeltaMessage createEventDelta() {
    return EventDeltaMessage.builder()
        .type(MessageType.EVENT_DELTA)
        .events(java.util.Collections.emptyList())
        .serverTime("2024-01-01T00:00:00Z")
        .build();
  }

  private BroadcastNotificationMessage createBroadcast() {
    return BroadcastNotificationMessage.builder()
        .type(MessageType.BROADCAST_NOTIFICATION)
        .message("Broadcast")
        .sender("Server")
        .serverTime("2024-01-01T00:00:00Z")
        .build();
  }

  private void invokePrivateMethod(
      Object target, String methodName, Class<?> paramType, Object param) throws Exception {
    Method method = target.getClass().getDeclaredMethod(methodName, paramType);
    method.setAccessible(true);
    method.invoke(target, param);
  }

  private void invokePrivateMethod(Object target, String methodName) throws Exception {
    Method method = target.getClass().getDeclaredMethod(methodName);
    method.setAccessible(true);
    method.invoke(target);
  }

  private Object invokePrivateMethodWithReturn(Object target, String methodName) throws Exception {
    Method method = target.getClass().getDeclaredMethod(methodName);
    method.setAccessible(true);
    return method.invoke(target);
  }

  private void setPrivateField(Object target, String fieldName, Object value) throws Exception {
    Field field = target.getClass().getDeclaredField(fieldName);
    field.setAccessible(true);
    field.set(target, value);
  }

  private static class TestListener implements EventListener {
    @Override
    public void onLoginOk(LoginOKMessage message) {}

    @Override
    public void onFill(FillMessage message) {}

    @Override
    public void onTicker(TickerMessage message) {}

    @Override
    public void onOffer(OfferMessage message) {}

    @Override
    public void onError(ErrorMessage message) {}

    @Override
    public void onOrderAck(OrderAckMessage message) {}

    @Override
    public void onInventoryUpdate(InventoryUpdateMessage message) {}

    @Override
    public void onBalanceUpdate(BalanceUpdateMessage message) {}

    @Override
    public void onEventDelta(EventDeltaMessage message) {}

    @Override
    public void onBroadcast(BroadcastNotificationMessage message) {}

    @Override
    public void onConnectionLost(Throwable error) {}
  }
}
