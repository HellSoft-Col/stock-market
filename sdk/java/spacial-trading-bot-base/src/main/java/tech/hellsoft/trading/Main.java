package tech.hellsoft.trading;

import java.util.Scanner;
import tech.hellsoft.trading.config.Configuration;
import tech.hellsoft.trading.util.ConfigLoader;
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

/**
 * CLI Trading Bot with interactive menu.
 *
 * Students should implement the TODO methods below to complete the trading bot
 * functionality.
 */
public final class Main {

  private Main() {
  }

  private static boolean running = true;

  public static void main(String[] args) {
    try {
      // 1. Load configuration (apiKey, team, host)
      Configuration config = ConfigLoader.load("src/main/resources/config.json");
      printBanner();
      System.out.println("🚀 Starting Trading Bot for team: " + config.team());
      System.out.println();

      // 2. Create connector and event listener
      ConectorBolsa connector = new ConectorBolsa();
      MyTradingBot bot = new MyTradingBot();
      connector.addListener(bot);

      // 3. Connect to server
      System.out.println("🔌 Connecting to: " + config.host());
      connector.conectar(config.host(), config.apiKey());
      System.out.println("✅ Connected! Waiting for login...");
      System.out.println();

      // 4. Interactive CLI menu
      runInteractiveCLI(connector, bot);

    } catch (Exception e) {
      System.err.println("❌ Error: " + e.getMessage());
      e.printStackTrace();
      System.exit(1);
    }
  }

  private static void printBanner() {
    System.out.println("╔══════════════════════════════════════════════════════════╗");
    System.out.println("║  🥑 Bolsa Interestelar de Aguacates Andorianos 🥑      ║");
    System.out.println("║  Trading Bot CLI - Java 25 Edition                      ║");
    System.out.println("╚══════════════════════════════════════════════════════════╝");
    System.out.println();
  }

  private static void runInteractiveCLI(ConectorBolsa connector, MyTradingBot bot) {
    Scanner scanner = new Scanner(System.in);

    while (running) {
      printMenu();
      System.out.print("\n> ");

      if (!scanner.hasNextLine()) {
        break;
      }

      String input = scanner.nextLine().trim();

      // Guard clause - skip empty input
      if (input.isEmpty()) {
        continue;
      }

      String[] parts = input.split("\\s+");
      String command = parts[0].toLowerCase();

      handleCommand(command, parts, connector, bot);
    }

    scanner.close();
    System.out.println("\n👋 Cerrando Trading Bot...");
    System.out.println("✅ ¡Hasta luego!");
  }

  private static void printMenu() {
    System.out.println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━");
    System.out.println("📋 COMANDOS DISPONIBLES:");
    System.out.println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━");
    System.out.println("  status              - Ver estado actual (saldo, P&L)");
    System.out.println("  inventario          - Ver productos en inventario");
    System.out.println("  precios             - Ver precios de mercado");
    System.out.println("  comprar <producto> <cantidad> [mensaje]");
    System.out.println("  vender <producto> <cantidad> [mensaje]");
    System.out.println("  producir <producto> <basico|premium>");
    System.out.println("  ofertas             - Ver ofertas pendientes");
    System.out.println("  aceptar <offerId>   - Aceptar una oferta");
    System.out.println("  ayuda               - Mostrar ayuda completa");
    System.out.println("  exit                - Salir del programa");
    System.out.println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━");
  }

  private static void handleCommand(String command, String[] parts, ConectorBolsa connector, MyTradingBot bot) {
    switch (command) {
    case "status" :
      handleStatus(bot);
      break;

    case "inventario" :
      handleInventario(bot);
      break;

    case "precios" :
      handlePrecios(bot);
      break;

    case "comprar" :
      handleComprar(parts, connector, bot);
      break;

    case "vender" :
      handleVender(parts, connector, bot);
      break;

    case "producir" :
      handleProducir(parts, connector, bot);
      break;

    case "ofertas" :
      handleOfertas(bot);
      break;

    case "aceptar" :
      handleAceptarOferta(parts, connector, bot);
      break;

    case "ayuda" :
    case "help" :
      printHelp();
      break;

    case "exit" :
    case "quit" :
    case "salir" :
      running = false;
      break;

    default :
      System.out.println("❌ Comando desconocido: " + command);
      System.out.println("💡 Escribe 'ayuda' para ver todos los comandos");
    }
  }

  // ==================== COMMAND HANDLERS ====================
  // TODO: Students should implement these methods

  private static void handleStatus(MyTradingBot bot) {
    System.out.println("\n📊 ESTADO ACTUAL");
    System.out.println("━━━━━━━━━━━━━━━━━━━━━");
    System.out.println("💰 Saldo: $0.00");
    System.out.println("📦 Valor inventario: $0.00");
    System.out.println("💎 Patrimonio neto: $0.00");
    System.out.println("📈 P&L: +0.00%");
    System.out.println();
    System.out.println("TODO: Implementar cálculo de estado real");
    System.out.println("      - Leer saldo de EstadoCliente");
    System.out.println("      - Calcular valor de inventario");
    System.out.println("      - Calcular P&L%");
  }

