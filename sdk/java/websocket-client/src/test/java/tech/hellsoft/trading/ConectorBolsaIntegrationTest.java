package tech.hellsoft.trading;

import static org.junit.jupiter.api.Assertions.*;

import java.lang.reflect.Method;
import java.time.Duration;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicBoolean;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.concurrent.atomic.AtomicReference;

import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import tech.hellsoft.trading.config.ConectorConfig;
import tech.hellsoft.trading.dto.server.*;
import tech.hellsoft.trading.enums.*;
import tech.hellsoft.trading.internal.serialization.JsonSerializer;

class ConectorBolsaIntegrationTest {

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
  void shouldNotifyListenersOnLoginOk() throws Exception {
    CountDownLatch latch = new CountDownLatch(1);
    AtomicReference<LoginOKMessage> received = new AtomicReference<>();

    EventListener listener =
        new TestListener() {
          @Override
          public void onLoginOk(LoginOKMessage message) {
            received.set(message);
            latch.countDown();
          }
        };

    connector.addListener(listener);

    LoginOKMessage loginOk =
        LoginOKMessage.builder()
            .type(MessageType.LOGIN_OK)
            .team("TestTeam")
            .species("TestSpecies")
            .initialBalance(1000.0)
            .currentBalance(1000.0)
            .inventory(java.util.Collections.emptyMap())
            .authorizedProducts(java.util.Collections.emptyList())
            .serverTime("2024-01-01T00:00:00Z")
            .build();

    String json = JsonSerializer.toJson(loginOk);
    invokePrivateMethod(connector, "onMessageReceived", String.class, json);

    assertTrue(latch.await(2, TimeUnit.SECONDS), "Listener should be notified");
    assertNotNull(received.get());
    assertEquals("TestTeam", received.get().getTeam());
  }

  @Test
  void shouldNotifyListenersOnFill() throws Exception {
    CountDownLatch latch = new CountDownLatch(1);
    AtomicReference<FillMessage> received = new AtomicReference<>();

    EventListener listener =
        new TestListener() {
          @Override
          public void onFill(FillMessage message) {
            received.set(message);
            latch.countDown();
          }
        };

    connector.addListener(listener);

    FillMessage fill =
        FillMessage.builder()
            .type(MessageType.FILL)
            .clOrdID("order-1")
            .fillQty(10)
            .fillPrice(100.0)
            .side(OrderSide.BUY)
            .product(Product.GUACA)
            .counterparty("TestCounterparty")
            .serverTime("2024-01-01T00:00:00Z")
            .build();

    String json = JsonSerializer.toJson(fill);
    invokePrivateMethod(connector, "onMessageReceived", String.class, json);

    assertTrue(latch.await(2, TimeUnit.SECONDS), "Listener should be notified");
    assertNotNull(received.get());
    assertEquals("order-1", received.get().getClOrdID());
  }

  @Test
  void shouldNotifyListenersOnTicker() throws Exception {
    CountDownLatch latch = new CountDownLatch(1);
    AtomicReference<TickerMessage> received = new AtomicReference<>();

    EventListener listener =
        new TestListener() {
          @Override
          public void onTicker(TickerMessage message) {
            received.set(message);
            latch.countDown();
          }
        };

    connector.addListener(listener);

    TickerMessage ticker =
        TickerMessage.builder()
            .type(MessageType.TICKER)
            .product(Product.GUACA)
            .bestBid(95.0)
            .bestAsk(105.0)
            .mid(100.0)
            .serverTime("2024-01-01T00:00:00Z")
            .build();

    String json = JsonSerializer.toJson(ticker);
    invokePrivateMethod(connector, "onMessageReceived", String.class, json);

    assertTrue(latch.await(2, TimeUnit.SECONDS), "Listener should be notified");
    assertNotNull(received.get());
    assertEquals(Product.GUACA, received.get().getProduct());
  }

