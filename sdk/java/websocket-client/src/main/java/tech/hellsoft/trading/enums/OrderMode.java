package tech.hellsoft.trading.enums;

import java.util.Arrays;

public enum OrderMode {
    MARKET("MARKET"),
    LIMIT("LIMIT");
    
    private final String value;
    
    OrderMode(String value) {
        this.value = value;
    }
    
    public String getValue() {
        return value;
    }
    
    public static OrderMode fromJson(String value) {
        return Arrays.stream(values())
            .filter(mode -> mode.value.equals(value))
            .findFirst()
            .orElseThrow(() -> new IllegalArgumentException("Unknown order mode: " + value));
    }
}
