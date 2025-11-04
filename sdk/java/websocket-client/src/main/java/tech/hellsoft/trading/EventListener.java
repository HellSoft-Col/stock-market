package tech.hellsoft.trading;

import tech.hellsoft.trading.dto.server.BalanceUpdateMessage;
import tech.hellsoft.trading.dto.server.BroadcastNotificationMessage;
import tech.hellsoft.trading.dto.server.ErrorMessage;
import tech.hellsoft.trading.dto.server.EventDeltaMessage;
import tech.hellsoft.trading.dto.server.FillMessage;
import tech.hellsoft.trading.dto.server.InventoryUpdateMessage;
import tech.hellsoft.trading.dto.server.LoginOKMessage;
import tech.hellsoft.trading.dto.server.OfferMessage;
import tech.hellsoft.trading.dto.server.OrderAckMessage;
import tech.hellsoft.trading.dto.server.TickerMessage;

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
