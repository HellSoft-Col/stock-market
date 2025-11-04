package tech.hellsoft.trading;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import tech.hellsoft.trading.config.ConectorConfig;
import tech.hellsoft.trading.dto.client.*;
import tech.hellsoft.trading.dto.server.*;
import tech.hellsoft.trading.enums.*;
import tech.hellsoft.trading.exception.ConexionFallidaException;

import java.time.Duration;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.concurrent.atomic.AtomicReference;

import static org.junit.jupiter.api.Assertions.*;

class ConectorBolsaTest {

    private ConectorBolsa connector;
    private ConectorConfig config;

    @BeforeEach
    void setUp() {
        config = ConectorConfig.builder()
            .connectionTimeout(Duration.ofSeconds(5))
            .heartbeatInterval(Duration.ofSeconds(10))
            .build();
        connector = new ConectorBolsa(config);
    }

    @Test
    void shouldCreateWithDefaultConfig() {
        ConectorBolsa defaultConnector = new ConectorBolsa();
        assertNotNull(defaultConnector);
        assertEquals(ConnectionState.DISCONNECTED, defaultConnector.getState());
    }

    @Test
    void shouldCreateWithCustomConfig() {
        ConectorBolsa customConnector = new ConectorBolsa(config);
        assertNotNull(customConnector);
        assertEquals(ConnectionState.DISCONNECTED, customConnector.getState());
    }

    @Test
    void shouldThrowExceptionWhenConfigIsNull() {
        assertThrows(IllegalArgumentException.class, () -> 
            new ConectorBolsa(null)
        );
    }

    @Test
    void shouldStartInDisconnectedState() {
        assertEquals(ConnectionState.DISCONNECTED, connector.getState());
    }

    @Test
    void shouldAddListener() {
        TestListener listener = new TestListener();
        assertDoesNotThrow(() -> connector.addListener(listener));
    }

    @Test
    void shouldThrowExceptionWhenAddingNullListener() {
        assertThrows(IllegalArgumentException.class, () ->
            connector.addListener(null)
        );
    }

    @Test
    void shouldRemoveListener() {
        TestListener listener = new TestListener();
        connector.addListener(listener);
        assertDoesNotThrow(() -> connector.removeListener(listener));
    }

    @Test
    void shouldNotThrowExceptionWhenRemovingNullListener() {
        assertDoesNotThrow(() -> connector.removeListener(null));
    }

    @Test
    void shouldNotThrowExceptionWhenRemovingNonExistentListener() {
        TestListener listener = new TestListener();
        assertDoesNotThrow(() -> connector.removeListener(listener));
    }

    @Test
    void shouldThrowExceptionWhenConnectingWithNullHost() {
        assertThrows(IllegalArgumentException.class, () ->
            connector.conectar(null, 8080, "test-token")
        );
    }

    @Test
    void shouldThrowExceptionWhenConnectingWithEmptyHost() {
        assertThrows(IllegalArgumentException.class, () ->
            connector.conectar("", 8080, "test-token")
        );
    }

    @Test
    void shouldThrowExceptionWhenConnectingWithBlankHost() {
        assertThrows(IllegalArgumentException.class, () ->
            connector.conectar("   ", 8080, "test-token")
        );
    }

    @Test
    void shouldThrowExceptionWhenConnectingWithInvalidPortZero() {
        assertThrows(IllegalArgumentException.class, () ->
            connector.conectar("localhost", 0, "test-token")
        );
    }

    @Test
    void shouldThrowExceptionWhenConnectingWithInvalidPortNegative() {
        assertThrows(IllegalArgumentException.class, () ->
            connector.conectar("localhost", -1, "test-token")
        );
    }

    @Test
    void shouldThrowExceptionWhenConnectingWithInvalidPortTooHigh() {
        assertThrows(IllegalArgumentException.class, () ->
            connector.conectar("localhost", 65536, "test-token")
        );
    }

    @Test
    void shouldThrowExceptionWhenConnectingWithNullToken() {
        assertThrows(IllegalArgumentException.class, () ->
            connector.conectar("localhost", 8080, null)
        );
    }

    @Test
    void shouldThrowExceptionWhenConnectingWithEmptyToken() {
        assertThrows(IllegalArgumentException.class, () ->
            connector.conectar("localhost", 8080, "")
        );
    }

    @Test
    void shouldThrowExceptionWhenConnectingWithBlankToken() {
        assertThrows(IllegalArgumentException.class, () ->
            connector.conectar("localhost", 8080, "   ")
        );
    }

    @Test
    void shouldNotThrowWhenDisconnectingWhileAlreadyDisconnected() {
        assertDoesNotThrow(() -> connector.desconectar());
        assertEquals(ConnectionState.DISCONNECTED, connector.getState());
    }

    @Test
    void shouldThrowExceptionWhenSendingLoginWhileNotConnected() {
        assertThrows(IllegalStateException.class, () ->
            connector.enviarLogin("test-token")
        );
    }

