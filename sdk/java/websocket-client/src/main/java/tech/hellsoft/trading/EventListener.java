package tech.hellsoft.trading;

import tech.hellsoft.trading.dto.server.*;

public interface EventListener {
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

  void onConnectionLost(Throwable error);
}
