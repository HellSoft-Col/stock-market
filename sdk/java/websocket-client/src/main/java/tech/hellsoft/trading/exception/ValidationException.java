package tech.hellsoft.trading.exception;

/**
 * Exception thrown when client-side validation fails.
 *
 * <p>This exception is thrown when message validation fails before sending to the server. This
 * includes missing required fields, invalid values, or other validation rule violations.
 *
 * <p>Unlike server-side errors which arrive via {@link
 * tech.hellsoft.trading.EventListener#onError(tech.hellsoft.trading.dto.server.ErrorMessage)},
 * validation exceptions are thrown immediately from the sending method.
 *
 * @see tech.hellsoft.trading.ConectorBolsa
 * @see tech.hellsoft.trading.dto.client.OrderMessage
 * @see tech.hellsoft.trading.ConectorBolsa#enviarCancelacion(String) where validation may fail
 */
public class ValidationException extends RuntimeException {
  private static final long serialVersionUID = 1L;

  /**
   * Creates a new validation exception.
   *
   * @param message the validation error message
   */
  public ValidationException(String message) {
    super(message);
  }

  /**
   * Creates a new validation exception with a cause.
   *
   * @param message the validation error message
   * @param cause the underlying cause
   */
  public ValidationException(String message, Throwable cause) {
    super(message, cause);
  }
}
