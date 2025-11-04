package tech.hellsoft.trading.internal.routing;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.MethodSource;
import tech.hellsoft.trading.exception.StateLockException;

import java.time.Duration;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.stream.Stream;

import static org.junit.jupiter.api.Assertions.*;

class StateLockerTest {

    @Test
    void shouldAcquireLockAndExecuteAction() throws StateLockException {
        StateLocker locker = new StateLocker(Duration.ofSeconds(1));
        AtomicInteger counter = new AtomicInteger(0);

        int result = locker.withLock(() -> {
            counter.incrementAndGet();
            return 42;
        });

        assertEquals(42, result);
        assertEquals(1, counter.get());
    }

    @Test
    void shouldExecuteVoidActionWithLock() throws StateLockException {
        StateLocker locker = new StateLocker(Duration.ofSeconds(1));
        AtomicInteger counter = new AtomicInteger(0);

        locker.withLockVoid(() -> counter.incrementAndGet());

        assertEquals(1, counter.get());
    }

    @Test
    void shouldReleaseLockAfterExecution() throws StateLockException {
        StateLocker locker = new StateLocker(Duration.ofSeconds(1));

        locker.withLock(() -> "first");
        locker.withLock(() -> "second");

        assertDoesNotThrow(() -> locker.withLock(() -> "third"));
    }

    @Test
    void shouldReleaseLockEvenWhenActionThrows() {
        StateLocker locker = new StateLocker(Duration.ofSeconds(1));

        assertThrows(RuntimeException.class, () ->
            locker.withLock(() -> {
                throw new RuntimeException("Test exception");
            })
        );

        assertDoesNotThrow(() -> locker.withLock(() -> "can acquire after exception"));
    }

    @Test
    void shouldReleaseLockEvenWhenVoidActionThrows() {
        StateLocker locker = new StateLocker(Duration.ofSeconds(1));

        assertThrows(RuntimeException.class, () ->
            locker.withLockVoid(() -> {
                throw new RuntimeException("Test exception");
            })
        );

        assertDoesNotThrow(() -> locker.withLockVoid(() -> {}));
    }

    @Test
    void shouldTimeoutWhenLockHeldByAnotherThread() throws Exception {
        StateLocker locker = new StateLocker(Duration.ofMillis(100));
        CountDownLatch lockAcquired = new CountDownLatch(1);
        CountDownLatch lockHeld = new CountDownLatch(1);

        Thread holder = Thread.startVirtualThread(() -> {
            try {
                locker.withLockVoid(() -> {
                    lockAcquired.countDown();
                    try {
                        lockHeld.await();
                    } catch (InterruptedException e) {
                        Thread.currentThread().interrupt();
                    }
                });
            } catch (StateLockException e) {
                fail("First thread should acquire lock");
            }
        });

        lockAcquired.await();

        assertThrows(StateLockException.class, () ->
            locker.withLock(() -> "should timeout")
        );

        lockHeld.countDown();
        holder.join(1000);
    }

    @Test
    void shouldHandleConcurrentAccessSerially() throws Exception {
        StateLocker locker = new StateLocker(Duration.ofSeconds(5));
        AtomicInteger counter = new AtomicInteger(0);
        int threadCount = 100;
        CountDownLatch latch = new CountDownLatch(threadCount);
        ExecutorService executor = Executors.newVirtualThreadPerTaskExecutor();

        for (int i = 0; i < threadCount; i++) {
            executor.submit(() -> {
                try {
                    locker.withLockVoid(() -> {
                        int current = counter.get();
                        Thread.yield();
                        counter.set(current + 1);
                    });
                } catch (StateLockException e) {
                    fail("Should not timeout with sufficient timeout");
                } finally {
                    latch.countDown();
                }
            });
        }

        assertTrue(latch.await(10, TimeUnit.SECONDS));
        assertEquals(threadCount, counter.get());
        
        executor.shutdown();
        assertTrue(executor.awaitTermination(5, TimeUnit.SECONDS));
    }

    @Test
    void shouldNotAllowNestedLocksFromSameThread() {
        StateLocker locker = new StateLocker(Duration.ofMillis(100));

        assertThrows(StateLockException.class, () ->
            locker.withLock(() ->
                locker.withLock(() -> "nested")
            )
        );
    }

