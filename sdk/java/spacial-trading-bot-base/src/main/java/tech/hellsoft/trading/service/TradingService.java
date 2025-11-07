package tech.hellsoft.trading.service;

import tech.hellsoft.trading.config.Configuration;
import tech.hellsoft.trading.exception.ConfiguracionInvalidaException;
import tech.hellsoft.trading.exception.TradingException;

public interface TradingService {
  void start(Configuration config) throws ConfiguracionInvalidaException, TradingException;

  void stop();

  boolean isRunning();
}
