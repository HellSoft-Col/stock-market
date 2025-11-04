package tech.hellsoft.trading.enums;

import java.util.Arrays;
import java.util.Set;
import java.util.stream.Collectors;

/**
 * Enumeration of all tradable products in the Stock Market.
 *
 * <p>Each product represents a different commodity that can be bought, sold, produced, or consumed
 * in the market simulation. Products have different properties and may be used in production
 * recipes.
 *
 * <p>Note that some products use hyphens in their JSON representation (e.g., PALTA-OIL) but
 * underscores in the Java enum names.
 */
public enum Product {
  /** Guacamole - a primary food product. */
  GUACA("GUACA"),

  /** Sebo - a raw material product. */
  SEBO("SEBO"),

  /**
   * Avocado oil - processed from avocados.
   *
   * <p>Uses hyphen in JSON: "PALTA-OIL"
   */
  PALTA_OIL("PALTA-OIL"),

  /** Fosfo - a chemical product. */
  FOSFO("FOSFO"),

  /** Nucrem - a nuclear product. */
  NUCREM("NUCREM"),

  /**
   * Cascar alloy - a metal alloy product.
   *
   * <p>Uses hyphen in JSON: "CASCAR-ALLOY"
   */
  CASCAR_ALLOY("CASCAR-ALLOY"),

  /** Pita - a bread product. */
  PITA("PITA"),

  /** GTRON - a technology product. */
  GTRON("GTRON"),

  /**
   * H-GUACA - premium guacamole.
   *
   * <p>Uses hyphen in JSON: "H-GUACA"
   */
  H_GUACA("H-GUACA");

  private final String value;

  Product(String value) {
    this.value = value;
  }

  /**
   * Gets the JSON string value for this product.
   *
   * @return the JSON string value
   */
  public String getValue() {
    return value;
  }

  /**
   * Creates a Product from its JSON string value.
   *
   * @param value the JSON string value
   * @return the corresponding Product
   * @throws IllegalArgumentException if the value is unknown
   */
  public static Product fromJson(String value) {
    return Arrays.stream(values())
        .filter(p -> p.value.equals(value))
        .findFirst()
        .orElseThrow(() -> new IllegalArgumentException("Unknown product: " + value));
  }

  /**
   * Gets all possible JSON values for products.
   *
   * @return set of all product JSON values
   */
  public static Set<String> getAllValues() {
    return Arrays.stream(values()).map(Product::getValue).collect(Collectors.toSet());
  }
}
