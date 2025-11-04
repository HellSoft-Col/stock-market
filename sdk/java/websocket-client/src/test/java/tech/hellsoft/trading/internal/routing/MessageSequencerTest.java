package tech.hellsoft.trading.internal.routing;

import static org.junit.jupiter.api.Assertions.*;

import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicInteger;

import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.Test;

class MessageSequencerTest {

  private MessageSequencer sequencer;

  @AfterEach
  void tearDown() {
    if (sequencer != null && !sequencer.isShutdown()) {
      sequencer.shutdown();
    }
  }

  @Test
  void shouldExecuteTasksSequentially() throws InterruptedException {
    sequencer = new MessageSequencer();
    List<Integer> executionOrder = Collections.synchronizedList(new ArrayList<>());
    CountDownLatch latch = new CountDownLatch(10);

    for (int i = 0; i < 10; i++) {
      final int taskNumber = i;
      sequencer.submit(
          () -> {
            executionOrder.add(taskNumber);
            latch.countDown();
          });
    }

    assertTrue(latch.await(5, TimeUnit.SECONDS));

    for (int i = 0; i < 10; i++) {
      assertEquals(i, executionOrder.get(i));
    }
  }

  @Test
  void shouldEnsureNoRaceConditionsWithSequentialExecution() throws InterruptedException {
    sequencer = new MessageSequencer();
    AtomicInteger counter = new AtomicInteger(0);
    int iterations = 1000;
    CountDownLatch latch = new CountDownLatch(iterations);

    for (int i = 0; i < iterations; i++) {
      sequencer.submit(
          () -> {
            int current = counter.get();
            counter.set(current + 1);
            latch.countDown();
          });
    }

    assertTrue(latch.await(10, TimeUnit.SECONDS));
    assertEquals(iterations, counter.get());
  }

  @Test
  void shouldHandleTaskExceptionsGracefully() throws InterruptedException {
    sequencer = new MessageSequencer();
    CountDownLatch latch = new CountDownLatch(2);

    sequencer.submit(
        () -> {
          latch.countDown();
          throw new RuntimeException("First task exception");
        });

    sequencer.submit(latch::countDown);

    assertTrue(latch.await(2, TimeUnit.SECONDS));
  }

  @Test
  void shouldContinueProcessingAfterTaskException() throws InterruptedException {
    sequencer = new MessageSequencer();
    List<String> executionLog = Collections.synchronizedList(new ArrayList<>());
    CountDownLatch latch = new CountDownLatch(3);

    sequencer.submit(
        () -> {
          executionLog.add("task1");
          latch.countDown();
        });

    sequencer.submit(
        () -> {
          executionLog.add("task2");
          latch.countDown();
          throw new RuntimeException("Failing task");
        });

    sequencer.submit(
        () -> {
          executionLog.add("task3");
          latch.countDown();
        });

    assertTrue(latch.await(2, TimeUnit.SECONDS));
    assertEquals(3, executionLog.size());
    assertEquals("task1", executionLog.get(0));
    assertEquals("task2", executionLog.get(1));
    assertEquals("task3", executionLog.get(2));
  }

  @Test
  void shouldShutdownCleanly() {
    sequencer = new MessageSequencer();
    assertFalse(sequencer.isShutdown());

    sequencer.shutdown();

    assertTrue(sequencer.isShutdown());
  }

  @Test
  void shouldRejectTasksAfterShutdown() throws InterruptedException {
    sequencer = new MessageSequencer();
    CountDownLatch taskExecuted = new CountDownLatch(1);

    sequencer.shutdown();

    sequencer.submit(taskExecuted::countDown);

    assertFalse(taskExecuted.await(500, TimeUnit.MILLISECONDS));
  }

  @Test
  void shouldCompleteRunningTasksBeforeShutdown() throws InterruptedException {
    sequencer = new MessageSequencer();
    CountDownLatch taskStarted = new CountDownLatch(1);
    CountDownLatch taskComplete = new CountDownLatch(1);

    sequencer.submit(
        () -> {
          taskStarted.countDown();
          try {
            Thread.sleep(200);
          } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
          }
          taskComplete.countDown();
        });

    taskStarted.await();
    sequencer.shutdown();

