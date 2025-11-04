package tech.hellsoft.trading.enums;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.MethodSource;

import java.util.stream.Stream;

import static org.junit.jupiter.api.Assertions.*;

class ErrorCodeTest {

    static Stream<Arguments> errorCodeValues() {
        return Stream.of(
            Arguments.of(ErrorCode.AUTH_FAILED, "AUTH_FAILED", ErrorCode.Severity.FATAL),
            Arguments.of(ErrorCode.INVALID_ORDER, "INVALID_ORDER", ErrorCode.Severity.ERROR),
            Arguments.of(ErrorCode.INVALID_PRODUCT, "INVALID_PRODUCT", ErrorCode.Severity.ERROR),
            Arguments.of(ErrorCode.INVALID_QUANTITY, "INVALID_QUANTITY", ErrorCode.Severity.ERROR),
            Arguments.of(ErrorCode.DUPLICATE_ORDER_ID, "DUPLICATE_ORDER_ID", ErrorCode.Severity.ERROR),
            Arguments.of(ErrorCode.UNAUTHORIZED_PRODUCTION, "UNAUTHORIZED_PRODUCTION", ErrorCode.Severity.ERROR),
            Arguments.of(ErrorCode.OFFER_EXPIRED, "OFFER_EXPIRED", ErrorCode.Severity.INFO),
            Arguments.of(ErrorCode.RATE_LIMIT_EXCEEDED, "RATE_LIMIT_EXCEEDED", ErrorCode.Severity.WARNING),
            Arguments.of(ErrorCode.SERVICE_UNAVAILABLE, "SERVICE_UNAVAILABLE", ErrorCode.Severity.TRANSIENT),
            Arguments.of(ErrorCode.INSUFFICIENT_INVENTORY, "INSUFFICIENT_INVENTORY", ErrorCode.Severity.ERROR),
            Arguments.of(ErrorCode.INVALID_MESSAGE, "INVALID_MESSAGE", ErrorCode.Severity.ERROR)
        );
    }

    @ParameterizedTest
    @MethodSource("errorCodeValues")
    void shouldReturnCorrectValue(ErrorCode code, String expectedValue, ErrorCode.Severity expectedSeverity) {
        assertEquals(expectedValue, code.getValue());
        assertEquals(expectedSeverity, code.getSeverity());
    }

    @ParameterizedTest
    @MethodSource("errorCodeValues")
    void shouldDeserializeFromJson(ErrorCode expectedCode, String jsonValue, ErrorCode.Severity severity) {
        ErrorCode result = ErrorCode.fromJson(jsonValue);
        assertEquals(expectedCode, result);
        assertEquals(severity, result.getSeverity());
    }

    @Test
    void shouldHaveFatalSeverityForAuthFailed() {
        assertEquals(ErrorCode.Severity.FATAL, ErrorCode.AUTH_FAILED.getSeverity());
    }

    @Test
    void shouldReturnNullForInvalidValue() {
        assertNull(ErrorCode.fromJson("INVALID_CODE"));
    }

    @Test
    void shouldReturnNullForNullValue() {
        assertNull(ErrorCode.fromJson(null));
    }

    @Test
    void shouldHaveAllSeveritiesRepresented() {
        assertTrue(Stream.of(ErrorCode.values()).anyMatch(c -> c.getSeverity() == ErrorCode.Severity.FATAL));
        assertTrue(Stream.of(ErrorCode.values()).anyMatch(c -> c.getSeverity() == ErrorCode.Severity.ERROR));
        assertTrue(Stream.of(ErrorCode.values()).anyMatch(c -> c.getSeverity() == ErrorCode.Severity.WARNING));
        assertTrue(Stream.of(ErrorCode.values()).anyMatch(c -> c.getSeverity() == ErrorCode.Severity.INFO));
        assertTrue(Stream.of(ErrorCode.values()).anyMatch(c -> c.getSeverity() == ErrorCode.Severity.TRANSIENT));
    }
}
