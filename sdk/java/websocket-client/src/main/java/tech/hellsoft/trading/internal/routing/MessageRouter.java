package tech.hellsoft.trading.internal.routing;

import com.google.gson.JsonObject;

import tech.hellsoft.trading.dto.server.*;
import tech.hellsoft.trading.enums.MessageType;
import tech.hellsoft.trading.internal.serialization.JsonSerializer;

import lombok.extern.slf4j.Slf4j;

@Slf4j
public class MessageRouter {

  public void routeMessage(String json, MessageHandlers handlers) {
    if (json == null || json.isBlank()) {
      log.warn("Received null or blank message");
      return;
    }

    try {
      JsonObject obj = JsonSerializer.parseObject(json);

      if (!obj.has("type")) {
        log.warn("Message missing 'type' field: {}", json);
        return;
      }

      String typeStr = obj.get("type").getAsString();
      MessageType type = MessageType.fromJson(typeStr);

      routeByType(type, json, handlers);

    } catch (Exception e) {
      log.error("Failed to route message: {}", json, e);
    }
  }

  private void routeByType(MessageType type, String json, MessageHandlers handlers) {
    switch (type) {
      case LOGIN_OK -> {
        LoginOKMessage msg = JsonSerializer.fromJson(json, LoginOKMessage.class);
        handlers.onLoginOk(msg);
      }
      case FILL -> {
        FillMessage msg = JsonSerializer.fromJson(json, FillMessage.class);
        handlers.onFill(msg);
      }
      case TICKER -> {
        TickerMessage msg = JsonSerializer.fromJson(json, TickerMessage.class);
        handlers.onTicker(msg);
      }
      case OFFER -> {
        OfferMessage msg = JsonSerializer.fromJson(json, OfferMessage.class);
        handlers.onOffer(msg);
      }
      case ERROR -> {
        ErrorMessage msg = JsonSerializer.fromJson(json, ErrorMessage.class);
        handlers.onError(msg);
      }
      case ORDER_ACK -> {
        OrderAckMessage msg = JsonSerializer.fromJson(json, OrderAckMessage.class);
        handlers.onOrderAck(msg);
      }
      case INVENTORY_UPDATE -> {
        InventoryUpdateMessage msg = JsonSerializer.fromJson(json, InventoryUpdateMessage.class);
        handlers.onInventoryUpdate(msg);
      }
      case BALANCE_UPDATE -> {
        BalanceUpdateMessage msg = JsonSerializer.fromJson(json, BalanceUpdateMessage.class);
        handlers.onBalanceUpdate(msg);
      }
      case EVENT_DELTA -> {
        EventDeltaMessage msg = JsonSerializer.fromJson(json, EventDeltaMessage.class);
        handlers.onEventDelta(msg);
      }
      case BROADCAST_NOTIFICATION -> {
        BroadcastNotificationMessage msg =
            JsonSerializer.fromJson(json, BroadcastNotificationMessage.class);
        handlers.onBroadcast(msg);
      }
      case PONG -> {
        PongMessage msg = JsonSerializer.fromJson(json, PongMessage.class);
        handlers.onPong(msg);
      }
      default -> log.warn("Unhandled message type: {}", type);
    }
  }

  public interface MessageHandlers {
    void onLoginOk(LoginOKMessage message);

    void onFill(FillMessage message);

    void onTicker(TickerMessage message);

    void onOffer(OfferMessage message);

    void onError(ErrorMessage message);

    void onOrderAck(OrderAckMessage message);

    void onInventoryUpdate(InventoryUpdateMessage message);

    void onBalanceUpdate(BalanceUpdateMessage message);

    void onEventDelta(EventDeltaMessage message);

    void onBroadcast(BroadcastNotificationMessage message);

    void onPong(PongMessage message);
  }
}
