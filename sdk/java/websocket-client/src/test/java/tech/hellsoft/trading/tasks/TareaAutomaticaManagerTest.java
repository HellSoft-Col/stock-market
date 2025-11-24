package tech.hellsoft.trading.tasks;

import static org.junit.jupiter.api.Assertions.*;

import java.time.Duration;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicInteger;

import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import lombok.extern.slf4j.Slf4j;

@Slf4j
class TareaAutomaticaManagerTest {

  private TareaAutomaticaManager manager;

  @BeforeEach
  void setUp() {
    manager = new TareaAutomaticaManager();
  }

  @AfterEach
  void tearDown() {
    if (manager != null) {
      manager.shutdown();
    }
  }

  @Test
  void testTaskRegistrationAndExecution() throws InterruptedException {
    AtomicInteger counter = new AtomicInteger(0);
    CountDownLatch latch = new CountDownLatch(3);

    TareaAutomatica task =
        new TareaAutomatica("test-task") {
          @Override
          protected void ejecutar() {
            counter.incrementAndGet();
            latch.countDown();
          }

          @Override
          protected Duration intervalo() {
            return Duration.ofMillis(50);
          }
        };

    manager.registrar(task);

    // Wait for at least 3 executions
    assertTrue(latch.await(5, TimeUnit.SECONDS), "Task should execute at least 3 times");
    assertTrue(counter.get() >= 3, "Counter should be at least 3");
    assertTrue(manager.isTaskRegistered("test-task"), "Task should be registered");
  }

  @Test
  void testContinuousExecution() throws InterruptedException {
    AtomicInteger counter = new AtomicInteger(0);
    CountDownLatch latch = new CountDownLatch(10);

    TareaAutomatica task =
        new TareaAutomatica("continuous-task") {
          @Override
          protected void ejecutar() {
            counter.incrementAndGet();
            latch.countDown();
          }

          @Override
          protected boolean ejecucionContinua() {
            return true;
          }
        };

    manager.registrar(task);

    // Continuous tasks should execute many times quickly
    assertTrue(latch.await(5, TimeUnit.SECONDS), "Task should execute at least 10 times");
    assertTrue(counter.get() >= 10, "Counter should be at least 10");
  }

  @Test
  void testTaskStop() throws InterruptedException {
    AtomicInteger counter = new AtomicInteger(0);

    TareaAutomatica task =
        new TareaAutomatica("stop-test-task") {
          @Override
          protected void ejecutar() {
            counter.incrementAndGet();
          }

          @Override
          protected Duration intervalo() {
            return Duration.ofMillis(50);
          }
        };

    manager.registrar(task);
    Thread.sleep(200); // Let it run a few times

    int countBeforeStop = counter.get();
    manager.detener("stop-test-task");

    Thread.sleep(200); // Wait after stopping
    int countAfterStop = counter.get();

    // Counter should not increase significantly after stopping
    assertTrue(
        countAfterStop <= countBeforeStop + 1,
        "Task should stop executing after detener() is called");
    assertFalse(manager.isTaskRegistered("stop-test-task"), "Task should not be registered");
  }

  @Test
  void testKeyBasedLocking() throws InterruptedException {
    String taskKey = "locked-task";
    AtomicInteger concurrentExecutions = new AtomicInteger(0);
    AtomicInteger maxConcurrent = new AtomicInteger(0);
    CountDownLatch latch = new CountDownLatch(10);

    // Create a single task that runs multiple times
    TareaAutomatica task =
        new TareaAutomatica(taskKey) {
          @Override
          protected void ejecutar() {
            int current = concurrentExecutions.incrementAndGet();
            maxConcurrent.updateAndGet(max -> Math.max(max, current));

            try {
              Thread.sleep(50); // Simulate work
            } catch (InterruptedException e) {
              Thread.currentThread().interrupt();
            }

            concurrentExecutions.decrementAndGet();
            latch.countDown();
          }

          @Override
          protected Duration intervalo() {
            return Duration.ofMillis(10); // Try to run frequently
          }
        };

    // Register the task
    manager.registrar(task);

    assertTrue(latch.await(10, TimeUnit.SECONDS), "Task should execute multiple times");

    // With proper locking, concurrent executions should never exceed 1
    // Even with a short interval, the lock prevents overlapping executions
    assertEquals(1, maxConcurrent.get(), "Only one execution should run at a time for same key");
  }

