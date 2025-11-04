package tech.hellsoft.trading.enums;

import java.util.Arrays;

public enum MessageType {
  LOGIN("LOGIN"),
  LOGIN_OK("LOGIN_OK"),
  ORDER("ORDER"),
  ORDER_ACK("ORDER_ACK"),
  FILL("FILL"),
  TICKER("TICKER"),
  OFFER("OFFER"),
  ACCEPT_OFFER("ACCEPT_OFFER"),
  PRODUCTION_UPDATE("PRODUCTION_UPDATE"),
  INVENTORY_UPDATE("INVENTORY_UPDATE"),
  BALANCE_UPDATE("BALANCE_UPDATE"),
  RESYNC("RESYNC"),
  EVENT_DELTA("EVENT_DELTA"),
  CANCEL("CANCEL"),
  ERROR("ERROR"),
  BROADCAST_NOTIFICATION("BROADCAST_NOTIFICATION"),
  PING("PING"),
  PONG("PONG");

  private final String value;

  MessageType(String value) {
    this.value = value;
  }

  public String getValue() {
    return value;
  }

  public static MessageType fromJson(String value) {
    return Arrays.stream(values())
        .filter(type -> type.value.equals(value))
        .findFirst()
        .orElseThrow(() -> new IllegalArgumentException("Unknown message type: " + value));
  }
}
