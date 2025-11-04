# Student Guide: Using the Trading SDK

This guide will help you set up and use the Stock Market Trading SDK in your Java project.

## Prerequisites

- Java 25 or higher installed
- Gradle or Maven (Gradle recommended)
- IntelliJ IDEA or any Java IDE
- GitHub account

## Quick Setup (Recommended for Students)

### Method 1: Download Pre-built JAR (Easiest)

This is the simplest method and doesn't require GitHub authentication.

1. **Download the JAR**:
   - Go to [GitHub Actions](../../../actions/workflows/publish-java-sdk.yml)
   - Click on the most recent successful workflow run (green checkmark ✓)
   - Scroll down to "Artifacts" section
   - Download `package` (it will download as `package.zip`)
   - Extract the ZIP file to get `websocket-client-1.0.0-SNAPSHOT.jar`

2. **Create your project structure**:
   ```
   my-trading-bot/
   ├── libs/
   │   └── websocket-client-1.0.0-SNAPSHOT.jar
   ├── src/
   │   └── main/
   │       └── java/
   │           └── MyTradingBot.java
   └── build.gradle.kts
   ```

3. **Create `build.gradle.kts`**:
   ```kotlin
   plugins {
       java
       application
   }
   
   group = "com.myteam"
   version = "1.0.0"
   
   repositories {
       mavenCentral()
   }
   
   dependencies {
       // Add the SDK JAR
       implementation(files("libs/websocket-client-1.0.0-SNAPSHOT.jar"))
       
       // Required dependencies
       implementation("com.google.code.gson:gson:2.10.1")
       implementation("org.projectlombok:lombok:1.18.30")
       annotationProcessor("org.projectlombok:lombok:1.18.30")
       implementation("org.slf4j:slf4j-api:2.0.9")
       implementation("org.slf4j:slf4j-simple:2.0.9")
   }
   
   java {
       toolchain {
           languageVersion.set(JavaLanguageVersion.of(25))
       }
   }
   
   application {
       mainClass.set("MyTradingBot")
   }
   ```

4. **Create your trading bot** (`src/main/java/MyTradingBot.java`):
   ```java
   import tech.hellsoft.trading.ConectorBolsa;
   import tech.hellsoft.trading.EventListener;
   import tech.hellsoft.trading.dto.client.OrderMessage;
   import tech.hellsoft.trading.dto.server.*;
   import tech.hellsoft.trading.enums.MessageType;
   import tech.hellsoft.trading.enums.OrderMode;
   import tech.hellsoft.trading.enums.OrderSide;
   import tech.hellsoft.trading.enums.Product;
   
   public class MyTradingBot implements EventListener {
       private ConectorBolsa connector;
       
       public static void main(String[] args) throws Exception {
           MyTradingBot bot = new MyTradingBot();
           bot.start();
       }
       
       public void start() throws Exception {
           connector = new ConectorBolsa();
           connector.addListener(this);
           
           // Connect to server (replace with your details)
           connector.conectar("localhost", 8080, "your-token-here");
           
           System.out.println("Bot started! Waiting for login...");
       }
       
       @Override
       public void onLoginOk(LoginOKMessage message) {
           System.out.println("✅ Logged in! Team: " + message.getTeam());
           System.out.println("💰 Balance: $" + message.getCash());
           
           // Example: Place your first order
           placeOrder();
       }
       
       private void placeOrder() {
           OrderMessage order = OrderMessage.builder()
               .type(MessageType.ORDER)
               .clOrdID("ORDER-001")
               .product(Product.USD)
               .side(OrderSide.BUY)
               .mode(OrderMode.LIMIT)
               .quantity(100)
               .price(1.0)
               .build();
               
           connector.enviarOrden(order);
           System.out.println("📤 Order sent!");
       }
       
       @Override
       public void onFill(FillMessage message) {
           System.out.println("✅ Order filled! " + message);
       }
       
       @Override
       public void onTicker(TickerMessage message) {
           System.out.println("📊 Market update: " + message.getProduct() + 
                            " Bid: " + message.getBid() + 
                            " Ask: " + message.getAsk());
       }
       
       @Override
       public void onOffer(OfferMessage message) {
           System.out.println("🎯 Offer received: " + message);
       }
       
       @Override
       public void onError(ErrorMessage message) {
           System.err.println("❌ Error: " + message.getError());
       }
       
       @Override
       public void onOrderAck(OrderAckMessage message) {
           System.out.println("✓ Order acknowledged: " + message.getClOrdID());
       }
       
       @Override
       public void onInventoryUpdate(InventoryUpdateMessage message) {
           System.out.println("📦 Inventory: " + message);
       }
       
       @Override
       public void onBalanceUpdate(BalanceUpdateMessage message) {
           System.out.println("💰 Balance: $" + message.getCash());
       }
       
       @Override
       public void onEventDelta(EventDeltaMessage message) {
           System.out.println("📅 Event: " + message);
       }
       
       @Override
       public void onBroadcast(BroadcastNotificationMessage message) {
           System.out.println("📢 Broadcast: " + message.getMessage());
       }
       
       @Override
       public void onConnectionLost(Throwable error) {
           System.err.println("❌ Connection lost: " + error.getMessage());
       }
   }
   ```

