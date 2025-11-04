package tech.hellsoft.trading.enums;

import java.util.Arrays;

/**
 * Enumeration representing the execution mode of an order.
 *
 * <p>Orders can be executed either at market price (immediate execution) or at a specified limit
 * price (execution only at that price or better).
 *
 * @see tech.hellsoft.trading.dto.client.OrderMessage
 */
public enum OrderMode {
  /**
   * Market order - executes immediately at current market price.
   *
   * <p>Market orders have guaranteed execution but the price is not guaranteed. They are filled at
   * the best available price.
   */
  MARKET("MARKET"),

  /**
   * Limit order - executes only at specified price or better.
   *
   * <p>Limit orders have guaranteed price but execution is not guaranteed. Buy limit orders execute
   * at the specified price or lower. Sell limit orders execute at the specified price or higher.
   */
  LIMIT("LIMIT");

  private final String value;

  OrderMode(String value) {
    this.value = value;
  }

  /**
   * Gets the JSON string value for this order mode.
   *
   * @return the JSON string value
   */
  public String getValue() {
    return value;
  }

  /**
   * Creates an OrderMode from its JSON string value.
   *
   * @param value the JSON string value
   * @return the corresponding OrderMode
   * @throws IllegalArgumentException if the value is unknown
   */
  public static OrderMode fromJson(String value) {
    return Arrays.stream(values())
        .filter(mode -> mode.value.equals(value))
        .findFirst()
        .orElseThrow(() -> new IllegalArgumentException("Unknown order mode: " + value));
  }
}
