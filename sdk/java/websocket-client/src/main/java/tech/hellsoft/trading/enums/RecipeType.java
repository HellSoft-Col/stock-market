package tech.hellsoft.trading.enums;

import java.util.Arrays;

public enum RecipeType {
    BASIC("BASIC"),
    PREMIUM("PREMIUM");
    
    private final String value;
    
    RecipeType(String value) {
        this.value = value;
    }
    
    public String getValue() {
        return value;
    }
    
    public static RecipeType fromJson(String value) {
        return Arrays.stream(values())
            .filter(type -> type.value.equals(value))
            .findFirst()
            .orElseThrow(() -> new IllegalArgumentException("Unknown recipe type: " + value));
    }
}