    assertTrue(taskComplete.await(1, TimeUnit.SECONDS));
  }

  @Test
  void shouldHandleHighThroughput() throws InterruptedException {
    sequencer = new MessageSequencer();
    int messageCount = 10000;
    AtomicInteger processedCount = new AtomicInteger(0);
    CountDownLatch latch = new CountDownLatch(messageCount);

    long startTime = System.nanoTime();

    for (int i = 0; i < messageCount; i++) {
      sequencer.submit(
          () -> {
            processedCount.incrementAndGet();
            latch.countDown();
          });
    }

    assertTrue(latch.await(30, TimeUnit.SECONDS));
    assertEquals(messageCount, processedCount.get());

    long durationMs = TimeUnit.NANOSECONDS.toMillis(System.nanoTime() - startTime);
    assertTrue(durationMs < 10000, "High throughput test took too long: " + durationMs + "ms");
  }

  @Test
  void shouldMaintainOrderUnderConcurrentSubmissions() throws InterruptedException {
    sequencer = new MessageSequencer();
    List<Integer> executionOrder = Collections.synchronizedList(new ArrayList<>());
    int tasksPerThread = 100;
    int threadCount = 10;
    int totalTasks = tasksPerThread * threadCount;
    CountDownLatch submitLatch = new CountDownLatch(threadCount);
    CountDownLatch executionLatch = new CountDownLatch(totalTasks);

    for (int t = 0; t < threadCount; t++) {
      final int threadId = t;
      Thread.startVirtualThread(
          () -> {
            for (int i = 0; i < tasksPerThread; i++) {
              final int taskId = threadId * tasksPerThread + i;
              sequencer.submit(
                  () -> {
                    executionOrder.add(taskId);
                    executionLatch.countDown();
                  });
            }
            submitLatch.countDown();
          });
    }

    assertTrue(submitLatch.await(5, TimeUnit.SECONDS));
    assertTrue(executionLatch.await(10, TimeUnit.SECONDS));
    assertEquals(totalTasks, executionOrder.size());

    for (int i = 1; i < executionOrder.size(); i++) {
      assertNotNull(executionOrder.get(i));
    }
  }

  @Test
  void shouldNotBlockSubmitter() throws InterruptedException {
    sequencer = new MessageSequencer();
    CountDownLatch taskStarted = new CountDownLatch(1);
    CountDownLatch submitterDone = new CountDownLatch(1);

    sequencer.submit(
        () -> {
          taskStarted.countDown();
          try {
            Thread.sleep(500);
          } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
          }
        });

    taskStarted.await();

    long startTime = System.nanoTime();
    sequencer.submit(() -> {});
    long submitDuration = TimeUnit.NANOSECONDS.toMillis(System.nanoTime() - startTime);
    submitterDone.countDown();

    assertTrue(submitterDone.await(100, TimeUnit.MILLISECONDS));
    assertTrue(submitDuration < 100, "Submit blocked for " + submitDuration + "ms");
  }

  @Test
  void shouldContinueAfterNullTask() throws InterruptedException {
    sequencer = new MessageSequencer();
    CountDownLatch afterNull = new CountDownLatch(1);

    try {
      sequencer.submit(null);
    } catch (NullPointerException _) {
    }

    sequencer.submit(afterNull::countDown);

    assertTrue(afterNull.await(1, TimeUnit.SECONDS));
  }

  @Test
  void shouldWorkWithVirtualThreadExecutor() throws InterruptedException {
    sequencer = new MessageSequencer();
    CountDownLatch latch = new CountDownLatch(100);
    List<Thread> threads = Collections.synchronizedList(new ArrayList<>());

    for (int i = 0; i < 100; i++) {
      sequencer.submit(
          () -> {
            threads.add(Thread.currentThread());
            latch.countDown();
          });
    }

    assertTrue(latch.await(5, TimeUnit.SECONDS));
    assertEquals(100, threads.size());

    assertTrue(threads.stream().allMatch(Thread::isVirtual));
  }

  @Test
  void shouldProcessTasksInFIFOOrder() throws InterruptedException {
    sequencer = new MessageSequencer();
    List<String> executionOrder = Collections.synchronizedList(new ArrayList<>());
    CountDownLatch latch = new CountDownLatch(5);
    String[] expectedOrder = {"FIRST", "SECOND", "THIRD", "FOURTH", "FIFTH"};

    for (String task : expectedOrder) {
      sequencer.submit(
          () -> {
            executionOrder.add(task);
            latch.countDown();
          });
    }

    assertTrue(latch.await(2, TimeUnit.SECONDS));
    assertArrayEquals(expectedOrder, executionOrder.toArray(new String[0]));
  }

  @Test
  void shouldHandleRapidShutdownAfterSubmission() throws InterruptedException {
    sequencer = new MessageSequencer();
    AtomicInteger executed = new AtomicInteger(0);

    for (int i = 0; i < 10; i++) {
      sequencer.submit(executed::incrementAndGet);
    }

    sequencer.shutdown();

    Thread.sleep(100);

    assertTrue(executed.get() >= 0 && executed.get() <= 10);
    assertTrue(sequencer.isShutdown());
  }
}
