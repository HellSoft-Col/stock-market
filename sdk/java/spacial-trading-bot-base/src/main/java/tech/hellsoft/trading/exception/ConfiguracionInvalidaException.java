package tech.hellsoft.trading.exception;

/**
 * Exception thrown when configuration is invalid or missing required fields.
 * Used for errors in loading or validating configuration files.
 */
public class ConfiguracionInvalidaException extends Exception {

  public ConfiguracionInvalidaException(String message) {
    super(message);
  }

  public ConfiguracionInvalidaException(String message, Throwable cause) {
    super(message, cause);
  }
}