  @Test
  void shouldNotifyListenersOnOffer() throws Exception {
    CountDownLatch latch = new CountDownLatch(1);
    AtomicReference<OfferMessage> received = new AtomicReference<>();

    EventListener listener =
        new TestListener() {
          @Override
          public void onOffer(OfferMessage message) {
            received.set(message);
            latch.countDown();
          }
        };

    connector.addListener(listener);

    OfferMessage offer =
        OfferMessage.builder()
            .type(MessageType.OFFER)
            .offerId("offer-1")
            .buyer("TestBuyer")
            .product(Product.GUACA)
            .quantityRequested(10)
            .maxPrice(100.0)
            .expiresIn(60)
            .timestamp("2024-01-01T00:00:00Z")
            .build();

    String json = JsonSerializer.toJson(offer);
    invokePrivateMethod(connector, "onMessageReceived", String.class, json);

    assertTrue(latch.await(2, TimeUnit.SECONDS), "Listener should be notified");
    assertNotNull(received.get());
    assertEquals("offer-1", received.get().getOfferId());
  }

  @Test
  void shouldNotifyListenersOnError() throws Exception {
    CountDownLatch latch = new CountDownLatch(1);
    AtomicReference<ErrorMessage> received = new AtomicReference<>();

    EventListener listener =
        new TestListener() {
          @Override
          public void onError(ErrorMessage message) {
            received.set(message);
            latch.countDown();
          }
        };

    connector.addListener(listener);

    ErrorMessage error =
        ErrorMessage.builder()
            .type(MessageType.ERROR)
            .code(ErrorCode.INVALID_ORDER)
            .reason("Invalid order quantity")
            .clOrdID("order-1")
            .timestamp("2024-01-01T00:00:00Z")
            .build();

    String json = JsonSerializer.toJson(error);
    invokePrivateMethod(connector, "onMessageReceived", String.class, json);

    assertTrue(latch.await(2, TimeUnit.SECONDS), "Listener should be notified");
    assertNotNull(received.get());
    assertEquals(ErrorCode.INVALID_ORDER, received.get().getCode());
  }

  @Test
  void shouldNotifyListenersOnOrderAck() throws Exception {
    CountDownLatch latch = new CountDownLatch(1);
    AtomicReference<OrderAckMessage> received = new AtomicReference<>();

    EventListener listener =
        new TestListener() {
          @Override
          public void onOrderAck(OrderAckMessage message) {
            received.set(message);
            latch.countDown();
          }
        };

    connector.addListener(listener);

    OrderAckMessage ack =
        OrderAckMessage.builder()
            .type(MessageType.ORDER_ACK)
            .clOrdID("order-1")
            .status(OrderStatus.FILLED)
            .serverTime("2024-01-01T00:00:00Z")
            .build();

    String json = JsonSerializer.toJson(ack);
    invokePrivateMethod(connector, "onMessageReceived", String.class, json);

    assertTrue(latch.await(2, TimeUnit.SECONDS), "Listener should be notified");
    assertNotNull(received.get());
    assertEquals("order-1", received.get().getClOrdID());
  }

  @Test
  void shouldNotifyListenersOnInventoryUpdate() throws Exception {
    CountDownLatch latch = new CountDownLatch(1);
    AtomicReference<InventoryUpdateMessage> received = new AtomicReference<>();

    EventListener listener =
        new TestListener() {
          @Override
          public void onInventoryUpdate(InventoryUpdateMessage message) {
            received.set(message);
            latch.countDown();
          }
        };

    connector.addListener(listener);

    InventoryUpdateMessage inventoryUpdate =
        InventoryUpdateMessage.builder()
            .type(MessageType.INVENTORY_UPDATE)
            .inventory(java.util.Map.of(Product.GUACA, 100))
            .serverTime("2024-01-01T00:00:00Z")
            .build();

    String json = JsonSerializer.toJson(inventoryUpdate);
    invokePrivateMethod(connector, "onMessageReceived", String.class, json);

    assertTrue(latch.await(2, TimeUnit.SECONDS), "Listener should be notified");
    assertNotNull(received.get());
  }

