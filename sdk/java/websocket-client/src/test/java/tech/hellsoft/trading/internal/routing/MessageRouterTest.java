package tech.hellsoft.trading.internal.routing;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.NullAndEmptySource;
import org.junit.jupiter.params.provider.ValueSource;
import org.mockito.ArgumentCaptor;
import tech.hellsoft.trading.dto.server.*;
import tech.hellsoft.trading.enums.MessageType;
import tech.hellsoft.trading.enums.Product;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.Mockito.*;

class MessageRouterTest {

    private MessageRouter router;
    private MessageRouter.MessageHandlers handlers;

    @BeforeEach
    void setUp() {
        router = new MessageRouter();
        handlers = mock(MessageRouter.MessageHandlers.class);
    }

    @Test
    void shouldRouteLoginOKMessage() {
        String json = "{\"type\":\"LOGIN_OK\",\"team\":\"TestTeam\",\"species\":\"TestSpecies\",\"initial_balance\":10000.0}";

        router.routeMessage(json, handlers);

        ArgumentCaptor<LoginOKMessage> captor = ArgumentCaptor.forClass(LoginOKMessage.class);
        verify(handlers).onLoginOk(captor.capture());
        
        LoginOKMessage captured = captor.getValue();
        assertEquals(MessageType.LOGIN_OK, captured.getType());
    }

    @Test
    void shouldRouteFillMessage() {
        String json = "{\"type\":\"FILL\",\"cl_ord_id\":\"ORDER-123\",\"fill_qty\":10,\"fill_price\":100.5,\"side\":\"BUY\",\"product\":\"GUACA\"}";

        router.routeMessage(json, handlers);

        ArgumentCaptor<FillMessage> captor = ArgumentCaptor.forClass(FillMessage.class);
        verify(handlers).onFill(captor.capture());
        
        FillMessage captured = captor.getValue();
        assertEquals(MessageType.FILL, captured.getType());
    }

    @Test
    void shouldRouteTickerMessage() {
        String json = "{\"type\":\"TICKER\",\"product\":\"GUACA\",\"best_bid\":99.5,\"best_ask\":100.5,\"mid\":100.0}";

        router.routeMessage(json, handlers);

        ArgumentCaptor<TickerMessage> captor = ArgumentCaptor.forClass(TickerMessage.class);
        verify(handlers).onTicker(captor.capture());
        
        TickerMessage captured = captor.getValue();
        assertEquals(MessageType.TICKER, captured.getType());
        assertEquals(Product.GUACA, captured.getProduct());
    }

    @Test
    void shouldRouteOfferMessage() {
        String json = "{\"type\":\"OFFER\",\"offer_id\":\"OFFER-123\",\"buyer\":\"BuyerTeam\",\"product\":\"SEBO\",\"quantity_requested\":50}";

        router.routeMessage(json, handlers);

        ArgumentCaptor<OfferMessage> captor = ArgumentCaptor.forClass(OfferMessage.class);
        verify(handlers).onOffer(captor.capture());
        
        OfferMessage captured = captor.getValue();
        assertEquals(MessageType.OFFER, captured.getType());
    }

    @Test
    void shouldRouteErrorMessage() {
        String json = "{\"type\":\"ERROR\",\"code\":\"INVALID_ORDER\",\"reason\":\"Order validation failed\"}";

        router.routeMessage(json, handlers);

        ArgumentCaptor<ErrorMessage> captor = ArgumentCaptor.forClass(ErrorMessage.class);
        verify(handlers).onError(captor.capture());
        
        ErrorMessage captured = captor.getValue();
        assertEquals(MessageType.ERROR, captured.getType());
    }

    @Test
    void shouldRouteOrderAckMessage() {
        String json = "{\"type\":\"ORDER_ACK\",\"cl_ord_id\":\"ORDER-123\",\"status\":\"ACCEPTED\"}";

        router.routeMessage(json, handlers);

        ArgumentCaptor<OrderAckMessage> captor = ArgumentCaptor.forClass(OrderAckMessage.class);
        verify(handlers).onOrderAck(captor.capture());
        
        OrderAckMessage captured = captor.getValue();
        assertEquals(MessageType.ORDER_ACK, captured.getType());
    }

    @Test
    void shouldRouteInventoryUpdateMessage() {
        String json = "{\"type\":\"INVENTORY_UPDATE\",\"inventory\":{\"GUACA\":100,\"SEBO\":50}}";

        router.routeMessage(json, handlers);

        ArgumentCaptor<InventoryUpdateMessage> captor = ArgumentCaptor.forClass(InventoryUpdateMessage.class);
        verify(handlers).onInventoryUpdate(captor.capture());
        
        InventoryUpdateMessage captured = captor.getValue();
        assertEquals(MessageType.INVENTORY_UPDATE, captured.getType());
    }

    @Test
    void shouldRouteBalanceUpdateMessage() {
        String json = "{\"type\":\"BALANCE_UPDATE\",\"balance\":15000.0}";

        router.routeMessage(json, handlers);

        ArgumentCaptor<BalanceUpdateMessage> captor = ArgumentCaptor.forClass(BalanceUpdateMessage.class);
        verify(handlers).onBalanceUpdate(captor.capture());
        
        BalanceUpdateMessage captured = captor.getValue();
        assertEquals(MessageType.BALANCE_UPDATE, captured.getType());
    }

    @Test
    void shouldRouteEventDeltaMessage() {
        String json = "{\"type\":\"EVENT_DELTA\",\"events\":[]}";

        router.routeMessage(json, handlers);

        ArgumentCaptor<EventDeltaMessage> captor = ArgumentCaptor.forClass(EventDeltaMessage.class);
        verify(handlers).onEventDelta(captor.capture());
        
        EventDeltaMessage captured = captor.getValue();
        assertEquals(MessageType.EVENT_DELTA, captured.getType());
    }

