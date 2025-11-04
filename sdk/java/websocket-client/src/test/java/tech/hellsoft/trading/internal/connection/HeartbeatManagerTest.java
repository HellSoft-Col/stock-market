package tech.hellsoft.trading.internal.connection;

import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.MethodSource;
import tech.hellsoft.trading.dto.client.PingMessage;
import tech.hellsoft.trading.enums.MessageType;

import java.time.Duration;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.concurrent.atomic.AtomicReference;
import java.util.function.Consumer;
import java.util.stream.Stream;

import static org.junit.jupiter.api.Assertions.*;

class HeartbeatManagerTest {

    private HeartbeatManager manager;
    private AtomicInteger pingCount;
    private AtomicInteger timeoutCount;
    private Consumer<PingMessage> sendPing;
    private Runnable onTimeout;

    @BeforeEach
    void setUp() {
        pingCount = new AtomicInteger(0);
        timeoutCount = new AtomicInteger(0);
        sendPing = msg -> pingCount.incrementAndGet();
        onTimeout = () -> timeoutCount.incrementAndGet();
    }

    @AfterEach
    void tearDown() {
        if (manager != null && manager.isRunning()) {
            manager.stop();
        }
    }

    @Test
    void shouldStartHeartbeatManager() {
        manager = new HeartbeatManager(
            Duration.ofMillis(100),
            Duration.ofSeconds(5),
            sendPing,
            onTimeout
        );

        manager.start();

        assertTrue(manager.isRunning());
    }

    @Test
    void shouldStopHeartbeatManager() {
        manager = new HeartbeatManager(
            Duration.ofMillis(100),
            Duration.ofSeconds(5),
            sendPing,
            onTimeout
        );

        manager.start();
        manager.stop();

        assertFalse(manager.isRunning());
    }

    @Test
    void shouldSendPeriodicPings() throws InterruptedException {
        manager = new HeartbeatManager(
            Duration.ofMillis(100),
            Duration.ofSeconds(5),
            sendPing,
            onTimeout
        );

        manager.start();
        Thread.sleep(350);
        manager.stop();

        assertTrue(pingCount.get() >= 2, "Expected at least 2 pings, got " + pingCount.get());
    }

    @Test
    void shouldIncludePingTypeInMessage() throws InterruptedException {
        AtomicReference<PingMessage> capturedPing = new AtomicReference<>();
        CountDownLatch latch = new CountDownLatch(1);

        manager = new HeartbeatManager(
            Duration.ofMillis(50),
            Duration.ofSeconds(5),
            msg -> {
                capturedPing.set(msg);
                latch.countDown();
            },
            onTimeout
        );

        manager.start();
        assertTrue(latch.await(500, TimeUnit.MILLISECONDS));
        manager.stop();

        assertNotNull(capturedPing.get());
        assertEquals(MessageType.PING, capturedPing.get().getType());
        assertNotNull(capturedPing.get().getTimestamp());
    }

    @Test
    void shouldDetectPongTimeout() throws InterruptedException {
        CountDownLatch timeoutLatch = new CountDownLatch(1);

        manager = new HeartbeatManager(
            Duration.ofMillis(50),
            Duration.ofMillis(150),
            sendPing,
            () -> {
                timeoutCount.incrementAndGet();
                timeoutLatch.countDown();
            }
        );

        manager.start();
        
        assertTrue(timeoutLatch.await(500, TimeUnit.MILLISECONDS));
        assertEquals(1, timeoutCount.get());
        assertFalse(manager.isRunning());
    }

    @Test
    void shouldResetTimeoutWhenPongReceived() throws InterruptedException {
        manager = new HeartbeatManager(
            Duration.ofMillis(50),
            Duration.ofMillis(200),
            sendPing,
            onTimeout
        );

        manager.start();
        
        Thread.sleep(100);
        manager.onPongReceived();
        
        Thread.sleep(100);
        manager.onPongReceived();
        
        Thread.sleep(100);
        manager.stop();

        assertEquals(0, timeoutCount.get());
    }

    @Test
    void shouldNotStartTwice() {
        manager = new HeartbeatManager(
            Duration.ofMillis(100),
            Duration.ofSeconds(5),
            sendPing,
            onTimeout
        );

        manager.start();
        manager.start();

        assertTrue(manager.isRunning());
    }

    @Test
    void shouldHandleStopWhenNotRunning() {
        manager = new HeartbeatManager(
            Duration.ofMillis(100),
            Duration.ofSeconds(5),
            sendPing,
            onTimeout
        );

        assertDoesNotThrow(() -> manager.stop());
        assertFalse(manager.isRunning());
    }

    @Test
    void shouldStopSendingPingsAfterStop() throws InterruptedException {
        manager = new HeartbeatManager(
            Duration.ofMillis(50),
            Duration.ofSeconds(5),
            sendPing,
            onTimeout
        );

        manager.start();
        Thread.sleep(150);
        int countAtStop = pingCount.get();
        manager.stop();
        
        Thread.sleep(150);
        int countAfterStop = pingCount.get();

        assertEquals(countAtStop, countAfterStop, "Ping count should not increase after stop");
    }

    @ParameterizedTest
    @MethodSource("invalidDurations")
    void shouldRejectInvalidPingInterval(Duration interval) {
        assertThrows(IllegalArgumentException.class, () ->
            new HeartbeatManager(interval, Duration.ofSeconds(5), sendPing, onTimeout)
        );
    }

