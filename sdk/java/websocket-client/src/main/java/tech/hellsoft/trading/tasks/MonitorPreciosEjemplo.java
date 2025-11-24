package tech.hellsoft.trading.tasks;

import java.time.Duration;

import tech.hellsoft.trading.ConectorBolsa;

import lombok.extern.slf4j.Slf4j;

/**
 * Example implementation of an automatic task that monitors prices.
 *
 * <p>This is a simple example demonstrating how to create custom automatic tasks that can interact
 * with the ConectorBolsa to monitor market data and execute trading strategies.
 */
@Slf4j
public class MonitorPreciosEjemplo extends TareaAutomatica {

  private final ConectorBolsa connector;
  private final String producto;
  private int executionCount = 0;

  /**
   * Creates a new price monitoring task for the specified product.
   *
   * @param connector the ConectorBolsa instance to use for sending orders
   * @param producto the product symbol to monitor
   */
  public MonitorPreciosEjemplo(ConectorBolsa connector, String producto) {
    super("monitor-precio-" + producto);
    this.connector = connector;
    this.producto = producto;
  }

  @Override
  protected void ejecutar() {
    executionCount++;
    log.info("Monitoring prices for {} (execution #{})", producto, executionCount);

    // Here you would implement your trading logic:
    // - Check current prices
    // - Analyze market conditions
    // - Send orders if conditions are met
    // Example:
    // if (shouldBuy()) {
    //     connector.enviarOrden(createBuyOrder());
    // }
  }

  @Override
  protected Duration intervalo() {
    // Run every 5 seconds
    return Duration.ofSeconds(5);
  }

  @Override
  protected void onDetener() {
    log.info("Stopping price monitor for {} after {} executions", producto, executionCount);
  }
}