    @Test
    void shouldRouteBroadcastNotificationMessage() {
        String json = "{\"type\":\"BROADCAST_NOTIFICATION\",\"message\":\"Market closing soon\",\"sender\":\"SYSTEM\"}";

        router.routeMessage(json, handlers);

        ArgumentCaptor<BroadcastNotificationMessage> captor = ArgumentCaptor.forClass(BroadcastNotificationMessage.class);
        verify(handlers).onBroadcast(captor.capture());
        
        BroadcastNotificationMessage captured = captor.getValue();
        assertEquals(MessageType.BROADCAST_NOTIFICATION, captured.getType());
    }

    @Test
    void shouldRoutePongMessage() {
        String json = "{\"type\":\"PONG\",\"timestamp\":\"2024-01-01T12:00:00Z\"}";

        router.routeMessage(json, handlers);

        ArgumentCaptor<PongMessage> captor = ArgumentCaptor.forClass(PongMessage.class);
        verify(handlers).onPong(captor.capture());
        
        PongMessage captured = captor.getValue();
        assertEquals(MessageType.PONG, captured.getType());
    }

    @ParameterizedTest
    @NullAndEmptySource
    @ValueSource(strings = {" ", "\t", "\n"})
    void shouldHandleInvalidJsonGracefully(String json) {
        router.routeMessage(json, handlers);

        verifyNoInteractions(handlers);
    }

    @Test
    void shouldHandleMessageWithoutType() {
        String json = "{\"data\":\"some data\"}";

        router.routeMessage(json, handlers);

        verifyNoInteractions(handlers);
    }

    @Test
    void shouldHandleUnknownMessageType() {
        String json = "{\"type\":\"UNKNOWN_TYPE\"}";

        assertDoesNotThrow(() ->
            router.routeMessage(json, handlers)
        );

        verifyNoInteractions(handlers);
    }

    @Test
    void shouldHandleClientMessageTypesGracefully() {
        String json = "{\"type\":\"ORDER\"}";

        assertDoesNotThrow(() ->
            router.routeMessage(json, handlers)
        );
    }

    @Test
    void shouldHandleCancelMessageTypeGracefully() {
        String json = "{\"type\":\"CANCEL\"}";

        assertDoesNotThrow(() ->
            router.routeMessage(json, handlers)
        );
    }

    @Test
    void shouldHandleResyncMessageTypeGracefully() {
        String json = "{\"type\":\"RESYNC\"}";

        assertDoesNotThrow(() ->
            router.routeMessage(json, handlers)
        );
    }

    @Test
    void shouldHandleMalformedJson() {
        String json = "{invalid json}";

        router.routeMessage(json, handlers);

        verifyNoInteractions(handlers);
    }

    @Test
    void shouldHandleExceptionInHandler() {
        String json = "{\"type\":\"FILL\",\"cl_ord_id\":\"ORDER-123\"}";

        doThrow(new RuntimeException("Handler exception"))
            .when(handlers).onFill(any(FillMessage.class));

        assertDoesNotThrow(() -> router.routeMessage(json, handlers));
    }

    @Test
    void shouldRouteMultipleMessagesSequentially() {
        String loginJson = "{\"type\":\"LOGIN_OK\",\"team\":\"TestTeam\"}";
        String tickerJson = "{\"type\":\"TICKER\",\"product\":\"GUACA\"}";
        String fillJson = "{\"type\":\"FILL\",\"cl_ord_id\":\"ORDER-123\"}";

        router.routeMessage(loginJson, handlers);
        router.routeMessage(tickerJson, handlers);
        router.routeMessage(fillJson, handlers);

        verify(handlers).onLoginOk(any(LoginOKMessage.class));
        verify(handlers).onTicker(any(TickerMessage.class));
        verify(handlers).onFill(any(FillMessage.class));
    }

    @Test
    void shouldHandleProductsWithHyphens() {
        String json = "{\"type\":\"TICKER\",\"product\":\"PALTA-OIL\",\"mid\":100.0}";

        router.routeMessage(json, handlers);

        ArgumentCaptor<TickerMessage> captor = ArgumentCaptor.forClass(TickerMessage.class);
        verify(handlers).onTicker(captor.capture());
        
        TickerMessage captured = captor.getValue();
        assertNotNull(captured);
        assertEquals(MessageType.TICKER, captured.getType());
    }

    @Test
    void shouldNotInvokeOtherHandlersWhenRoutingSpecificMessage() {
        String json = "{\"type\":\"FILL\",\"cl_ord_id\":\"ORDER-123\"}";

        router.routeMessage(json, handlers);

        verify(handlers).onFill(any(FillMessage.class));
        verify(handlers, never()).onLoginOk(any());
        verify(handlers, never()).onTicker(any());
        verify(handlers, never()).onOffer(any());
        verify(handlers, never()).onError(any());
    }

    @Test
    void shouldHandleNullHandlersGracefully() {
        String json = "{\"type\":\"FILL\",\"cl_ord_id\":\"ORDER-123\"}";

        assertDoesNotThrow(() ->
            router.routeMessage(json, null)
        );
    }

    @Test
    void shouldHandleComplexNestedJson() {
        String json = "{\"type\":\"LOGIN_OK\",\"team\":\"TestTeam\",\"species\":\"TestSpecies\"," +
                     "\"initial_balance\":10000.0,\"current_balance\":9500.0," +
                     "\"inventory\":{\"GUACA\":50,\"SEBO\":25}," +
                     "\"authorized_products\":[\"GUACA\",\"SEBO\"]}";

        router.routeMessage(json, handlers);

        verify(handlers).onLoginOk(any(LoginOKMessage.class));
    }
}
