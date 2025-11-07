package tech.hellsoft.trading.config;

import java.io.Serializable;

public record Configuration(String apiKey, String team, String host) implements Serializable {

  public Configuration {
    if (apiKey == null || apiKey.isBlank()) {
      throw new IllegalArgumentException("API key cannot be null or blank");
    }
    if (team == null || team.isBlank()) {
      throw new IllegalArgumentException("Team cannot be null or blank");
    }
    if (host == null || host.isBlank()) {
      throw new IllegalArgumentException("Host cannot be null or blank");
    }
  }
}