  @Test
  void shouldNotifyListenersOnBalanceUpdate() throws Exception {
    CountDownLatch latch = new CountDownLatch(1);
    AtomicReference<BalanceUpdateMessage> received = new AtomicReference<>();

    EventListener listener =
        new TestListener() {
          @Override
          public void onBalanceUpdate(BalanceUpdateMessage message) {
            received.set(message);
            latch.countDown();
          }
        };

    connector.addListener(listener);

    BalanceUpdateMessage balanceUpdate =
        BalanceUpdateMessage.builder()
            .type(MessageType.BALANCE_UPDATE)
            .balance(1500.0)
            .serverTime("2024-01-01T00:00:00Z")
            .build();

    String json = JsonSerializer.toJson(balanceUpdate);
    invokePrivateMethod(connector, "onMessageReceived", String.class, json);

    assertTrue(latch.await(2, TimeUnit.SECONDS), "Listener should be notified");
    assertNotNull(received.get());
    assertEquals(1500.0, received.get().getBalance());
  }

  @Test
  void shouldNotifyListenersOnEventDelta() throws Exception {
    CountDownLatch latch = new CountDownLatch(1);
    AtomicReference<EventDeltaMessage> received = new AtomicReference<>();

    EventListener listener =
        new TestListener() {
          @Override
          public void onEventDelta(EventDeltaMessage message) {
            received.set(message);
            latch.countDown();
          }
        };

    connector.addListener(listener);

    EventDeltaMessage eventDelta =
        EventDeltaMessage.builder()
            .type(MessageType.EVENT_DELTA)
            .events(java.util.Collections.emptyList())
            .serverTime("2024-01-01T00:00:00Z")
            .build();

    String json = JsonSerializer.toJson(eventDelta);
    invokePrivateMethod(connector, "onMessageReceived", String.class, json);

    assertTrue(latch.await(2, TimeUnit.SECONDS), "Listener should be notified");
    assertNotNull(received.get());
  }

  @Test
  void shouldNotifyListenersOnBroadcast() throws Exception {
    CountDownLatch latch = new CountDownLatch(1);
    AtomicReference<BroadcastNotificationMessage> received = new AtomicReference<>();

    EventListener listener =
        new TestListener() {
          @Override
          public void onBroadcast(BroadcastNotificationMessage message) {
            received.set(message);
            latch.countDown();
          }
        };

    connector.addListener(listener);

    BroadcastNotificationMessage broadcast =
        BroadcastNotificationMessage.builder()
            .type(MessageType.BROADCAST_NOTIFICATION)
            .message("Test broadcast")
            .sender("Server")
            .serverTime("2024-01-01T00:00:00Z")
            .build();

    String json = JsonSerializer.toJson(broadcast);
    invokePrivateMethod(connector, "onMessageReceived", String.class, json);

    assertTrue(latch.await(2, TimeUnit.SECONDS), "Listener should be notified");
    assertNotNull(received.get());
    assertEquals("Test broadcast", received.get().getMessage());
  }

  @Test
  void shouldNotifyMultipleListenersForSameMessage() throws Exception {
    CountDownLatch latch = new CountDownLatch(3);
    AtomicInteger callCount = new AtomicInteger(0);

    EventListener listener1 =
        new TestListener() {
          @Override
          public void onError(ErrorMessage message) {
            callCount.incrementAndGet();
            latch.countDown();
          }
        };

    EventListener listener2 =
        new TestListener() {
          @Override
          public void onError(ErrorMessage message) {
            callCount.incrementAndGet();
            latch.countDown();
          }
        };

    EventListener listener3 =
        new TestListener() {
          @Override
          public void onError(ErrorMessage message) {
            callCount.incrementAndGet();
            latch.countDown();
          }
        };

    connector.addListener(listener1);
    connector.addListener(listener2);
    connector.addListener(listener3);

    ErrorMessage error =
        ErrorMessage.builder()
            .type(MessageType.ERROR)
            .code(ErrorCode.INVALID_ORDER)
            .reason("Test error")
            .timestamp("2024-01-01T00:00:00Z")
            .build();

    String json = JsonSerializer.toJson(error);
    invokePrivateMethod(connector, "onMessageReceived", String.class, json);

    assertTrue(latch.await(2, TimeUnit.SECONDS), "All listeners should be notified");
    assertEquals(3, callCount.get());
  }

