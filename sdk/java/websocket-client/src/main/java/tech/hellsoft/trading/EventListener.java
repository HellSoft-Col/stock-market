package tech.hellsoft.trading;

import tech.hellsoft.trading.dto.server.*;

/**
 * Callback interface for receiving server messages and connection events.
 *
 * <p>Implement this interface to handle asynchronous messages from the Stock Market server. All
 * callback methods are executed on virtual threads to avoid blocking the WebSocket processing
 * thread.
 *
 * <p>Example implementation:
 *
 * <pre>{@code
 * public class MyEventListener implements EventListener {
 *     @Override
 *     public void onLoginOk(LoginOKMessage message) {
 *         System.out.println("Logged in as team: " + message.getTeam());
 *         System.out.println("Balance: " + message.getInitialBalance());
 *     }
 *
 *     @Override
 *     public void onFill(FillMessage message) {
 *         System.out.println("Order filled: " + message.getFillQty() +
 *                           " of " + message.getProduct());
 *     }
 *
 *     // Implement other methods as needed
 * }
 * }</pre>
 *
 * @see ConectorBolsa#addListener(EventListener) to register for callbacks
 */
public interface EventListener {

  /**
   * Called when login authentication is successful.
   *
   * @param message the login confirmation with session details
   */
  void onLoginOk(LoginOKMessage message);

  /**
   * Called when an order is partially or fully filled.
   *
   * @param message the fill details including quantity and price
   */
  void onFill(FillMessage message);

  /**
   * Called periodically with market ticker data.
   *
   * @param message the current market prices and statistics
   */
  void onTicker(TickerMessage message);

  /**
   * Called when another party makes an offer to your team.
   *
   * @param message the offer details including terms and expiration
   */
  void onOffer(OfferMessage message);

  /**
   * Called when the server reports an error.
   *
   * @param message the error details including error code and description
   */
  void onError(ErrorMessage message);

  /**
   * Called when the server acknowledges an order or cancellation.
   *
   * @param message the acknowledgment with order status and details
   */
  void onOrderAck(OrderAckMessage message);

  /**
   * Called when your inventory levels change.
   *
   * @param message the inventory update with product quantities
   */
  void onInventoryUpdate(InventoryUpdateMessage message);

  /**
   * Called when your account balance changes.
   *
   * @param message the balance update with new balance amount
   */
  void onBalanceUpdate(BalanceUpdateMessage message);

  /**
   * Called when significant market events occur.
   *
   * @param message the event details and market impact
   */
  void onEventDelta(EventDeltaMessage message);

  /**
   * Called when the server broadcasts notifications to all participants.
   *
   * @param message the broadcast notification content
   */
  void onBroadcast(BroadcastNotificationMessage message);

  /**
   * Called when the WebSocket connection is lost unexpectedly.
   *
   * <p>This method is NOT called for intentional disconnections via {@link
   * ConectorBolsa#desconectar()}.
   *
   * @param error the cause of the connection loss
   */
  void onConnectionLost(Throwable error);

  /**
   * Called when you need to show the globar performance report
   *
   * @param message the message for global report
   */
  void onGlobalPerformanceReport(GlobalPerformanceReportMessage message);
}
