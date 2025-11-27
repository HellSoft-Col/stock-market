package tech.hellsoft.trading.util;

import java.util.concurrent.atomic.AtomicLong;

/**
 * Thread-safe utility for generating unique order IDs.
 *
 * <p>This class provides atomic counter-based order ID generation that is safe for use across
 * multiple threads and concurrent trading strategies. Using this generator prevents duplicate order
 * ID errors that can occur when multiple orders are placed simultaneously.
 *
 * <p><b>Best Practice Usage:</b>
 *
 * <pre>{@code
 * public class MyTradingBot extends EventListenerAdapter {
 *     private final OrderIdGenerator orderIdGen = new OrderIdGenerator("BOT");
 *
 *     public void placeOrder() {
 *         OrderMessage order = OrderMessage.builder()
 *             .clOrdID(orderIdGen.next())  // Thread-safe unique ID
 *             .side(OrderSide.BUY)
 *             .product(Product.GUACA)
 *             .qty(10)
 *             .mode(OrderMode.LIMIT)
 *             .limitPrice(15.50)
 *             .build();
 *         connector.enviarOrden(order);
 *     }
 * }
 * }</pre>
 *
 * <p><b>Format:</b> Generated IDs follow the pattern: {@code PREFIX-TIMESTAMP-COUNTER}
 *
 * <ul>
 *   <li>{@code PREFIX}: Custom identifier (e.g., team name, strategy name)
 *   <li>{@code TIMESTAMP}: Milliseconds since epoch (for sorting/debugging)
 *   <li>{@code COUNTER}: Atomic counter (ensures uniqueness within same millisecond)
 * </ul>
 *
 * <p>Example generated IDs:
 *
 * <ul>
 *   <li>{@code TEAM-A-1732475123456-1}
 *   <li>{@code TEAM-A-1732475123456-2}
 *   <li>{@code STRAT-MM-1732475123457-1}
 * </ul>
 *
 * <p><b>Thread Safety:</b> All methods are thread-safe and lock-free using {@link AtomicLong} for
 * the counter.
 *
 * @see tech.hellsoft.trading.dto.client.OrderMessage
 */
public class OrderIdGenerator {
  private final String prefix;
  private final AtomicLong counter = new AtomicLong(0);

  /**
   * Creates a new order ID generator with the specified prefix.
   *
   * @param prefix the prefix to use for all generated IDs (e.g., team name)
   * @throws IllegalArgumentException if prefix is null or empty
   */
  public OrderIdGenerator(String prefix) {
    if (prefix == null || prefix.trim().isEmpty()) {
      throw new IllegalArgumentException("Prefix cannot be null or empty");
    }
    this.prefix = prefix.trim();
  }

  /**
   * Generates the next unique order ID.
   *
   * <p>Format: {@code PREFIX-TIMESTAMP-COUNTER}
   *
   * <p>This method is thread-safe and can be called concurrently from multiple threads. Each call
   * is guaranteed to return a unique ID.
   *
   * @return a unique order ID string
   */
  public String next() {
    long timestamp = System.currentTimeMillis();
    long count = counter.incrementAndGet();
    return String.format("%s-%d-%d", prefix, timestamp, count);
  }

  /**
   * Generates the next unique order ID using increment-and-get semantics.
   *
   * <p>This method is equivalent to {@link #next()} and provided for consistency with {@link
   * AtomicLong#incrementAndGet()} naming conventions.
   *
   * @return a unique order ID string
   * @see #next()
   */
  public String incrementAndGet() {
    return next();
  }

  /**
   * Generates the next unique order ID using get-and-increment semantics.
   *
   * <p>Note: This method returns the same result as {@link #incrementAndGet()}. The counter is
   * incremented before ID generation, so both methods are equivalent.
   *
   * @return a unique order ID string
   * @see #next()
   */
  public String getAndIncrement() {
    return next();
  }

  /**
   * Gets the current counter value without generating an ID.
   *
   * <p>This is primarily useful for debugging or monitoring how many IDs have been generated.
   *
   * @return the current counter value
   */
  public long getCurrentCount() {
    return counter.get();
  }

  /**
   * Resets the counter to zero.
   *
   * <p><b>Warning:</b> This method should rarely be used in production code as it may lead to
   * duplicate IDs if called while orders are being placed. It's primarily intended for testing
   * purposes.
   */
  public void reset() {
    counter.set(0);
  }
}
