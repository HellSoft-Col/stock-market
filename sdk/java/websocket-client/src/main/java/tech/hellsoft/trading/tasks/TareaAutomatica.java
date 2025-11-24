package tech.hellsoft.trading.tasks;

import java.time.Duration;

import lombok.Getter;

/**
 * Abstract base class for automatic tasks that run periodically or continuously.
 *
 * <p>Tasks are executed on virtual threads and support key-based locking to prevent concurrent
 * execution of tasks with the same key. This is useful for implementing trading strategies,
 * monitoring systems, and automated market operations.
 *
 * <p>Example usage:
 *
 * <pre>{@code
 * public class MonitorPreciosTarea extends TareaAutomatica {
 *     private final ConectorBolsa connector;
 *     private final String producto;
 *
 *     public MonitorPreciosTarea(ConectorBolsa connector, String producto) {
 *         super("monitor-" + producto);
 *         this.connector = connector;
 *         this.producto = producto;
 *     }
 *
 *     @Override
 *     protected void ejecutar() {
 *         // Check prices and send orders
 *         log.info("Monitoring prices for {}", producto);
 *     }
 *
 *     @Override
 *     protected Duration intervalo() {
 *         return Duration.ofMillis(100); // Fast execution - 100ms between runs
 *     }
 * }
 *
 * // Register with ConectorBolsa
 * connector.registrarTarea(new MonitorPreciosTarea(connector, "GUACA"));
 * }</pre>
 *
 * <p>For continuous execution with minimal delay:
 *
 * <pre>{@code
 * public class MonitorContinuoTarea extends TareaAutomatica {
 *     @Override
 *     protected boolean ejecucionContinua() {
 *         return true; // Run continuously with no sleep
 *     }
 *
 *     @Override
 *     protected void ejecutar() {
 *         // This will run in a tight loop
 *     }
 * }
 * }</pre>
 *
 * @see tech.hellsoft.trading.tasks.TareaAutomaticaManager
 */
public abstract class TareaAutomatica {

  /** Unique key identifying this task, used for locking and management. */
  @Getter private final String taskKey;

  /**
   * Creates a new automatic task with the specified key.
   *
   * <p>The task key is used for:
   *
   * <ul>
   *   <li>Preventing concurrent execution of tasks with the same key
   *   <li>Identifying the task in logs
   *   <li>Managing the task lifecycle (stop, restart)
   * </ul>
   *
   * @param taskKey unique identifier for this task (must not be null or blank)
   * @throws IllegalArgumentException if taskKey is null or blank
   */
  protected TareaAutomatica(String taskKey) {
    if (taskKey == null || taskKey.isBlank()) {
      throw new IllegalArgumentException("taskKey cannot be null or blank");
    }
    this.taskKey = taskKey;
  }

  /**
   * Executes the task logic.
   *
   * <p>This method is called repeatedly according to the configured interval or continuously if
   * {@link #ejecucionContinua()} returns true. Implementations should:
   *
   * <ul>
   *   <li>Keep execution time reasonable to avoid blocking other tasks
   *   <li>Handle exceptions gracefully (uncaught exceptions are logged but won't stop the task)
   *   <li>Check for interruption if performing long operations
   * </ul>
   *
   * <p>This method is guaranteed not to run concurrently for tasks with the same key.
   */
  protected abstract void ejecutar();

  /**
   * Returns the delay between task executions.
   *
   * <p>This method is ignored if {@link #ejecucionContinua()} returns true. The default
   * implementation returns 1 second. Override this method to customize the execution interval.
   *
   * <p>Examples:
   *
   * <ul>
   *   <li>{@code Duration.ofMillis(100)} - Run every 100ms (fast execution)
   *   <li>{@code Duration.ofSeconds(5)} - Run every 5 seconds
   *   <li>{@code Duration.ofMinutes(1)} - Run every minute
   * </ul>
   *
   * @return the delay between executions (must not be null or negative)
   */
  protected Duration intervalo() {
    return Duration.ofSeconds(1); // Default 1 second interval
  }

  /**
   * Returns whether this task should run continuously without delay between executions.
   *
   * <p>When this returns true, the task runs in a tight loop with minimal delay, ignoring the value
   * returned by {@link #intervalo()}. This is useful for high-frequency tasks that need to process
   * data as fast as possible.
   *
   * <p><strong>Warning:</strong> Continuous execution can consume significant CPU resources. Use
   * this mode only when necessary.
   *
   * @return true if the task should run continuously, false otherwise (default)
   */
  protected boolean ejecucionContinua() {
    return false; // Default to interval-based execution
  }

  /**
   * Called when the task is stopped.
   *
   * <p>Override this method to perform cleanup operations such as releasing resources, closing
   * connections, or saving state. This method is called exactly once when the task is stopped or
   * during manager shutdown.
   *
   * <p>The default implementation does nothing.
   */
  protected void onDetener() {
    // Override to perform cleanup
  }
}
