package tech.hellsoft.trading.enums;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.MethodSource;

import java.util.stream.Stream;

import static org.junit.jupiter.api.Assertions.*;

class MessageTypeTest {

    static Stream<Arguments> messageTypeValues() {
        return Stream.of(
            Arguments.of(MessageType.LOGIN, "LOGIN"),
            Arguments.of(MessageType.LOGIN_OK, "LOGIN_OK"),
            Arguments.of(MessageType.ORDER, "ORDER"),
            Arguments.of(MessageType.ORDER_ACK, "ORDER_ACK"),
            Arguments.of(MessageType.FILL, "FILL"),
            Arguments.of(MessageType.TICKER, "TICKER"),
            Arguments.of(MessageType.OFFER, "OFFER"),
            Arguments.of(MessageType.ACCEPT_OFFER, "ACCEPT_OFFER"),
            Arguments.of(MessageType.ERROR, "ERROR"),
            Arguments.of(MessageType.CANCEL, "CANCEL"),
            Arguments.of(MessageType.PRODUCTION_UPDATE, "PRODUCTION_UPDATE"),
            Arguments.of(MessageType.INVENTORY_UPDATE, "INVENTORY_UPDATE"),
            Arguments.of(MessageType.BALANCE_UPDATE, "BALANCE_UPDATE"),
            Arguments.of(MessageType.EVENT_DELTA, "EVENT_DELTA"),
            Arguments.of(MessageType.BROADCAST_NOTIFICATION, "BROADCAST_NOTIFICATION"),
            Arguments.of(MessageType.RESYNC, "RESYNC"),
            Arguments.of(MessageType.PING, "PING"),
            Arguments.of(MessageType.PONG, "PONG")
        );
    }

    @ParameterizedTest
    @MethodSource("messageTypeValues")
    void shouldReturnCorrectValue(MessageType type, String expectedValue) {
        assertEquals(expectedValue, type.getValue());
    }

    @ParameterizedTest
    @MethodSource("messageTypeValues")
    void shouldDeserializeFromJson(MessageType expectedType, String jsonValue) {
        MessageType result = MessageType.fromJson(jsonValue);
        assertEquals(expectedType, result);
    }

    @Test
    void shouldThrowExceptionForInvalidValue() {
        assertThrows(IllegalArgumentException.class, () -> 
            MessageType.fromJson("INVALID_TYPE")
        );
    }

    @Test
    void shouldThrowExceptionForNullValue() {
        assertThrows(IllegalArgumentException.class, () -> 
            MessageType.fromJson(null)
        );
    }

    @Test
    void shouldHaveAllValuesUnique() {
        long uniqueCount = Stream.of(MessageType.values())
            .map(MessageType::getValue)
            .distinct()
            .count();
        
        assertEquals(MessageType.values().length, uniqueCount);
    }
}
