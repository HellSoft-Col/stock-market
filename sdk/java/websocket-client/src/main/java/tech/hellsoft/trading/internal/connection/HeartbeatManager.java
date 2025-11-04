package tech.hellsoft.trading.internal.connection;

import lombok.extern.slf4j.Slf4j;
import tech.hellsoft.trading.dto.client.PingMessage;
import tech.hellsoft.trading.enums.MessageType;

import java.time.Duration;
import java.time.Instant;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;
import java.util.function.Consumer;

@Slf4j
public class HeartbeatManager {
    private final Duration pingInterval;
    private final Duration pongTimeout;
    private final Consumer<PingMessage> sendPing;
    private final Runnable onPongTimeout;
    
    private ScheduledExecutorService scheduler;
    private volatile Instant lastPongReceived;
    private volatile boolean running;

    public HeartbeatManager(Duration pingInterval, 
                           Duration pongTimeout,
                           Consumer<PingMessage> sendPing,
                           Runnable onPongTimeout) {
        if (pingInterval == null || pingInterval.isNegative() || pingInterval.isZero()) {
            throw new IllegalArgumentException("pingInterval must be positive");
        }
        
        if (pongTimeout == null || pongTimeout.isNegative() || pongTimeout.isZero()) {
            throw new IllegalArgumentException("pongTimeout must be positive");
        }
        
        if (sendPing == null) {
            throw new IllegalArgumentException("sendPing cannot be null");
        }
        
        if (onPongTimeout == null) {
            throw new IllegalArgumentException("onPongTimeout cannot be null");
        }
        
        this.pingInterval = pingInterval;
        this.pongTimeout = pongTimeout;
        this.sendPing = sendPing;
        this.onPongTimeout = onPongTimeout;
        this.lastPongReceived = Instant.now();
    }

    public void start() {
        if (running) {
            log.warn("Heartbeat manager already running");
            return;
        }

        running = true;
        lastPongReceived = Instant.now();
        
        scheduler = Executors.newScheduledThreadPool(1, Thread.ofVirtual().factory());
        
        scheduler.scheduleAtFixedRate(
            this::sendHeartbeat,
            pingInterval.toMillis(),
            pingInterval.toMillis(),
            TimeUnit.MILLISECONDS
        );
        
        log.debug("Heartbeat manager started");
    }

    public void stop() {
        if (!running) {
            return;
        }

        running = false;
        
        if (scheduler != null) {
            scheduler.shutdown();
            scheduler = null;
        }
        
        log.debug("Heartbeat manager stopped");
    }

    public void onPongReceived() {
        lastPongReceived = Instant.now();
        log.trace("Pong received at {}", lastPongReceived);
    }

    private void sendHeartbeat() {
        if (!running) {
            return;
        }

        try {
            checkPongTimeout();
            
            Instant now = Instant.now();
            PingMessage ping = PingMessage.builder()
                .type(MessageType.PING)
                .timestamp(now.toString())
                .build();
            
            sendPing.accept(ping);
            log.trace("Ping sent at {}", now);
            
        } catch (Exception e) {
            log.error("Error in heartbeat cycle", e);
        }
    }

    private void checkPongTimeout() {
        Instant now = Instant.now();
        Duration timeSinceLastPong = Duration.between(lastPongReceived, now);
        
        if (timeSinceLastPong.compareTo(pongTimeout) > 0) {
            log.warn("Pong timeout: {} since last pong (timeout: {})", 
                timeSinceLastPong, pongTimeout);
            
            stop();
            onPongTimeout.run();
        }
    }

    public boolean isRunning() {
        return running;
    }
}
