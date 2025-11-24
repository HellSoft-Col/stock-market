package tech.hellsoft.trading.enums;

import java.util.Arrays;

/**
 * Enumeration of all message types supported by the Stock Market WebSocket protocol.
 *
 * <p>Each message type corresponds to a specific type of communication between the client and
 * server. The enum provides JSON serialization support for proper protocol communication.
 *
 * <p>Client-to-server messages: LOGIN, ORDER, CANCEL, PRODUCTION_UPDATE, ACCEPT_OFFER, RESYNC
 *
 * <p>Server-to-client messages: LOGIN_OK, ORDER_ACK, FILL, TICKER, OFFER, INVENTORY_UPDATE,
 * BALANCE_UPDATE, EVENT_DELTA, ERROR, BROADCAST_NOTIFICATION, PONG
 *
 * <p>Protocol messages: PING, PONG (heartbeat)
 */
public enum MessageType {
  /** Client authentication request. */
  LOGIN("LOGIN"),

  /** Server authentication confirmation. */
  LOGIN_OK("LOGIN_OK"),

  /** Client order submission. */
  ORDER("ORDER"),

  /** Server order acknowledgment. */
  ORDER_ACK("ORDER_ACK"),

  /** Server order execution notification. */
  FILL("FILL"),

  /** Server market data update. */
  TICKER("TICKER"),

  /** Server offer notification. */
  OFFER("OFFER"),

  /** Client offer response. */
  ACCEPT_OFFER("ACCEPT_OFFER"),

  /** Client production update. */
  PRODUCTION_UPDATE("PRODUCTION_UPDATE"),

  /** Server inventory change notification. */
  INVENTORY_UPDATE("INVENTORY_UPDATE"),

  /** Server balance change notification. */
  BALANCE_UPDATE("BALANCE_UPDATE"),

  /** Client resynchronization request. */
  RESYNC("RESYNC"),

  /** Server market event notification. */
  EVENT_DELTA("EVENT_DELTA"),

  /** Client order cancellation request. */
  CANCEL("CANCEL"),

  /** Server error notification. */
  ERROR("ERROR"),

  /** Server broadcast notification to all participants. */
  BROADCAST_NOTIFICATION("BROADCAST_NOTIFICATION"),

  /** Client heartbeat ping. */
  PING("PING"),

  /** Server heartbeat pong response. */
  PONG("PONG"),

  /** Performance Reporte */
  GLOBAL_PERFORMANCE_REPORT("GLOBAL_PERFORMANCE_REPORT");

  private final String value;

  MessageType(String value) {
    this.value = value;
  }

  /**
   * Gets the JSON string value for this message type.
   *
   * @return the JSON string value
   */
  public String getValue() {
    return value;
  }

  /**
   * Creates a MessageType from its JSON string value.
   *
   * @param value the JSON string value
   * @return the corresponding MessageType
   * @throws IllegalArgumentException if the value is unknown
   */
  public static MessageType fromJson(String value) {
    return Arrays.stream(values())
        .filter(type -> type.value.equals(value))
        .findFirst()
        .orElseThrow(() -> new IllegalArgumentException("Unknown message type: " + value));
  }
}
