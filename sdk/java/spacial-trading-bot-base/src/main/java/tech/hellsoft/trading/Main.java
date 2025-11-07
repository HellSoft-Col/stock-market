package tech.hellsoft.trading;

import tech.hellsoft.trading.config.Configuration;
import tech.hellsoft.trading.exception.ConfiguracionInvalidaException;
import tech.hellsoft.trading.exception.TradingException;
import tech.hellsoft.trading.service.TradingService;
import tech.hellsoft.trading.service.UIService;
import tech.hellsoft.trading.service.impl.SDKTradingService;
import tech.hellsoft.trading.service.impl.ConsoleUIService;
import tech.hellsoft.trading.util.ConfigLoader;
import tech.hellsoft.trading.util.TradingUtils;

public final class Main {

  private Main() {
  }

  private static final String CONFIG_PATH = "src/main/resources/config.json";

  static void main(String[] args) {
    UIService uiService = new ConsoleUIService();
    TradingService tradingService = new SDKTradingService(uiService);

    try {
      uiService.printWelcomeBanner();

      Configuration config = loadConfiguration(uiService);

      String maskedApiKey = TradingUtils.maskApiKey(config.apiKey());
      uiService.printConfigSummary(config.team(), config.host(), maskedApiKey);

      tradingService.start(config);

      uiService.printFooter();

      waitForShutdown(tradingService, uiService);

    } catch (ConfiguracionInvalidaException e) {
      uiService.printError("Configuration Error: " + e.getMessage());
      System.exit(1);
    } catch (TradingException e) {
      uiService.printError("Trading Error: " + e.getMessage());
      System.exit(1);
    } catch (Exception e) {
      uiService.printError("Unexpected Error: " + e.getMessage());
      e.printStackTrace();
      System.exit(1);
    } finally {
      tradingService.stop();
    }
  }

  private static Configuration loadConfiguration(UIService uiService) throws ConfiguracionInvalidaException {
    if (CONFIG_PATH == null || CONFIG_PATH.isBlank()) {
      throw new ConfiguracionInvalidaException("Configuration path cannot be null or empty");
    }

    uiService.printStatus("📂", "Loading configuration from: " + CONFIG_PATH);
    return ConfigLoader.load(CONFIG_PATH);
  }

  private static void waitForShutdown(TradingService tradingService, UIService uiService) {
    uiService.printInfo("Press Ctrl+C to shutdown gracefully...");

    Runtime.getRuntime().addShutdownHook(new Thread(() -> {
      uiService.printStatus("🛑", "Shutting down...");
      tradingService.stop();
    }));

    try {
      while (tradingService.isRunning()) {
        Thread.sleep(1000);
      }
    } catch (InterruptedException e) {
      Thread.currentThread().interrupt();
      uiService.printWarning("Main thread interrupted");
    }
  }
}
