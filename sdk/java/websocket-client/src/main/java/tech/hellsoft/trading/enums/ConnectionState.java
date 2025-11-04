package tech.hellsoft.trading.enums;

/**
 * Represents the current connection state of the WebSocket client.
 *
 * <p>The connection state transitions through these phases during the lifecycle of a connection.
 * The current state can be accessed via ConectorBolsa.getState().
 *
 * @see tech.hellsoft.trading.ConectorBolsa
 */
public enum ConnectionState {
  /**
   * Not connected to the server.
   *
   * <p>This is the initial state and the state after disconnection.
   */
  DISCONNECTED,

  /**
   * Currently establishing a WebSocket connection.
   *
   * <p>The client is in the process of connecting to the server.
   */
  CONNECTING,

  /**
   * WebSocket connection established but not yet authenticated.
   *
   * <p>The connection is active but login authentication has not completed.
   */
  CONNECTED,

  /**
   * Successfully authenticated and ready for trading operations.
   *
   * <p>This is the normal operational state where orders and other messages can be sent to the
   * server.
   */
  AUTHENTICATED,

  /**
   * Attempting to reconnect after connection loss.
   *
   * <p>Automatic reconnection is in progress, using exponential backoff.
   */
  RECONNECTING,

  /**
   * Connection permanently closed.
   *
   * <p>The connection has been intentionally closed and cannot be reused. A new ConectorBolsa
   * instance is required to reconnect.
   */
  CLOSED
}