    @ParameterizedTest
    @MethodSource("invalidTimeouts")
    void shouldRejectInvalidTimeouts(Duration timeout) {
        assertThrows(IllegalArgumentException.class, () ->
            new StateLocker(timeout)
        );
    }

    static Stream<Duration> invalidTimeouts() {
        return Stream.of(
            null,
            Duration.ZERO,
            Duration.ofSeconds(-1),
            Duration.ofMillis(-100)
        );
    }

    @Test
    void shouldPreserveInterruptedStatus() throws Exception {
        StateLocker locker = new StateLocker(Duration.ofSeconds(1));
        CountDownLatch lockAcquired = new CountDownLatch(1);
        CountDownLatch testComplete = new CountDownLatch(1);

        Thread testThread = Thread.startVirtualThread(() -> {
            try {
                locker.withLockVoid(() -> {
                    lockAcquired.countDown();
                    try {
                        Thread.sleep(10000);
                    } catch (InterruptedException e) {
                        Thread.currentThread().interrupt();
                    }
                });
            } catch (StateLockException e) {
                fail("Should not throw StateLockException");
            }
            
            assertTrue(Thread.currentThread().isInterrupted());
            testComplete.countDown();
        });

        lockAcquired.await();
        testThread.interrupt();
        
        assertTrue(testComplete.await(2, TimeUnit.SECONDS));
    }

    @Test
    void shouldThrowStateLockExceptionWithMessage() {
        StateLocker locker = new StateLocker(Duration.ofMillis(50));
        
        try {
            locker.withLockVoid(() -> {
                try {
                    locker.withLock(() -> "will timeout");
                } catch (StateLockException e) {
                    assertNotNull(e.getMessage());
                    assertTrue(e.getMessage().contains("Failed to acquire state lock"));
                    throw e;
                }
            });
            fail("Should throw StateLockException");
        } catch (StateLockException e) {
            assertNotNull(e.getMessage());
        }
    }

    @Test
    void shouldWorkWithVirtualThreads() throws Exception {
        StateLocker locker = new StateLocker(Duration.ofSeconds(2));
        int iterations = 1000;
        CountDownLatch latch = new CountDownLatch(iterations);
        AtomicInteger successCount = new AtomicInteger(0);

        try (ExecutorService executor = Executors.newVirtualThreadPerTaskExecutor()) {
            for (int i = 0; i < iterations; i++) {
                executor.submit(() -> {
                    try {
                        locker.withLockVoid(() -> {
                            successCount.incrementAndGet();
                        });
                    } catch (StateLockException e) {
                        fail("Virtual thread should acquire lock");
                    } finally {
                        latch.countDown();
                    }
                });
            }

            assertTrue(latch.await(30, TimeUnit.SECONDS));
            assertEquals(iterations, successCount.get());
        }
    }

    @Test
    void shouldHandleRapidAcquireReleaseSequence() throws StateLockException {
        StateLocker locker = new StateLocker(Duration.ofSeconds(1));

        for (int i = 0; i < 1000; i++) {
            final int iteration = i;
            int result = locker.withLock(() -> iteration);
            assertEquals(i, result);
        }
    }

    @Test
    void shouldHandleInterruptDuringLockAcquisition() throws Exception {
        StateLocker locker = new StateLocker(Duration.ofSeconds(10));
        CountDownLatch lockAcquired = new CountDownLatch(1);
        CountDownLatch attemptingLock = new CountDownLatch(1);
        CountDownLatch testDone = new CountDownLatch(1);

        Thread holder = Thread.startVirtualThread(() -> {
            try {
                locker.withLockVoid(() -> {
                    lockAcquired.countDown();
                    try {
                        Thread.sleep(5000);
                    } catch (InterruptedException e) {
                        Thread.currentThread().interrupt();
                    }
                });
            } catch (StateLockException e) {
                fail("Holder should not fail");
            }
        });

        lockAcquired.await();

        Thread waiter = Thread.startVirtualThread(() -> {
            try {
                attemptingLock.countDown();
                locker.withLock(() -> "should not get here");
                fail("Should throw StateLockException due to interrupt");
            } catch (StateLockException e) {
                assertTrue(Thread.currentThread().isInterrupted());
                testDone.countDown();
            }
        });

        attemptingLock.await();
        Thread.sleep(100);
        waiter.interrupt();

        assertTrue(testDone.await(2, TimeUnit.SECONDS));
        holder.interrupt();
        holder.join(1000);
    }
}
