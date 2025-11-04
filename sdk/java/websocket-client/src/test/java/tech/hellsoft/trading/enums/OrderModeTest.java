package tech.hellsoft.trading.enums;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.MethodSource;

import java.util.stream.Stream;

import static org.junit.jupiter.api.Assertions.*;

class OrderModeTest {

    @Test
    void shouldHaveCorrectNumberOfModes() {
        assertEquals(2, OrderMode.values().length);
    }

    @ParameterizedTest
    @MethodSource("allModes")
    void shouldHaveNonNullValue(OrderMode mode) {
        assertNotNull(mode.getValue());
        assertFalse(mode.getValue().isEmpty());
    }

    @Test
    void shouldHaveMarketMode() {
        assertEquals("MARKET", OrderMode.MARKET.getValue());
    }

    @Test
    void shouldHaveLimitMode() {
        assertEquals("LIMIT", OrderMode.LIMIT.getValue());
    }

    @ParameterizedTest
    @MethodSource("allModes")
    void shouldSerializeToJson(OrderMode mode) {
        String json = mode.getValue();
        assertNotNull(json);
        assertEquals(mode.getValue(), json);
    }

    @Test
    void shouldDeserializeMarketFromJson() {
        OrderMode result = OrderMode.fromJson("MARKET");
        assertEquals(OrderMode.MARKET, result);
    }

    @Test
    void shouldDeserializeLimitFromJson() {
        OrderMode result = OrderMode.fromJson("LIMIT");
        assertEquals(OrderMode.LIMIT, result);
    }

    @Test
    void shouldThrowExceptionForInvalidMode() {
        assertThrows(IllegalArgumentException.class, () ->
            OrderMode.fromJson("INVALID")
        );
    }

    @Test
    void shouldThrowExceptionForNullMode() {
        assertThrows(IllegalArgumentException.class, () ->
            OrderMode.fromJson(null)
        );
    }

    @Test
    void shouldThrowExceptionForEmptyMode() {
        assertThrows(IllegalArgumentException.class, () ->
            OrderMode.fromJson("")
        );
    }

    @ParameterizedTest
    @MethodSource("allModes")
    void shouldRoundTripThroughJson(OrderMode mode) {
        String json = mode.getValue();
        OrderMode result = OrderMode.fromJson(json);
        assertEquals(mode, result);
    }

    static Stream<OrderMode> allModes() {
        return Stream.of(OrderMode.values());
    }
}
