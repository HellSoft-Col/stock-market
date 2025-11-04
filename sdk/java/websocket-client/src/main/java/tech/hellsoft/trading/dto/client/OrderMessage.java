package tech.hellsoft.trading.dto.client;

import com.google.gson.annotations.SerializedName;

import tech.hellsoft.trading.enums.MessageType;
import tech.hellsoft.trading.enums.OrderMode;
import tech.hellsoft.trading.enums.OrderSide;
import tech.hellsoft.trading.enums.Product;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

/**
 * Message representing a buy or sell order to be sent to the server.
 *
 * <p>This message is used to place orders on the market. All orders must have a unique client order
 * ID (clOrdID) for tracking purposes.
 *
 * <p>Example usage:
 *
 * <pre>{@code
 * OrderMessage order = OrderMessage.builder()
 *     .clOrdID("order-123")
 *     .side(OrderSide.BUY)
 *     .product(Product.GUACA)
 *     .qty(100)
 *     .mode(OrderMode.MARKET)
 *     .limitPrice(50.0) // Only for LIMIT orders
 *     .expiresAt("2024-12-31T23:59:59Z") // Optional expiration
 *     .message("Buy order for guacamole")
 *     .build();
 * }</pre>
 *
 * @see tech.hellsoft.trading.ConectorBolsa
 * @see tech.hellsoft.trading.enums.OrderSide for buy/sell direction
 * @see tech.hellsoft.trading.enums.OrderMode for market/limit order types
 * @see tech.hellsoft.trading.enums.Product for available products
 */
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class OrderMessage {
  /** The message type (automatically set to ORDER). */
  private MessageType type;

  /**
   * Unique client order identifier for tracking.
   *
   * <p>Must be unique across all orders from this client.
   */
  @SerializedName("cl_ord_id")
  private String clOrdID;

  /** The order side (BUY or SELL). */
  private OrderSide side;

  /** The order mode (MARKET or LIMIT). */
  private OrderMode mode;

  /** The product to trade. */
  private Product product;

  /** The quantity to trade (must be positive). */
  private Integer qty;

  /** The limit price (required only for LIMIT orders). */
  @SerializedName("limit_price")
  private Double limitPrice;

  /** Optional expiration timestamp for the order. */
  @SerializedName("expires_at")
  private String expiresAt;

  /** Optional message or comment for the order. */
  private String message;

  /** Optional debug mode for testing. */
  @SerializedName("debug_mode")
  private String debugMode;
}
