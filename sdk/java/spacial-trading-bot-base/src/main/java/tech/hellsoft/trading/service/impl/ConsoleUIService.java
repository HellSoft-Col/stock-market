package tech.hellsoft.trading.service.impl;

import tech.hellsoft.trading.service.UIService;

public class ConsoleUIService implements UIService {

  private static final String RESET = "\u001B[0m";
  private static final String GREEN = "\u001B[32m";
  private static final String RED = "\u001B[31m";
  private static final String BLUE = "\u001B[34m";
  private static final String YELLOW = "\u001B[33m";
  private static final String CYAN = "\u001B[36m";
  private static final String BOLD = "\u001B[1m";

  @Override
  public void printSuccess(String message) {
    System.out.println(GREEN + "✅ " + message + RESET);
  }

  @Override
  public void printError(String message) {
    System.err.println(RED + "❌ " + message + RESET);
  }

  @Override
  public void printInfo(String message) {
    System.out.println(BLUE + "ℹ️  " + message + RESET);
  }

  @Override
  public void printWarning(String message) {
    System.out.println(YELLOW + "⚠️  " + message + RESET);
  }

  @Override
  public void printHeader(String title) {
    String border = "═".repeat(title.length() + 4);
    System.out.println();
    System.out.println(CYAN + BOLD + "╔" + border + "╗" + RESET);
    System.out.println(CYAN + BOLD + "║  " + title + "  ║" + RESET);
    System.out.println(CYAN + BOLD + "╚" + border + "╝" + RESET);
    System.out.println();
  }

  @Override
  public void printSeparator() {
    System.out.println(CYAN + "─".repeat(60) + RESET);
  }

  @Override
  public void printConfigSummary(String team, String host, String maskedApiKey) {
    printInfo("Configuration loaded successfully:");
    System.out.println("   " + CYAN + "Team:" + RESET + " " + team);
    System.out.println("   " + CYAN + "Host:" + RESET + " " + host);
    System.out.println("   " + CYAN + "API Key:" + RESET + " " + maskedApiKey);
  }

  @Override
  public void printStatus(String emoji, String message) {
    System.out.println(emoji + " " + message);
  }

  public void printWelcomeBanner() {
    String banner = CYAN + "╔══════════════════════════════════════════════════════════════╗\n" + CYAN + BOLD
        + "║                                                              ║\n" + CYAN + BOLD
        + "║  🚀 SPACIAL TRADING BOT CLIENT - Java 25 Edition 🚀         ║\n" + CYAN + BOLD
        + "║                                                              ║\n" + CYAN + BOLD
        + "║  🥑 Bolsa Interestelar de Aguacates Andorianos              ║\n" + CYAN + BOLD
        + "║  Ready for trading operations...                            ║\n" + CYAN + BOLD
        + "║                                                              ║\n" + CYAN + BOLD
        + "╚════════════════════════════════════════════════════════════╝" + RESET;

    System.out.println(banner);
  }

  public void printFooter() {
    System.out.println();
    printSeparator();
    printInfo("💡 Students can now add their trading logic here!");
    printInfo("   Implement your strategies in the TradingService implementation.");
    printSeparator();
  }
}
