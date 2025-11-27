package com.myteam;

import tech.hellsoft.trading.ConectorBolsa;
import tech.hellsoft.trading.EventListener;
import tech.hellsoft.trading.dto.client.OrderMessage;
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
import tech.hellsoft.trading.enums.MessageType;
import tech.hellsoft.trading.enums.OrderMode;
import tech.hellsoft.trading.enums.OrderSide;
import tech.hellsoft.trading.enums.Product;
import tech.hellsoft.trading.util.OrderIdGenerator;

public class TradingBot implements EventListener {

    private ConectorBolsa connector;
    private String serverHost = "localhost";
    private int serverPort = 8080;
    private String token = "YOUR_TOKEN_HERE";
    
    // Thread-safe order ID generator - prevents duplicate order IDs
    private final OrderIdGenerator orderIdGen = new OrderIdGenerator("MYTEAM");

    public static void main(String[] args) throws Exception {
        TradingBot bot = new TradingBot();
        bot.start();
    }

    public void start() throws Exception {
        System.out.println("🤖 Starting Trading Bot...");
        
        connector = new ConectorBolsa();
        connector.addListener(this);
        
        System.out.println("📡 Connecting to " + serverHost + ":" + serverPort);
        connector.conectar(serverHost, serverPort, token);
        
        System.out.println("⏳ Waiting for login confirmation...");
    }

    @Override
    public void onLoginOk(LoginOKMessage message) {
        System.out.println("\n✅ LOGIN SUCCESSFUL!");
        System.out.println("👥 Team: " + message.getTeam());
        System.out.println("💰 Cash Balance: $" + message.getCash());
        System.out.println("📦 Inventory: " + message.getInventory());
        System.out.println("📊 Recipes: " + message.getRecipes());
        System.out.println();
        
        // TODO: Implement your trading strategy here
    }

    @Override
    public void onFill(FillMessage message) {
        System.out.println("✅ ORDER FILLED: " + message.getClOrdID());
        // TODO: Handle order fills
    }

    @Override
    public void onTicker(TickerMessage message) {
        // TODO: Analyze market prices
    }

    @Override
    public void onOffer(OfferMessage message) {
        System.out.println("🎯 OFFER RECEIVED: " + message.getOfferID());
        // TODO: Decide whether to accept
    }

    @Override
    public void onError(ErrorMessage message) {
        System.err.println("❌ ERROR: " + message.getError());
        // TODO: Handle errors
    }

    @Override
    public void onOrderAck(OrderAckMessage message) {
        System.out.println("✓ Order Acknowledged: " + message.getClOrdID());
        // TODO: Track order status
    }

    @Override
    public void onInventoryUpdate(InventoryUpdateMessage message) {
        System.out.println("📦 INVENTORY UPDATE: " + message.getInventory());
        // TODO: Update inventory tracking
    }

    @Override
    public void onBalanceUpdate(BalanceUpdateMessage message) {
        System.out.println("💰 BALANCE UPDATE: $" + message.getCash());
        // TODO: Track cash balance
    }

    @Override
    public void onEventDelta(EventDeltaMessage message) {
        System.out.println("📅 EVENT: " + message.getEvent());
        // TODO: Handle game events
    }

    @Override
    public void onBroadcast(BroadcastNotificationMessage message) {
        System.out.println("📢 BROADCAST: " + message.getMessage());
        // TODO: Process broadcasts
    }

    @Override
    public void onConnectionLost(Throwable error) {
        System.err.println("❌ CONNECTION LOST: " + error.getMessage());
        // TODO: Implement reconnection logic
    }

    private void placeExampleOrder() {
        OrderMessage order = OrderMessage.builder()
            .type(MessageType.ORDER)
            .clOrdID(orderIdGen.next())  // Thread-safe unique ID
            .product(Product.USD)
            .side(OrderSide.BUY)
            .mode(OrderMode.LIMIT)
            .quantity(100)
            .price(1.0)
            .build();
            
        connector.enviarOrden(order);
        System.out.println("📤 Order sent: " + order.getClOrdID());
    }
}