  @Test
  void shouldHandleListenerExceptionGracefully() throws Exception {
    CountDownLatch latch = new CountDownLatch(1);
    AtomicBoolean goodListenerCalled = new AtomicBoolean(false);

    EventListener throwingListener =
        new TestListener() {
          @Override
          public void onError(ErrorMessage message) {
            throw new RuntimeException("Intentional test exception");
          }
        };

    EventListener goodListener =
        new TestListener() {
          @Override
          public void onError(ErrorMessage message) {
            goodListenerCalled.set(true);
            latch.countDown();
          }
        };

    connector.addListener(throwingListener);
    connector.addListener(goodListener);

    ErrorMessage error =
        ErrorMessage.builder()
            .type(MessageType.ERROR)
            .code(ErrorCode.INVALID_ORDER)
            .reason("Test error")
            .timestamp("2024-01-01T00:00:00Z")
            .build();

    String json = JsonSerializer.toJson(error);
    invokePrivateMethod(connector, "onMessageReceived", String.class, json);

    assertTrue(
        latch.await(2, TimeUnit.SECONDS), "Good listener should be called despite exception");
    assertTrue(goodListenerCalled.get());
  }

  @Test
  void shouldNotifyConnectionLostOnWebSocketError() throws Exception {
    CountDownLatch latch = new CountDownLatch(1);
    AtomicReference<Throwable> receivedError = new AtomicReference<>();

    EventListener listener =
        new TestListener() {
          @Override
          public void onConnectionLost(Throwable error) {
            receivedError.set(error);
            latch.countDown();
          }
        };

    connector.addListener(listener);

    Throwable testError = new RuntimeException("WebSocket connection failed");
    invokePrivateMethod(connector, "onWebSocketError", Throwable.class, testError);

    assertTrue(latch.await(2, TimeUnit.SECONDS), "Listener should be notified of connection loss");
    assertNotNull(receivedError.get());
    assertEquals("WebSocket connection failed", receivedError.get().getMessage());
    assertEquals(ConnectionState.DISCONNECTED, connector.getState());
  }

  @Test
  void shouldUpdateStateOnWebSocketClosed() throws Exception {
    invokePrivateMethod(connector, "onWebSocketClosed");
    assertEquals(ConnectionState.DISCONNECTED, connector.getState());
  }

  @Test
  void shouldTransitionToAuthenticatedOnLoginOk() throws Exception {
    CountDownLatch latch = new CountDownLatch(1);

    EventListener listener =
        new TestListener() {
          @Override
          public void onLoginOk(LoginOKMessage message) {
            latch.countDown();
          }
        };

    connector.addListener(listener);

    LoginOKMessage loginOk =
        LoginOKMessage.builder()
            .type(MessageType.LOGIN_OK)
            .team("TestTeam")
            .species("TestSpecies")
            .initialBalance(1000.0)
            .currentBalance(1000.0)
            .inventory(java.util.Collections.emptyMap())
            .authorizedProducts(java.util.Collections.emptyList())
            .serverTime("2024-01-01T00:00:00Z")
            .build();

    String json = JsonSerializer.toJson(loginOk);
    invokePrivateMethod(connector, "onMessageReceived", String.class, json);

    assertTrue(latch.await(2, TimeUnit.SECONDS));
  }

