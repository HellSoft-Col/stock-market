package tech.hellsoft.trading.enums;

import java.util.Arrays;

/**
 * Enumeration representing the direction of an order.
 *
 * <p>Every order must specify whether it's a buy order (purchasing a product) or a sell order
 * (selling a product). This determines how the order will be matched in the market.
 *
 * @see tech.hellsoft.trading.dto.client.OrderMessage
 */
public enum OrderSide {
  /**
   * Buy order - purchasing a product.
   *
   * <p>Buy orders specify the quantity and maximum price willing to pay.
   */
  BUY("BUY"),

  /**
   * Sell order - selling a product.
   *
   * <p>Sell orders specify the quantity and minimum price willing to accept.
   */
  SELL("SELL");

  private final String value;

  OrderSide(String value) {
    this.value = value;
  }

  /**
   * Gets the JSON string value for this order side.
   *
   * @return the JSON string value
   */
  public String getValue() {
    return value;
  }

  /**
   * Creates an OrderSide from its JSON string value.
   *
   * @param value the JSON string value
   * @return the corresponding OrderSide
   * @throws IllegalArgumentException if the value is unknown
   */
  public static OrderSide fromJson(String value) {
    return Arrays.stream(values())
        .filter(side -> side.value.equals(value))
        .findFirst()
        .orElseThrow(() -> new IllegalArgumentException("Unknown order side: " + value));
  }
}
