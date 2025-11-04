package tech.hellsoft.trading.enums;

import java.util.Arrays;

public enum OrderSide {
    BUY("BUY"),
    SELL("SELL");
    
    private final String value;
    
    OrderSide(String value) {
        this.value = value;
    }
    
    public String getValue() {
        return value;
    }
    
    public static OrderSide fromJson(String value) {
        return Arrays.stream(values())
            .filter(side -> side.value.equals(value))
            .findFirst()
            .orElseThrow(() -> new IllegalArgumentException("Unknown order side: " + value));
    }
}
