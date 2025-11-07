package tech.hellsoft.trading.service;

public interface UIService {

  void printSuccess(String message);

  void printError(String message);

  void printInfo(String message);

  void printWarning(String message);

  void printHeader(String title);

  void printSeparator();

  void printConfigSummary(String team, String host, String maskedApiKey);

  void printStatus(String emoji, String message);

  void printWelcomeBanner();

  void printFooter();
}