  @Test
  void shouldRemoveListenerAndNotReceiveMessages() throws Exception {
    CountDownLatch latch = new CountDownLatch(1);
    AtomicBoolean listenerCalled = new AtomicBoolean(false);

    EventListener listener =
        new TestListener() {
          @Override
          public void onError(ErrorMessage message) {
            listenerCalled.set(true);
            latch.countDown();
          }
        };

    connector.addListener(listener);
    connector.removeListener(listener);

    ErrorMessage error =
        ErrorMessage.builder()
            .type(MessageType.ERROR)
            .code(ErrorCode.INVALID_ORDER)
            .reason("Test error")
            .timestamp("2024-01-01T00:00:00Z")
            .build();

    String json = JsonSerializer.toJson(error);
    invokePrivateMethod(connector, "onMessageReceived", String.class, json);

    assertFalse(latch.await(500, TimeUnit.MILLISECONDS), "Removed listener should not be called");
    assertFalse(listenerCalled.get());
  }

  @Test
  void shouldShutdownGracefully() {
    assertDoesNotThrow(() -> connector.shutdown());
    assertEquals(ConnectionState.DISCONNECTED, connector.getState());
  }

  @Test
  void shouldHandleMultipleShutdowns() {
    connector.shutdown();
    assertDoesNotThrow(() -> connector.shutdown());
  }

  @Test
  void shouldNotifyConnectionLostOnPongTimeout() throws Exception {
    CountDownLatch latch = new CountDownLatch(1);
    AtomicReference<Throwable> receivedError = new AtomicReference<>();

    EventListener listener =
        new TestListener() {
          @Override
          public void onConnectionLost(Throwable error) {
            receivedError.set(error);
            latch.countDown();
          }
        };

    connector.addListener(listener);
    invokePrivateMethod(connector, "onPongTimeout");

    assertTrue(latch.await(2, TimeUnit.SECONDS), "Listener should be notified of pong timeout");
    assertNotNull(receivedError.get());
    assertTrue(receivedError.get().getMessage().contains("Pong timeout"));
    assertEquals(ConnectionState.DISCONNECTED, connector.getState());
  }

  @Test
  void shouldHandleConnectionLostListenerException() throws Exception {
    CountDownLatch latch = new CountDownLatch(1);
    AtomicBoolean goodListenerCalled = new AtomicBoolean(false);

    EventListener throwingListener =
        new TestListener() {
          @Override
          public void onConnectionLost(Throwable error) {
            throw new RuntimeException("Intentional exception in onConnectionLost");
          }
        };

    EventListener goodListener =
        new TestListener() {
          @Override
          public void onConnectionLost(Throwable error) {
            goodListenerCalled.set(true);
            latch.countDown();
          }
        };

    connector.addListener(throwingListener);
    connector.addListener(goodListener);

    Throwable testError = new RuntimeException("Test error");
    invokePrivateMethod(connector, "notifyConnectionLost", Throwable.class, testError);

    assertTrue(
        latch.await(2, TimeUnit.SECONDS), "Good listener should be called despite exception");
    assertTrue(goodListenerCalled.get());
  }

  @Test
  void shouldHandlePongMessageWhenHeartbeatManagerExists() throws Exception {
    PongMessage pong =
        PongMessage.builder().type(MessageType.PONG).timestamp("2024-01-01T00:00:00Z").build();

    String json = JsonSerializer.toJson(pong);
    assertDoesNotThrow(
        () -> invokePrivateMethod(connector, "onMessageReceived", String.class, json));
  }

  @Test
  void shouldDesconectarWhenAlreadyDisconnected() {
    connector.desconectar();
    assertDoesNotThrow(() -> connector.desconectar());
    assertEquals(ConnectionState.DISCONNECTED, connector.getState());
  }

  @Test
  void shouldCallStopHeartbeatOnDisconnect() {
    assertDoesNotThrow(() -> connector.desconectar());
    assertEquals(ConnectionState.DISCONNECTED, connector.getState());
  }

  @Test
  void shouldShutdownSequencerAndExecutor() {
    connector.shutdown();
    assertEquals(ConnectionState.DISCONNECTED, connector.getState());
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