    @ParameterizedTest
    @MethodSource("invalidDurations")
    void shouldRejectInvalidPongTimeout(Duration timeout) {
        assertThrows(IllegalArgumentException.class, () ->
            new HeartbeatManager(Duration.ofSeconds(1), timeout, sendPing, onTimeout)
        );
    }

    static Stream<Duration> invalidDurations() {
        return Stream.of(
            null,
            Duration.ZERO,
            Duration.ofSeconds(-1),
            Duration.ofMillis(-100)
        );
    }

    @Test
    void shouldRejectNullSendPing() {
        assertThrows(IllegalArgumentException.class, () ->
            new HeartbeatManager(Duration.ofSeconds(1), Duration.ofSeconds(5), null, onTimeout)
        );
    }

    @Test
    void shouldRejectNullOnTimeout() {
        assertThrows(IllegalArgumentException.class, () ->
            new HeartbeatManager(Duration.ofSeconds(1), Duration.ofSeconds(5), sendPing, null)
        );
    }

    @Test
    void shouldHandleExceptionInSendPing() throws InterruptedException {
        AtomicInteger exceptionCount = new AtomicInteger(0);
        AtomicInteger successCount = new AtomicInteger(0);

        Consumer<PingMessage> faultySender = msg -> {
            if (exceptionCount.incrementAndGet() <= 2) {
                throw new RuntimeException("Simulated send failure");
            }
            successCount.incrementAndGet();
        };

        manager = new HeartbeatManager(
            Duration.ofMillis(50),
            Duration.ofSeconds(5),
            faultySender,
            onTimeout
        );

        manager.start();
        Thread.sleep(250);
        manager.stop();

        assertTrue(exceptionCount.get() >= 2, "Should have exceptions");
        assertTrue(successCount.get() >= 1, "Should recover and send successful pings");
    }

    @Test
    void shouldHandleExceptionInTimeoutCallback() throws InterruptedException {
        CountDownLatch timeoutCalled = new CountDownLatch(1);

        Runnable faultyTimeout = () -> {
            timeoutCalled.countDown();
            throw new RuntimeException("Timeout callback error");
        };

        manager = new HeartbeatManager(
            Duration.ofMillis(50),
            Duration.ofMillis(150),
            sendPing,
            faultyTimeout
        );

        manager.start();
        
        assertTrue(timeoutCalled.await(500, TimeUnit.MILLISECONDS));
        assertFalse(manager.isRunning());
    }

    @Test
    void shouldUseVirtualThreadForScheduling() throws InterruptedException {
        AtomicReference<Thread> executingThread = new AtomicReference<>();
        CountDownLatch latch = new CountDownLatch(1);

        manager = new HeartbeatManager(
            Duration.ofMillis(50),
            Duration.ofSeconds(5),
            msg -> {
                executingThread.set(Thread.currentThread());
                latch.countDown();
            },
            onTimeout
        );

        manager.start();
        assertTrue(latch.await(200, TimeUnit.MILLISECONDS));
        manager.stop();

        assertNotNull(executingThread.get());
        assertTrue(executingThread.get().isVirtual());
    }

    @Test
    void shouldMaintainCorrectPingInterval() throws InterruptedException {
        AtomicInteger pings = new AtomicInteger(0);
        long startTime = System.nanoTime();
        CountDownLatch latch = new CountDownLatch(5);

        manager = new HeartbeatManager(
            Duration.ofMillis(100),
            Duration.ofSeconds(5),
            msg -> {
                pings.incrementAndGet();
                latch.countDown();
            },
            onTimeout
        );

        manager.start();
        assertTrue(latch.await(1, TimeUnit.SECONDS));
        manager.stop();

        long elapsedMs = TimeUnit.NANOSECONDS.toMillis(System.nanoTime() - startTime);
        
        assertTrue(elapsedMs >= 400, "Should take at least 400ms for 5 pings at 100ms interval");
        assertTrue(elapsedMs < 800, "Should not take more than 800ms");
    }

    @Test
    void shouldHandleMultipleStartStopCycles() throws InterruptedException {
        manager = new HeartbeatManager(
            Duration.ofMillis(50),
            Duration.ofSeconds(5),
            sendPing,
            onTimeout
        );

        for (int i = 0; i < 3; i++) {
            pingCount.set(0);
            
            manager.start();
            assertTrue(manager.isRunning());
            
            Thread.sleep(150);
            assertTrue(pingCount.get() >= 1);
            
            manager.stop();
            assertFalse(manager.isRunning());
            
            Thread.sleep(50);
        }
    }

    @Test
    void shouldHandleRapidPongUpdates() throws InterruptedException {
        manager = new HeartbeatManager(
            Duration.ofMillis(50),
            Duration.ofMillis(200),
            sendPing,
            onTimeout
        );

        manager.start();

        for (int i = 0; i < 10; i++) {
            manager.onPongReceived();
            Thread.sleep(20);
        }

        manager.stop();

        assertEquals(0, timeoutCount.get());
    }

    @Test
    void shouldStopAfterTimeoutDetected() throws InterruptedException {
        CountDownLatch timeoutLatch = new CountDownLatch(1);

        manager = new HeartbeatManager(
            Duration.ofMillis(50),
            Duration.ofMillis(150),
            sendPing,
            () -> {
                timeoutLatch.countDown();
                timeoutCount.incrementAndGet();
            }
        );

        manager.start();
        assertTrue(timeoutLatch.await(500, TimeUnit.MILLISECONDS));
        
        Thread.sleep(100);

        assertFalse(manager.isRunning(), "Should stop running after timeout");
        assertEquals(1, timeoutCount.get());
    }
}
