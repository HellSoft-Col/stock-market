package tech.hellsoft.trading.exception;

import lombok.Getter;

/**
 * Exception thrown when WebSocket connection to server fails.
 *
 * <p>This exception is thrown by ConectorBolsa.conectar() when the connection cannot be established
 * due to network issues, server unavailability, or authentication failures.
 *
 * @see tech.hellsoft.trading.ConectorBolsa
 */
@Getter
public class ConexionFallidaException extends Exception {
  private static final long serialVersionUID = 1L;

  /** The hostname that connection was attempted to. */
  private final String host;

  /** The port number that connection was attempted to. */
  private final int port;

  /**
   * Creates a new connection failure exception.
   *
   * @param message the error message
   * @param host the target hostname
   * @param port the target port
   */
  public ConexionFallidaException(String message, String host, int port) {
    super(message);
    this.host = host;
    this.port = port;
  }

  /**
   * Creates a new connection failure exception with a cause.
   *
   * @param message the error message
   * @param host the target hostname
   * @param port the target port
   * @param cause the underlying cause of the failure
   */
  public ConexionFallidaException(String message, String host, int port, Throwable cause) {
    super(message, cause);
    this.host = host;
    this.port = port;
  }
}
