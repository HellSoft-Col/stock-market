package tech.hellsoft.trading.enums;

import java.util.Arrays;
import java.util.Set;
import java.util.stream.Collectors;

public enum Product {
  GUACA("GUACA"),
  SEBO("SEBO"),
  PALTA_OIL("PALTA-OIL"),
  FOSFO("FOSFO"),
  NUCREM("NUCREM"),
  CASCAR_ALLOY("CASCAR-ALLOY"),
  PITA("PITA");

  private final String value;

  Product(String value) {
    this.value = value;
  }

  public String getValue() {
    return value;
  }

  public static Product fromJson(String value) {
    return Arrays.stream(values())
        .filter(p -> p.value.equals(value))
        .findFirst()
        .orElseThrow(() -> new IllegalArgumentException("Unknown product: " + value));
  }

  public static Set<String> getAllValues() {
    return Arrays.stream(values()).map(Product::getValue).collect(Collectors.toSet());
  }
}
