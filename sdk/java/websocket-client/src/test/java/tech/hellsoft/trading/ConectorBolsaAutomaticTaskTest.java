package tech.hellsoft.trading;

import static org.junit.jupiter.api.Assertions.*;

import java.time.Duration;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicInteger;

import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import tech.hellsoft.trading.tasks.TareaAutomatica;

import lombok.extern.slf4j.Slf4j;

@Slf4j
class ConectorBolsaAutomaticTaskTest {

  private ConectorBolsa conector;

  @BeforeEach
  void setUp() {
    conector = new ConectorBolsa();
  }

  @AfterEach
  void tearDown() {
    if (conector != null) {
      conector.shutdown();
    }
  }

  @Test
  void testRegistrarTarea() throws InterruptedException {
    AtomicInteger executionCount = new AtomicInteger(0);
    CountDownLatch latch = new CountDownLatch(3);

    TareaAutomatica task =
        new TareaAutomatica("test-task") {
          @Override
          protected void ejecutar() {
            executionCount.incrementAndGet();
            latch.countDown();
          }

          @Override
          protected Duration intervalo() {
            return Duration.ofMillis(50);
          }
        };

    conector.registrarTarea(task);

    assertTrue(latch.await(5, TimeUnit.SECONDS), "Task should execute at least 3 times");
    assertTrue(executionCount.get() >= 3);
  }

  @Test
  void testDetenerTarea() throws InterruptedException {
    AtomicInteger executionCount = new AtomicInteger(0);

    TareaAutomatica task =
        new TareaAutomatica("stoppable-task") {
          @Override
          protected void ejecutar() {
            executionCount.incrementAndGet();
          }

          @Override
          protected Duration intervalo() {
            return Duration.ofMillis(50);
          }
        };

    conector.registrarTarea(task);
    Thread.sleep(200); // Let it run

    int countBeforeStop = executionCount.get();
    conector.detenerTarea("stoppable-task");

    Thread.sleep(200); // Wait after stop
    int countAfterStop = executionCount.get();

    assertTrue(
        countAfterStop <= countBeforeStop + 1, "Task should stop after detenerTarea is called");
  }

  @Test
  void testMultipleTasks() throws InterruptedException {
    AtomicInteger task1Count = new AtomicInteger(0);
    AtomicInteger task2Count = new AtomicInteger(0);
    CountDownLatch latch = new CountDownLatch(6);

    TareaAutomatica task1 =
        new TareaAutomatica("task-1") {
          @Override
          protected void ejecutar() {
            task1Count.incrementAndGet();
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
            task2Count.incrementAndGet();
            latch.countDown();
          }

          @Override
          protected Duration intervalo() {
            return Duration.ofMillis(50);
          }
        };

    conector.registrarTarea(task1);
    conector.registrarTarea(task2);

    assertTrue(latch.await(5, TimeUnit.SECONDS), "Both tasks should execute");
    assertTrue(task1Count.get() >= 3);
    assertTrue(task2Count.get() >= 3);
  }

  @Test
  void testShutdownStopsAllTasks() throws InterruptedException {
    AtomicInteger executionCount = new AtomicInteger(0);

    TareaAutomatica task =
        new TareaAutomatica("shutdown-task") {
          @Override
          protected void ejecutar() {
            executionCount.incrementAndGet();
          }

          @Override
          protected Duration intervalo() {
            return Duration.ofMillis(50);
          }
        };

    conector.registrarTarea(task);
    Thread.sleep(200); // Let it run

    int countBeforeShutdown = executionCount.get();
    conector.shutdown();

    Thread.sleep(200); // Wait after shutdown
    int countAfterShutdown = executionCount.get();

    assertTrue(
        countAfterShutdown <= countBeforeShutdown + 1, "All tasks should stop after shutdown");
  }

  @Test
  void testRegistrarTareaWithNullThrowsException() {
    assertThrows(IllegalArgumentException.class, () -> conector.registrarTarea(null));
  }

  @Test
  void testDetenerTareaWithNullDoesNothing() {
    assertDoesNotThrow(() -> conector.detenerTarea(null));
  }

  @Test
  void testTaskCanAccessConectorBolsa() throws InterruptedException {
    CountDownLatch latch = new CountDownLatch(1);
    AtomicInteger hasAccess = new AtomicInteger(0);

    TareaAutomatica task =
        new TareaAutomatica("access-test") {
          @Override
          protected void ejecutar() {
            // Verify we can access the connector
            if (conector != null) {
              hasAccess.set(1);
              latch.countDown();
            }
          }

          @Override
          protected Duration intervalo() {
            return Duration.ofMillis(50);
          }
        };

    conector.registrarTarea(task);

    assertTrue(latch.await(5, TimeUnit.SECONDS), "Task should have access to conector");
    assertEquals(1, hasAccess.get());
  }

  @Test
  void testContinuousExecutionTask() throws InterruptedException {
    AtomicInteger executionCount = new AtomicInteger(0);
    CountDownLatch latch = new CountDownLatch(20);

    TareaAutomatica task =
        new TareaAutomatica("continuous-task") {
          @Override
          protected void ejecutar() {
            executionCount.incrementAndGet();
            latch.countDown();
          }

          @Override
          protected boolean ejecucionContinua() {
            return true; // Continuous execution
          }
        };

    conector.registrarTarea(task);

    // Continuous tasks execute rapidly
    assertTrue(latch.await(5, TimeUnit.SECONDS), "Continuous task should execute rapidly");
    assertTrue(executionCount.get() >= 20);

    conector.detenerTarea("continuous-task");
  }

  @Test
  void testTaskWithCleanup() throws InterruptedException {
    AtomicInteger cleanupCalled = new AtomicInteger(0);

    TareaAutomatica task =
        new TareaAutomatica("cleanup-task") {
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
            cleanupCalled.incrementAndGet();
          }
        };

    conector.registrarTarea(task);
    Thread.sleep(100);

    conector.detenerTarea("cleanup-task");
    Thread.sleep(100);

    assertEquals(1, cleanupCalled.get(), "Cleanup should be called exactly once");
  }
}