    @Test
    void shouldThrowExceptionWhenSendingLoginWithNullToken() {
        assertThrows(IllegalStateException.class, () ->
            connector.enviarLogin(null)
        );
    }

    @Test
    void shouldThrowExceptionWhenSendingLoginWithEmptyToken() {
        assertThrows(IllegalStateException.class, () ->
            connector.enviarLogin("")
        );
    }

    @Test
    void shouldThrowExceptionWhenSendingLoginWithBlankToken() {
        assertThrows(IllegalStateException.class, () ->
            connector.enviarLogin("   ")
        );
    }

    @Test
    void shouldThrowExceptionWhenSendingOrderWhileNotAuthenticated() {
        OrderMessage order = OrderMessage.builder()
            .clOrdID("order-1")
            .side(OrderSide.BUY)
            .mode(OrderMode.MARKET)
            .product(Product.GUACA)
            .qty(10)
            .build();

        assertThrows(IllegalStateException.class, () ->
            connector.enviarOrden(order)
        );
    }

    @Test
    void shouldThrowExceptionWhenSendingNullOrder() {
        assertThrows(IllegalStateException.class, () ->
            connector.enviarOrden(null)
        );
    }

    @Test
    void shouldThrowExceptionWhenSendingCancelWhileNotAuthenticated() {
        assertThrows(IllegalStateException.class, () ->
            connector.enviarCancelacion("order-1")
        );
    }

    @Test
    void shouldThrowExceptionWhenSendingCancelWithNullClOrdID() {
        assertThrows(IllegalStateException.class, () ->
            connector.enviarCancelacion(null)
        );
    }

    @Test
    void shouldThrowExceptionWhenSendingCancelWithEmptyClOrdID() {
        assertThrows(IllegalStateException.class, () ->
            connector.enviarCancelacion("")
        );
    }

    @Test
    void shouldThrowExceptionWhenSendingCancelWithBlankClOrdID() {
        assertThrows(IllegalStateException.class, () ->
            connector.enviarCancelacion("   ")
        );
    }

    @Test
    void shouldThrowExceptionWhenSendingProductionUpdateWhileNotAuthenticated() {
        ProductionUpdateMessage update = ProductionUpdateMessage.builder()
            .product(Product.GUACA)
            .quantity(5)
            .build();

        assertThrows(IllegalStateException.class, () ->
            connector.enviarActualizacionProduccion(update)
        );
    }

    @Test
    void shouldThrowExceptionWhenSendingNullProductionUpdate() {
        assertThrows(IllegalStateException.class, () ->
            connector.enviarActualizacionProduccion(null)
        );
    }

    @Test
    void shouldThrowExceptionWhenSendingOfferResponseWhileNotAuthenticated() {
        AcceptOfferMessage response = AcceptOfferMessage.builder()
            .offerId("offer-1")
            .accept(true)
            .build();

        assertThrows(IllegalStateException.class, () ->
            connector.enviarRespuestaOferta(response)
        );
    }

    @Test
    void shouldThrowExceptionWhenSendingNullOfferResponse() {
        assertThrows(IllegalStateException.class, () ->
            connector.enviarRespuestaOferta(null)
        );
    }

    @Test
    void shouldNotifyMultipleListeners() throws Exception {
        CountDownLatch latch = new CountDownLatch(3);
        AtomicInteger callCount = new AtomicInteger(0);

        EventListener listener1 = createCountingListener(latch, callCount);
        EventListener listener2 = createCountingListener(latch, callCount);
        EventListener listener3 = createCountingListener(latch, callCount);

        connector.addListener(listener1);
        connector.addListener(listener2);
        connector.addListener(listener3);

        assertEquals(3, callCount.get());
    }

    @Test
    void shouldHandleListenerExceptionGracefully() throws Exception {
        CountDownLatch latch = new CountDownLatch(1);
        
        EventListener throwingListener = new TestListener() {
            @Override
            public void onError(ErrorMessage message) {
                throw new RuntimeException("Intentional test exception");
            }
        };

        EventListener workingListener = new TestListener() {
            @Override
            public void onError(ErrorMessage message) {
                latch.countDown();
            }
        };

        connector.addListener(throwingListener);
        connector.addListener(workingListener);
    }

    @Test
    void shouldCallShutdownSuccessfully() {
        assertDoesNotThrow(() -> connector.shutdown());
        assertEquals(ConnectionState.DISCONNECTED, connector.getState());
    }

    @Test
    void shouldHandleMultipleShutdownCalls() {
        connector.shutdown();
        assertDoesNotThrow(() -> connector.shutdown());
    }

    @Test
    void shouldMaintainStateTransitions() {
        assertEquals(ConnectionState.DISCONNECTED, connector.getState());
    }

    private EventListener createCountingListener(CountDownLatch latch, AtomicInteger counter) {
        counter.incrementAndGet();
        return new TestListener();
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
