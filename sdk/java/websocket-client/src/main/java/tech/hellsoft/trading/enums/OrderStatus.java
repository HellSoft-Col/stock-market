package tech.hellsoft.trading.enums;

import java.util.Arrays;

public enum OrderStatus {
    PENDING("PENDING"),
    FILLED("FILLED"),
    PARTIALLY_FILLED("PARTIALLY_FILLED"),
    CANCELLED("CANCELLED");
    
    private final String value;
    
    OrderStatus(String value) {
        this.value = value;
    }
    
    public String getValue() {
        return value;
    }
    
    public static OrderStatus fromJson(String value) {
        return Arrays.stream(values())
            .filter(status -> status.value.equals(value))
            .findFirst()
            .orElseThrow(() -> new IllegalArgumentException("Unknown order status: " + value));
    }
}
