package tech.hellsoft.trading.tasks;

import java.time.Duration;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.Semaphore;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicBoolean;

import lombok.extern.slf4j.Slf4j;

/**
 * Manages the lifecycle of automatic tasks with key-based locking.
 *
 * <p>This manager handles:
 *
 * <ul>
 *   <li>Automatic task scheduling and execution on virtual threads
 *   <li>Key-based locking to prevent concurrent execution of tasks with the same key
 *   <li>Graceful shutdown with resource cleanup
 *   <li>Error handling and recovery
 * </ul>
 *
 * <p>This class is thread-safe and can be shared across multiple threads.
 *
 * @see TareaAutomatica for creating custom tasks
 */
@Slf4j
public class TareaAutomaticaManager {

  /** Default timeout for executor termination in seconds. */
  private static final int SHUTDOWN_TIMEOUT_SECONDS = 10;

  /** Permit count for task execution semaphore (1 = mutual exclusion). */
  private static final int SEMAPHORE_PERMITS = 1;

  /** Shared executor for all tasks using virtual threads. */
  private final ExecutorService executor = Executors.newVirtualThreadPerTaskExecutor();

  /** Lock management per task key to prevent concurrent execution. */
  private final ConcurrentHashMap<String, Semaphore> locksByKey = new ConcurrentHashMap<>();

  /** Running tasks tracked by their key. */
  private final ConcurrentHashMap<String, TaskRunner> runningTasks = new ConcurrentHashMap<>();

  /** Flag indicating if the manager has been shut down. */
  private final AtomicBoolean shutdown = new AtomicBoolean(false);

  /**
   * Registers and starts an automatic task.
   *
   * <p>The task will start executing immediately according to its configuration. If a task with the
   * same key is already registered, it will be stopped and replaced with the new task.
   *
   * @param tarea the task to register and start (must not be null)
   * @throws IllegalArgumentException if tarea is null
   * @throws IllegalStateException if the manager has been shut down
   */
  public void registrar(TareaAutomatica tarea) {
    if (tarea == null) {
      throw new IllegalArgumentException("tarea cannot be null");
    }

    if (shutdown.get()) {
      throw new IllegalStateException("Manager has been shut down");
    }

    String taskKey = tarea.getTaskKey();

    // Stop existing task with the same key if present
    detener(taskKey);

    log.info("Registering automatic task: {}", taskKey);

    // Get or create semaphore for this task key
    Semaphore lock = locksByKey.computeIfAbsent(taskKey, k -> new Semaphore(SEMAPHORE_PERMITS));

    // Create and start the task runner
    TaskRunner runner = new TaskRunner(tarea, lock);
    runningTasks.put(taskKey, runner);
    runner.start();

    log.debug("Task registered and started: {}", taskKey);
  }

  /**
   * Stops and removes a registered task.
   *
   * <p>The task will complete its current execution and then stop. If the task is not registered,
   * this method does nothing.
   *
   * @param taskKey the key of the task to stop (may be null)
   */
  public void detener(String taskKey) {
    if (taskKey == null) {
      return;
    }

    TaskRunner runner = runningTasks.remove(taskKey);
    if (runner == null) {
      return;
    }

    log.info("Stopping automatic task: {}", taskKey);
    runner.stop();
    cleanupLockIfNotInUse(taskKey);
    log.debug("Task stopped: {}", taskKey);
  }

  private void cleanupLockIfNotInUse(String taskKey) {
    locksByKey.computeIfPresent(taskKey, (key, lock) -> lock.hasQueuedThreads() ? lock : null);
  }

  /**
   * Stops all registered tasks and shuts down the executor.
   *
   * <p>This method blocks until all tasks have stopped or the timeout expires. After calling this
   * method, no new tasks can be registered.
   *
   * <p>This method is idempotent and can be called multiple times safely.
   */
  public void shutdown() {
    if (!shutdown.compareAndSet(false, true)) {
      return; // Already shut down
    }

    log.info("Shutting down task manager ({} tasks)", runningTasks.size());

    stopAllTasks();
    shutdownExecutor();
    locksByKey.clear();

    log.info("Task manager shutdown complete");
  }

  private void stopAllTasks() {
    runningTasks.keySet().forEach(this::detener);
    runningTasks.clear();
  }