  @Test
  void testOnDetenerCallback() throws InterruptedException {
    AtomicInteger detenerCalled = new AtomicInteger(0);

    TareaAutomatica task =
        new TareaAutomatica("callback-task") {
          @Override
          protected void ejecutar() {
            // Do nothing
          }

          @Override
          protected Duration intervalo() {
            return Duration.ofMillis(100);
          }

          @Override
          protected void onDetener() {
            detenerCalled.incrementAndGet();
          }
        };

    manager.registrar(task);
    Thread.sleep(100); // Let it start

    manager.detener("callback-task");
    Thread.sleep(100); // Give time for cleanup

    assertEquals(1, detenerCalled.get(), "onDetener should be called exactly once");
  }

  @Test
  void testMultipleTasks() throws InterruptedException {
    AtomicInteger task1Counter = new AtomicInteger(0);
    AtomicInteger task2Counter = new AtomicInteger(0);
    CountDownLatch latch = new CountDownLatch(6); // 3 executions per task

    TareaAutomatica task1 =
        new TareaAutomatica("task-1") {
          @Override
          protected void ejecutar() {
            task1Counter.incrementAndGet();
            latch.countDown();
          }

          @Override
          protected Duration intervalo() {
            return Duration.ofMillis(50);
          }
        };

    TareaAutomatica task2 =
        new TareaAutomatica("task-2") {
          @Override
          protected void ejecutar() {
            task2Counter.incrementAndGet();
            latch.countDown();
          }

          @Override
          protected Duration intervalo() {
            return Duration.ofMillis(50);
          }
        };

    manager.registrar(task1);
    manager.registrar(task2);

    assertTrue(latch.await(5, TimeUnit.SECONDS), "Both tasks should execute");
    assertTrue(task1Counter.get() >= 3, "Task 1 should execute at least 3 times");
    assertTrue(task2Counter.get() >= 3, "Task 2 should execute at least 3 times");
    assertEquals(2, manager.getTaskCount(), "Should have 2 registered tasks");
  }

  @Test
  void testShutdown() throws InterruptedException {
    AtomicInteger counter = new AtomicInteger(0);

    TareaAutomatica task =
        new TareaAutomatica("shutdown-test") {
          @Override
          protected void ejecutar() {
            counter.incrementAndGet();
          }

          @Override
          protected Duration intervalo() {
            return Duration.ofMillis(50);
          }
        };

    manager.registrar(task);
    Thread.sleep(200); // Let it run a few times

    int countBeforeShutdown = counter.get();
    manager.shutdown();

    Thread.sleep(200); // Wait after shutdown
    int countAfterShutdown = counter.get();

    // Task should stop after shutdown
    assertTrue(countAfterShutdown <= countBeforeShutdown + 1, "Task should stop after shutdown");
    assertEquals(0, manager.getTaskCount(), "All tasks should be removed after shutdown");
  }

  @Test
  void testExceptionHandling() throws InterruptedException {
    AtomicInteger executionCount = new AtomicInteger(0);
    CountDownLatch latch = new CountDownLatch(5);

    TareaAutomatica task =
        new TareaAutomatica("exception-task") {
          @Override
          protected void ejecutar() {
            executionCount.incrementAndGet();
            latch.countDown();
            if (executionCount.get() == 2) {
              throw new RuntimeException("Test exception");
            }
          }

          @Override
          protected Duration intervalo() {
            return Duration.ofMillis(50);
          }
        };

    manager.registrar(task);

    // Task should continue executing despite exception
    assertTrue(latch.await(5, TimeUnit.SECONDS), "Task should continue executing after exception");
    assertTrue(executionCount.get() >= 5, "Task should execute at least 5 times despite exception");
  }

  @Test
  void testTaskReplacement() throws InterruptedException {
    AtomicInteger task1Executions = new AtomicInteger(0);
    AtomicInteger task2Executions = new AtomicInteger(0);

    TareaAutomatica task1 =
        new TareaAutomatica("replaceable-task") {
          @Override
          protected void ejecutar() {
            task1Executions.incrementAndGet();
          }

          @Override
          protected Duration intervalo() {
            return Duration.ofMillis(50);
          }
        };

    manager.registrar(task1);
    Thread.sleep(200); // Let first task run

    // Register another task with same key
    TareaAutomatica task2 =
        new TareaAutomatica("replaceable-task") {
          @Override
          protected void ejecutar() {
            task2Executions.incrementAndGet();
          }

          @Override
          protected Duration intervalo() {
            return Duration.ofMillis(50);
          }
        };

    int task1CountBeforeReplacement = task1Executions.get();
    manager.registrar(task2); // Should replace task1
    Thread.sleep(200); // Let second task run

    // First task should have stopped
    assertEquals(
        task1CountBeforeReplacement,
        task1Executions.get(),
        "First task should stop after replacement");

    // Second task should be running
    assertTrue(task2Executions.get() > 0, "Second task should be executing");
    assertEquals(1, manager.getTaskCount(), "Should have exactly 1 task");
  }
}
