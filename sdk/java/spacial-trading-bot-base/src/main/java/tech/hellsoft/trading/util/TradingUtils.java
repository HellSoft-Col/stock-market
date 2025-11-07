package tech.hellsoft.trading.util;

public final class TradingUtils {

  private TradingUtils() {
  }

  public static String maskApiKey(String apiKey) {
    if (apiKey == null || apiKey.length() < 8) {
      return "***";
    }
    return apiKey.substring(0, 4) + "***" + apiKey.substring(apiKey.length() - 4);
  }

  public static String formatCurrency(double amount) {
    return String.format("%.2f", amount);
  }

  public static String formatPercentage(double percentage) {
    return String.format("%.2f%%", percentage);
  }

  public static boolean isValidString(String value) {
    return value != null && !value.trim().isBlank();
  }

  public static boolean isPositive(double value) {
    return value > 0;
  }

  public static boolean isPositive(int value) {
    return value > 0;
  }
}
