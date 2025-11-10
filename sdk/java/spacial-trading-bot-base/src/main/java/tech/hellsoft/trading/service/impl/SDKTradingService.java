package tech.hellsoft.trading.service.impl;

import lombok.Getter;
import tech.hellsoft.trading.config.Configuration;
import tech.hellsoft.trading.exception.ConfiguracionInvalidaException;
import tech.hellsoft.trading.exception.TradingException;
import tech.hellsoft.trading.service.TradingService;
import tech.hellsoft.trading.service.UIService;
import tech.hellsoft.trading.ConectorBolsa;
import tech.hellsoft.trading.EventListener;
import tech.hellsoft.trading.dto.server.LoginOKMessage;
import tech.hellsoft.trading.dto.server.ErrorMessage;
import tech.hellsoft.trading.dto.server.FillMessage;
import tech.hellsoft.trading.dto.server.TickerMessage;
import tech.hellsoft.trading.dto.server.OfferMessage;
import tech.hellsoft.trading.dto.server.OrderAckMessage;
import tech.hellsoft.trading.dto.server.InventoryUpdateMessage;
import tech.hellsoft.trading.dto.server.BalanceUpdateMessage;
import tech.hellsoft.trading.dto.server.EventDeltaMessage;
import tech.hellsoft.trading.dto.server.BroadcastNotificationMessage;

public class SDKTradingService implements TradingService {

  private final UIService uiService;
  private ConectorBolsa connector;
  private TradingEventListener eventListener;
  private boolean running = false;

  public SDKTradingService(UIService service) {
    this.uiService = service;
  }

  @Override
  public void start(Configuration config) throws ConfiguracionInvalidaException, TradingException {
    if (config == null) {
      throw new ConfiguracionInvalidaException("Configuration cannot be null");
    }

    try {
      uiService.printStatus("🔌", "Connecting to trading server...");

      connector = new ConectorBolsa();
      eventListener = new TradingEventListener(uiService);
      connector.addListener(eventListener);

      connector.conectar(config.host(), config.apiKey());

      uiService.printStatus("⏳", "Waiting for login response...");

      Thread.sleep(3000);

      if (eventListener.isLoggedIn()) {
        running = true;
        uiService.printSuccess("Login successful! Ready for trading operations.");
      } else {
        throw new TradingException("Login failed. Check API key and connection.");
      }

    } catch (InterruptedException e) {
      Thread.currentThread().interrupt();
      throw new TradingException("Operation interrupted", e);
    } catch (Exception e) {
      throw new TradingException("Failed to start trading service: " + e.getMessage(), e);
    }
  }

  @Override
  public void stop() {
    if (running && connector != null) {
      uiService.printStatus("🛑", "Stopping trading service...");
      running = false;
      uiService.printSuccess("Trading service stopped.");
    }
  }

  @Override
  public boolean isRunning() {
    return running;
  }

  private static class TradingEventListener implements EventListener {
    private final UIService listenerUiService;
    @Getter
    private boolean loggedIn = false;

    TradingEventListener(UIService uiService) {
      this.listenerUiService = uiService;
    }

    @Override
    public void onLoginOk(LoginOKMessage loginOk) {
      if (loginOk != null) {
        loggedIn = true;
        listenerUiService.printSuccess("Login OK received!");
        listenerUiService.printStatus("   Team:", loginOk.getTeam());
        listenerUiService.printStatus("   Species:", loginOk.getSpecies());
        listenerUiService.printStatus("   Initial Balance:", String.valueOf(loginOk.getInitialBalance()));
        listenerUiService.printStatus("   Current Balance:", String.valueOf(loginOk.getCurrentBalance()));
      }
    }

    @Override
    public void onError(ErrorMessage error) {
      if (error != null) {
        listenerUiService.printError("Server Error [" + error.getCode() + "]: " + error.getReason());
      }
    }

    @Override
    public void onFill(FillMessage fill) {
      listenerUiService.printStatus("📈",
          "Fill received: " + fill.getProduct() + " " + fill.getFillQty() + "@" + fill.getFillPrice());
    }

    @Override
    public void onTicker(TickerMessage ticker) {
      listenerUiService.printStatus("📊",
          "Ticker update: " + ticker.getProduct() + " Bid:" + ticker.getBestBid() + " Ask:" + ticker.getBestAsk());
    }

    @Override
    public void onOffer(OfferMessage offer) {
      listenerUiService.printStatus("💰", "Offer received: " + offer);
    }

    @Override
    public void onOrderAck(OrderAckMessage orderAck) {
      listenerUiService.printStatus("📋", "Order acknowledgment: " + orderAck);
    }

    @Override
    public void onInventoryUpdate(InventoryUpdateMessage inventoryUpdate) {
      listenerUiService.printStatus("📦", "Inventory update: " + inventoryUpdate);
    }

    @Override
    public void onBalanceUpdate(BalanceUpdateMessage balanceUpdate) {
      listenerUiService.printStatus("💳", "Balance update: " + balanceUpdate);
    }

    @Override
    public void onEventDelta(EventDeltaMessage eventDelta) {
      listenerUiService.printStatus("🔄", "Event delta: " + eventDelta);
    }

    @Override
    public void onBroadcast(BroadcastNotificationMessage broadcast) {
      listenerUiService.printStatus("📢", "Broadcast: " + broadcast);
    }

    @Override
    public void onConnectionLost(Throwable throwable) {
      listenerUiService.printError("Connection lost: " + throwable.getMessage());
    }

  }
}