  private static void handleInventario(MyTradingBot bot) {
    System.out.println("\n📦 INVENTARIO");
    System.out.println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━");
    System.out.println("(vacío)");
    System.out.println();
    System.out.println("TODO: Implementar listado de inventario");
    System.out.println("      - Obtener Map<String, Integer> de EstadoCliente");
    System.out.println("      - Para cada producto: mostrar cantidad y valor");
  }

  private static void handlePrecios(MyTradingBot bot) {
    System.out.println("\n💹 PRECIOS DE MERCADO");
    System.out.println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━");
    System.out.println("(esperando tickers...)");
    System.out.println();
    System.out.println("TODO: Implementar listado de precios");
    System.out.println("      - Obtener Map<String, Double> de EstadoCliente");
    System.out.println("      - Mostrar bid, ask, mid de cada producto");
  }

  private static void handleComprar(String[] parts, ConectorBolsa connector, MyTradingBot bot) {
    if (parts.length < 3) {
      System.out.println("❌ Uso: comprar <producto> <cantidad> [mensaje]");
      return;
    }

    String producto = parts[1];
    int cantidad = Integer.parseInt(parts[2]);
    String mensaje = parts.length > 3
        ? String.join(" ", java.util.Arrays.copyOfRange(parts, 3, parts.length))
        : "Orden de compra";

    System.out.println("\n📤 Enviando orden de compra:");
    System.out.println("   Producto: " + producto);
    System.out.println("   Cantidad: " + cantidad);
    System.out.println("   Mensaje: " + mensaje);
    System.out.println();
    System.out.println("TODO: Implementar lógica de compra");
    System.out.println("      1. Validar saldo suficiente");
    System.out.println("      2. Crear objeto Orden");
    System.out.println("      3. Llamar connector.enviarOrden()");
  }

  private static void handleVender(String[] parts, ConectorBolsa connector, MyTradingBot bot) {
    if (parts.length < 3) {
      System.out.println("❌ Uso: vender <producto> <cantidad> [mensaje]");
      return;
    }

    String producto = parts[1];
    int cantidad = Integer.parseInt(parts[2]);
    String mensaje = parts.length > 3
        ? String.join(" ", java.util.Arrays.copyOfRange(parts, 3, parts.length))
        : "Orden de venta";

    System.out.println("\n📤 Enviando orden de venta:");
    System.out.println("   Producto: " + producto);
    System.out.println("   Cantidad: " + cantidad);
    System.out.println("   Mensaje: " + mensaje);
    System.out.println();
    System.out.println("TODO: Implementar lógica de venta");
    System.out.println("      1. Validar inventario suficiente");
    System.out.println("      2. Crear objeto Orden");
    System.out.println("      3. Llamar connector.enviarOrden()");
  }

  private static void handleProducir(String[] parts, ConectorBolsa connector, MyTradingBot bot) {
    if (parts.length < 3) {
      System.out.println("❌ Uso: producir <producto> <basico|premium>");
      return;
    }

    String producto = parts[1];
    String tipo = parts[2].toLowerCase();
    boolean premium = tipo.equals("premium");

    System.out.println("\n🏭 Produciendo " + producto + " (" + tipo + "):");
    System.out.println();
    System.out.println("TODO: Implementar lógica de producción");
    System.out.println("      1. Validar producto autorizado");
    System.out.println("      2. Si premium: validar ingredientes");
    System.out.println("      3. Calcular unidades (algoritmo recursivo)");
    System.out.println("      4. Actualizar inventario");
    System.out.println("      5. Llamar connector.enviarProduccion()");
  }

  private static void handleOfertas(MyTradingBot bot) {
    System.out.println("\n📬 OFERTAS PENDIENTES");
    System.out.println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━");
    System.out.println("(sin ofertas pendientes)");
    System.out.println();
    System.out.println("TODO: Implementar listado de ofertas");
    System.out.println("      - Guardar ofertas en onOffer()");
    System.out.println("      - Mostrar: offerId, producto, cantidad, precio");
  }

  private static void handleAceptarOferta(String[] parts, ConectorBolsa connector, MyTradingBot bot) {
    if (parts.length < 2) {
      System.out.println("❌ Uso: aceptar <offerId>");
      return;
    }

    String offerId = parts[1];

    System.out.println("\n✅ Aceptando oferta: " + offerId);
    System.out.println();
    System.out.println("TODO: Implementar aceptación de oferta");
    System.out.println("      1. Buscar oferta en Map de ofertas pendientes");
    System.out.println("      2. Validar que tengas el producto");
    System.out.println("      3. Llamar connector.aceptarOferta()");
  }