  private void shutdownExecutor() {
    executor.shutdown();
    try {
      awaitExecutorTermination();
    } catch (InterruptedException e) {
      Thread.currentThread().interrupt();
      executor.shutdownNow();
    }
  }

  private void awaitExecutorTermination() throws InterruptedException {
    if (executor.awaitTermination(SHUTDOWN_TIMEOUT_SECONDS, TimeUnit.SECONDS)) {
      return;
    }
    log.warn("Executor did not terminate in time, forcing shutdown");
    executor.shutdownNow();
  }

  /**
   * Returns the number of currently registered tasks.
   *
   * @return the number of running tasks
   */
  public int getTaskCount() {
    return runningTasks.size();
  }

  /**
   * Checks if a task with the given key is currently registered.
   *
   * @param taskKey the task key to check
   * @return true if the task is registered, false otherwise
   */
  public boolean isTaskRegistered(String taskKey) {
    return runningTasks.containsKey(taskKey);
  }

  /** Internal class to manage task execution lifecycle. */
  private class TaskRunner {
    private final TareaAutomatica tarea;
    private final Semaphore lock;
    private final AtomicBoolean running = new AtomicBoolean(false);
    private volatile CompletableFuture<Void> currentExecution;

    TaskRunner(TareaAutomatica tarea, Semaphore lock) {
      this.tarea = tarea;
      this.lock = lock;
    }

    void start() {
      if (!running.compareAndSet(false, true)) {
        return; // Already running
      }

      currentExecution = CompletableFuture.runAsync(this::executionLoop, executor);
      currentExecution.whenComplete(this::handleCompletion);
    }

    void stop() {
      if (!running.compareAndSet(true, false)) {
        return; // Already stopped
      }

      cancelExecution();
      invokeCleanupHook();
    }

    private void executionLoop() {
      String taskKey = tarea.getTaskKey();
      boolean continuous = tarea.ejecucionContinua();
      Duration interval = getValidatedInterval(continuous);

      log.debug("Task started: {} (continuous: {}, interval: {})", taskKey, continuous, interval);

      while (running.get()) {
        if (!executeTaskSafely(taskKey)) {
          break;
        }

        if (shouldSleepBetweenExecutions(continuous, interval)) {
          if (!sleepBetweenExecutions(interval, taskKey)) {
            break;
          }
        }
      }

      log.debug("Task execution loop ended: {}", taskKey);
    }

    private boolean executeTaskSafely(String taskKey) {
      try {
        lock.acquire();
        try {
          tarea.ejecutar();
        } finally {
          lock.release();
        }
        return true;
      } catch (InterruptedException e) {
        Thread.currentThread().interrupt();
        log.debug("Task interrupted: {}", taskKey);
        return false;
      } catch (Exception e) {
        log.error("Error executing task {}: {}", taskKey, e.getMessage(), e);
        return true; // Continue despite errors
      }
    }

    private boolean shouldSleepBetweenExecutions(boolean continuous, Duration interval) {
      return !continuous && interval != null && !interval.isZero();
    }

    private boolean sleepBetweenExecutions(Duration interval, String taskKey) {
      try {
        Thread.sleep(interval.toMillis());
        return true;
      } catch (InterruptedException e) {
        Thread.currentThread().interrupt();
        log.debug("Task interrupted during sleep: {}", taskKey);
        return false;
      }
    }

    private void handleCompletion(Void result, Throwable throwable) {
      if (throwable == null) {
        return;
      }
      log.error("Task {} failed: {}", tarea.getTaskKey(), throwable.getMessage(), throwable);
    }

    private void cancelExecution() {
      if (currentExecution == null) {
        return;
      }
      currentExecution.cancel(true);
    }

    private void invokeCleanupHook() {
      try {
        tarea.onDetener();
      } catch (Exception e) {
        log.error("Error in onDetener for task {}: {}", tarea.getTaskKey(), e.getMessage(), e);
      }
    }

    private Duration getValidatedInterval(boolean continuous) {
      if (continuous) {
        return Duration.ZERO;
      }

      Duration interval = tarea.intervalo();
      if (interval == null) {
        log.warn("Task {} returned null interval, using default 1 second", tarea.getTaskKey());
        return Duration.ofSeconds(1);
      }

      if (interval.isNegative()) {
        log.warn("Task {} returned negative interval, using default 1 second", tarea.getTaskKey());
        return Duration.ofSeconds(1);
      }

      return interval;
    }
  }
}
