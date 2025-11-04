package tech.hellsoft.trading.config;

import java.time.Duration;

import lombok.Builder;
import lombok.Data;
import lombok.extern.slf4j.Slf4j;

/**
 * Configuration settings for the Stock Market WebSocket client.
 *
 * <p>This class provides comprehensive configuration options for connection management,
 * reconnection behavior, message processing, and synchronization.
 *
 * <p>Example usage:
 *
 * <pre>{@code
 * ConectorConfig config = ConectorConfig.builder()
 *     .heartbeatInterval(Duration.ofSeconds(15))
 *     .connectionTimeout(Duration.ofSeconds(5))
 *     .autoReconnect(true)
 *     .maxReconnectAttempts(10)
 *     .build();
 *
 * ConectorBolsa connector = new ConectorBolsa(config);
 * }</pre>
 *
 * @see tech.hellsoft.trading.ConectorBolsa
 */
@Data
@Builder
@Slf4j
public class ConectorConfig {

  /**
   * The interval between heartbeat messages to keep the connection alive.
   *
   * <p>Default: 30 seconds
   */
  @Builder.Default private Duration heartbeatInterval = Duration.ofSeconds(30);

  /**
   * The timeout for establishing the WebSocket connection.
   *
   * <p>Default: 10 seconds
   */
  @Builder.Default private Duration connectionTimeout = Duration.ofSeconds(10);

  /**
   * Whether to automatically reconnect when the connection is lost.
   *
   * <p>Default: true
   */
  @Builder.Default private boolean autoReconnect = true;

  /**
   * Maximum number of reconnection attempts.
   *
   * <p>Set to -1 for unlimited attempts. Default: 5
   */
  @Builder.Default private int maxReconnectAttempts = 5;

  /**
   * Initial delay before the first reconnection attempt.
   *
   * <p>Default: 1 second
   */
  @Builder.Default private Duration reconnectInitialDelay = Duration.ofSeconds(1);

  /**
   * Maximum delay between reconnection attempts.
   *
   * <p>Default: 30 seconds
   */
  @Builder.Default private Duration reconnectMaxDelay = Duration.ofSeconds(30);

  /**
   * Multiplier for exponential backoff in reconnection delays.
   *
   * <p>Default: 2.0 (doubling each attempt)
   */
  @Builder.Default private double reconnectBackoffMultiplier = 2.0;

  /**
   * Whether to enable message sequencing to ensure ordered processing.
   *
   * <p>Default: true
   */
  @Builder.Default private boolean enableMessageSequencing = true;

  /**
   * Timeout for message sequencing operations.
   *
   * <p>Default: 30 seconds
   */
  @Builder.Default private Duration messageSequencingTimeout = Duration.ofSeconds(30);

  /**
   * Whether to enable state locking for thread-safe operations.
   *
   * <p>Default: true
   */
  @Builder.Default private boolean enableStateLocking = true;

  /**
   * Timeout for acquiring state locks.
   *
   * <p>Default: 5 seconds
   */
  @Builder.Default private Duration stateLockTimeout = Duration.ofSeconds(5);

  /**
   * Whether to automatically resynchronize data after reconnection.
   *
   * <p>Default: true
   */
  @Builder.Default private boolean autoResyncOnReconnect = true;

  /**
   * Lookback period for data resynchronization after reconnection.
   *
   * <p>Default: 5 minutes
   */
  @Builder.Default private Duration resyncLookback = Duration.ofMinutes(5);

  /**
   * Creates a configuration with default settings.
   *
   * <p>Equivalent to {@code ConectorConfig.builder().build()}
   *
   * @return a configuration with all default values
   */
  public static ConectorConfig defaultConfig() {
    return ConectorConfig.builder().build();
  }

  /**
   * Validates all configuration settings.
   *
   * <p>This method is called automatically during ConectorBolsa construction to ensure all settings
   * are valid. Throws IllegalArgumentException for any invalid configuration values.
   *
   * @throws IllegalArgumentException if any configuration value is invalid
   */
  public void validate() {
    if (heartbeatInterval.isNegative() || heartbeatInterval.isZero()) {
      throw new IllegalArgumentException("Heartbeat interval must be positive");
    }

    if (connectionTimeout.isNegative() || connectionTimeout.isZero()) {
      throw new IllegalArgumentException("Connection timeout must be positive");
    }

    if (maxReconnectAttempts < -1 || maxReconnectAttempts == 0) {
      throw new IllegalArgumentException("Max reconnect attempts must be -1 or positive");
    }

    if (reconnectInitialDelay.isNegative()) {
      throw new IllegalArgumentException("Reconnect initial delay cannot be negative");
    }

    if (reconnectMaxDelay.compareTo(reconnectInitialDelay) < 0) {
      throw new IllegalArgumentException("Reconnect max delay must be >= initial delay");
    }

    if (reconnectBackoffMultiplier < 1.0) {
      throw new IllegalArgumentException("Reconnect backoff multiplier must be >= 1.0");
    }

    if (stateLockTimeout.isNegative() || stateLockTimeout.isZero()) {
      throw new IllegalArgumentException("State lock timeout must be positive");
    }

    if (messageSequencingTimeout.isNegative() || messageSequencingTimeout.isZero()) {
      throw new IllegalArgumentException("Message sequencing timeout must be positive");
    }

    if (resyncLookback.isNegative()) {
      throw new IllegalArgumentException("Resync lookback cannot be negative");
    }

    log.debug("Configuration validated successfully");
  }
}