5. **Run your bot**:
   ```bash
   ./gradlew run
   ```

---

### Method 2: GitHub Packages (Advanced)

If you want automatic updates when the SDK is published, use GitHub Packages.

1. **Create a GitHub Personal Access Token**:
   - Go to: https://github.com/settings/tokens
   - Click "Generate new token (classic)"
   - Give it a name like "Maven Packages"
   - Select scope: ✓ `read:packages`
   - Click "Generate token"
   - **Copy the token** (you won't see it again!)

2. **Configure Gradle**:

   Create `gradle.properties` in your home directory (`~/.gradle/gradle.properties`):
   ```properties
   gpr.user=YOUR_GITHUB_USERNAME
   gpr.token=ghp_xxxxxxxxxxxx
   ```

   Add to your `build.gradle.kts`:
   ```kotlin
   repositories {
       mavenCentral()
       maven {
           url = uri("https://maven.pkg.github.com/HellSoft-Col/stock-market")
           credentials {
               username = project.findProperty("gpr.user") as String? ?: System.getenv("GITHUB_ACTOR")
               password = project.findProperty("gpr.token") as String? ?: System.getenv("GITHUB_TOKEN")
           }
       }
   }
   
   dependencies {
       implementation("tech.hellsoft.trading:websocket-client:1.0.0-SNAPSHOT")
   }
   ```

3. **Sync and run**:
   ```bash
   ./gradlew build
   ./gradlew run
   ```

---

## IntelliJ IDEA Setup

1. **Open IntelliJ IDEA**
2. **File → New → Project from Existing Sources**
3. Select your project directory
4. Choose "Gradle" as the build tool
5. Click "Finish"
6. Wait for Gradle to sync
7. Run your bot: Right-click on `MyTradingBot.java` → Run

---

## Common Issues

### Issue: "Cannot resolve tech.hellsoft.trading"
**Solution**: Make sure you either:
- Have the JAR file in `libs/` folder (Method 1), OR
- Have configured GitHub Packages credentials correctly (Method 2)

### Issue: "java: error: release version 25 not supported"
**Solution**: Install Java 25:
- Download from: https://jdk.java.net/25/
- Set IntelliJ SDK: File → Project Structure → Project → SDK → Add SDK

### Issue: "Connection refused"
**Solution**: Make sure the trading server is running and accessible at the host/port you specified.

### Issue: "Authentication failed"
**Solution**: Check your token is valid. Ask your instructor for a valid token.

---

## API Reference

### Connecting
```java
ConectorBolsa connector = new ConectorBolsa();
connector.addListener(this);
connector.conectar("localhost", 8080, "your-token");
```

### Placing Orders
```java
OrderMessage order = OrderMessage.builder()
    .type(MessageType.ORDER)
    .clOrdID("unique-id")      // Your order ID
    .product(Product.USD)       // What to trade
    .side(OrderSide.BUY)       // BUY or SELL
    .mode(OrderMode.LIMIT)     // LIMIT or MARKET
    .quantity(100)              // How many
    .price(1.0)                 // At what price (LIMIT only)
    .build();
    
connector.enviarOrden(order);
```

### Accepting Offers
```java
AcceptOfferMessage response = AcceptOfferMessage.builder()
    .type(MessageType.ACCEPT_OFFER)
    .offerID(offer.getOfferID())
    .build();
    
connector.enviarRespuestaOferta(response);
```

### Canceling Orders
```java
connector.cancelarOrden("ORDER-001");
```

### Updating Production
```java
ProductionUpdateMessage update = ProductionUpdateMessage.builder()
    .type(MessageType.PRODUCTION_UPDATE)
    .recipe(RecipeType.R1)
    .enabled(true)
    .build();
    
connector.enviarActualizacionProduccion(update);
```

### Disconnecting
```java
connector.desconectar();
```

---

## Support

- Check the [main README](websocket-client/README.md) for detailed documentation
- Ask your instructor for help
- Review example code in the SDK repository

---

## Tips for Success

1. **Always implement all EventListener methods** - Even if empty, you must implement them all
2. **Use unique order IDs** - Make sure each order has a unique `clOrdID`
3. **Handle errors gracefully** - Check `onError()` for error messages
4. **Test locally first** - Connect to localhost before production server
5. **Read the market data** - Use `onTicker()` to understand current prices
6. **Don't spam orders** - The server has rate limits
7. **Keep your token secret** - Don't commit it to Git!

Good luck with your trading bot! 🚀
