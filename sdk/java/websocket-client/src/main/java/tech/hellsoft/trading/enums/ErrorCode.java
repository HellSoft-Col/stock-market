package tech.hellsoft.trading.enums;

import java.util.Arrays;

public enum ErrorCode {
    AUTH_FAILED("AUTH_FAILED", Severity.FATAL),
    INVALID_ORDER("INVALID_ORDER", Severity.ERROR),
    INVALID_PRODUCT("INVALID_PRODUCT", Severity.ERROR),
    INVALID_QUANTITY("INVALID_QUANTITY", Severity.ERROR),
    DUPLICATE_ORDER_ID("DUPLICATE_ORDER_ID", Severity.ERROR),
    UNAUTHORIZED_PRODUCTION("UNAUTHORIZED_PRODUCTION", Severity.ERROR),
    OFFER_EXPIRED("OFFER_EXPIRED", Severity.INFO),
    RATE_LIMIT_EXCEEDED("RATE_LIMIT_EXCEEDED", Severity.WARNING),
    SERVICE_UNAVAILABLE("SERVICE_UNAVAILABLE", Severity.TRANSIENT),
    INSUFFICIENT_INVENTORY("INSUFFICIENT_INVENTORY", Severity.ERROR),
    INVALID_MESSAGE("INVALID_MESSAGE", Severity.ERROR);
    
    private final String value;
    private final Severity severity;
    
    ErrorCode(String value, Severity severity) {
        this.value = value;
        this.severity = severity;
    }
    
    public String getValue() {
        return value;
    }
    
    public Severity getSeverity() {
        return severity;
    }
    
    public static ErrorCode fromJson(String value) {
        return Arrays.stream(values())
            .filter(code -> code.value.equals(value))
            .findFirst()
            .orElse(null);
    }
    
    public enum Severity {
        FATAL,
        ERROR,
        WARNING,
        INFO,
        TRANSIENT
    }
}
