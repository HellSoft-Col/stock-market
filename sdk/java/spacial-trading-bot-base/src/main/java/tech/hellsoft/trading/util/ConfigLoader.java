package tech.hellsoft.trading.util;

import com.google.gson.Gson;
import tech.hellsoft.trading.config.Configuration;
import tech.hellsoft.trading.exception.ConfiguracionInvalidaException;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;

public final class ConfigLoader {

  private static final Gson GSON = new Gson();

  private ConfigLoader() {
  }

  public static Configuration load(String path) throws ConfiguracionInvalidaException {
    if (path == null || path.isBlank()) {
      throw new ConfiguracionInvalidaException("Configuration file path cannot be null or empty");
    }

    Path filePath = validatePath(path);

    try {
      String json = Files.readString(filePath);
      return parseConfiguration(json);
    } catch (IOException e) {
      throw new ConfiguracionInvalidaException("Failed to read configuration file: " + e.getMessage(), e);
    }
  }

  private static Path validatePath(String path) throws ConfiguracionInvalidaException {
    Path filePath = Paths.get(path);

    if (!Files.exists(filePath)) {
      throw new ConfiguracionInvalidaException("Configuration file not found: " + path);
    }

    if (!Files.isReadable(filePath)) {
      throw new ConfiguracionInvalidaException("Configuration file not readable: " + path);
    }

    if (!Files.isRegularFile(filePath)) {
      throw new ConfiguracionInvalidaException("Configuration path is not a regular file: " + path);
    }

    return filePath;
  }

  private static Configuration parseConfiguration(String json) throws ConfiguracionInvalidaException {
    if (json == null || json.isBlank()) {
      throw new ConfiguracionInvalidaException("Configuration file content is empty");
    }

    try {
      return GSON.fromJson(json, Configuration.class);
    } catch (Exception e) {
      throw new ConfiguracionInvalidaException("Failed to parse configuration JSON: " + e.getMessage(), e);
    }
  }
}
