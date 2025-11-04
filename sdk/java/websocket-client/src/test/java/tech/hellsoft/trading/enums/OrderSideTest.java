package tech.hellsoft.trading.enums;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.MethodSource;

import java.util.stream.Stream;

import static org.junit.jupiter.api.Assertions.*;

class OrderSideTest {

    @Test
    void shouldHaveCorrectNumberOfSides() {
        assertEquals(2, OrderSide.values().length);
    }

    @ParameterizedTest
    @MethodSource("allSides")
    void shouldHaveNonNullValue(OrderSide side) {
        assertNotNull(side.getValue());
        assertFalse(side.getValue().isEmpty());
    }

    @Test
    void shouldHaveBuySide() {
        assertEquals("BUY", OrderSide.BUY.getValue());
    }

    @Test
    void shouldHaveSellSide() {
        assertEquals("SELL", OrderSide.SELL.getValue());
    }

    @ParameterizedTest
    @MethodSource("allSides")
    void shouldSerializeToJson(OrderSide side) {
        String json = side.getValue();
        assertNotNull(json);
        assertEquals(side.getValue(), json);
    }

    @Test
    void shouldDeserializeBuyFromJson() {
        OrderSide result = OrderSide.fromJson("BUY");
        assertEquals(OrderSide.BUY, result);
    }

    @Test
    void shouldDeserializeSellFromJson() {
        OrderSide result = OrderSide.fromJson("SELL");
        assertEquals(OrderSide.SELL, result);
    }

    @Test
    void shouldThrowExceptionForInvalidSide() {
        assertThrows(IllegalArgumentException.class, () ->
            OrderSide.fromJson("INVALID")
        );
    }

    @Test
    void shouldThrowExceptionForNullSide() {
        assertThrows(IllegalArgumentException.class, () ->
            OrderSide.fromJson(null)
        );
    }

    @Test
    void shouldThrowExceptionForEmptySide() {
        assertThrows(IllegalArgumentException.class, () ->
            OrderSide.fromJson("")
        );
    }

    @ParameterizedTest
    @MethodSource("allSides")
    void shouldRoundTripThroughJson(OrderSide side) {
        String json = side.getValue();
        OrderSide result = OrderSide.fromJson(json);
        assertEquals(side, result);
    }

    static Stream<OrderSide> allSides() {
        return Stream.of(OrderSide.values());
    }
}