  private static void printHelp() {
    System.out.println("\n📚 AYUDA COMPLETA - Comandos del Trading Bot");
    System.out.println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━");
    System.out.println();
    System.out.println("INFORMACIÓN:");
    System.out.println("  status              - Muestra saldo, inventario, P&L%");
    System.out.println("  inventario          - Lista todos tus productos");
    System.out.println("  precios             - Precios actuales de mercado");
    System.out.println();
    System.out.println("TRADING:");
    System.out.println("  comprar PALTA-OIL 10 \"mensaje opcional\"");
    System.out.println("  vender FOSFO 5 \"otro mensaje\"");
    System.out.println();
    System.out.println("PRODUCCIÓN:");
    System.out.println("  producir PALTA-OIL basico");
    System.out.println("  producir GUACA premium");
    System.out.println();
    System.out.println("OFERTAS:");
    System.out.println("  ofertas             - Ver ofertas de otros traders");
    System.out.println("  aceptar OFFER-123   - Aceptar una oferta específica");
    System.out.println();
    System.out.println("OTROS:");
    System.out.println("  ayuda               - Muestra esta ayuda");
    System.out.println("  exit                - Salir del programa");
    System.out.println();
    System.out.println("💡 TIP: Lee AGENTS.md para guía de implementación");
    System.out.println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━");
  }

  /**
   * Your trading bot implementation.
   *
   * TODO for students: - Add your trading logic in each callback method - Store
   * state (inventory, balance, prices, etc.) - Implement buy/sell strategies -
   * Handle production logic
   */
  private static class MyTradingBot implements EventListener {

    @Override
    public void onLoginOk(LoginOKMessage loginOk) {
      // Guard clause
      if (loginOk == null) {
        return;
      }

      System.out.println("✅ LOGIN SUCCESSFUL!");
      System.out.println("   Team: " + loginOk.getTeam());
      System.out.println("   Species: " + loginOk.getSpecies());
      System.out.println("   Balance: $" + loginOk.getCurrentBalance());
      System.out.println();

      // TODO: Initialize your bot state here
      // - Store initial balance
      // - Store available products
      // - Initialize your strategy
    }

    @Override
    public void onError(ErrorMessage error) {
      // Guard clause
      if (error == null) {
        return;
      }

      System.err.println("❌ ERROR [" + error.getCode() + "]: " + error.getReason());

      // TODO: Handle errors
      // - Log the error
      // - Retry if needed
      // - Update your strategy
    }

    @Override
    public void onTicker(TickerMessage ticker) {
      // Guard clause
      if (ticker == null) {
        return;
      }

      // Print market data
      System.out.println("📊 TICKER: " + ticker.getProduct() + " | Bid: $" + ticker.getBestBid() + " | Ask: $"
          + ticker.getBestAsk() + " | Mid: $" + ticker.getMid());

      // TODO: Implement your trading strategy here
      // - Update price tracking
      // - Decide when to buy/sell
      // - Calculate profit opportunities
    }

    @Override
    public void onFill(FillMessage fill) {
      // Guard clause
      if (fill == null) {
        return;
      }

      System.out.println("✅ FILL: " + fill.getSide() + " " + fill.getFillQty() + " " + fill.getProduct() + " @ $"
          + fill.getFillPrice());

      // TODO: Update your state after a fill
      // - Update inventory
      // - Update balance
      // - Log the transaction
    }

    @Override
    public void onBalanceUpdate(BalanceUpdateMessage balanceUpdate) {
      // Guard clause
      if (balanceUpdate == null) {
        return;
      }

      System.out.println("💰 BALANCE UPDATE: " + balanceUpdate);

      // TODO: Track balance changes
      // - Extract balance from message
      // - Update your internal state
    }

    @Override
    public void onInventoryUpdate(InventoryUpdateMessage inventoryUpdate) {
      // Guard clause
      if (inventoryUpdate == null) {
        return;
      }

      System.out.println("📦 INVENTORY UPDATE: " + inventoryUpdate);

      // TODO: Track inventory changes
      // - Extract product and quantity from message
      // - Update your internal inventory map
    }

    @Override
    public void onOffer(OfferMessage offer) {
      // Students can implement if needed
    }

    @Override
    public void onOrderAck(OrderAckMessage orderAck) {
      // Students can implement if needed
    }

    @Override
    public void onEventDelta(EventDeltaMessage eventDelta) {
      // Students can implement if needed
    }

    @Override
    public void onBroadcast(BroadcastNotificationMessage broadcast) {
      // Guard clause
      if (broadcast == null) {
        return;
      }

      System.out.println("📢 BROADCAST: " + broadcast.getMessage());
    }

    @Override
    public void onConnectionLost(Throwable throwable) {
      System.err.println("💔 CONNECTION LOST: " + throwable.getMessage());

      // TODO: Implement reconnection logic
    }
  }
}
