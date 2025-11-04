package tech.hellsoft.trading.config;

import lombok.Builder;
import lombok.Data;
import lombok.extern.slf4j.Slf4j;

import java.time.Duration;

@Data
@Builder
@Slf4j
public class ConectorConfig {
    
    @Builder.Default
    private Duration heartbeatInterval = Duration.ofSeconds(30);
    
    @Builder.Default
    private Duration connectionTimeout = Duration.ofSeconds(10);
    
    @Builder.Default
    private boolean autoReconnect = true;
    
    @Builder.Default
    private int maxReconnectAttempts = 5;
    
    @Builder.Default
    private Duration reconnectInitialDelay = Duration.ofSeconds(1);
    
    @Builder.Default
    private Duration reconnectMaxDelay = Duration.ofSeconds(30);
    
    @Builder.Default
    private double reconnectBackoffMultiplier = 2.0;
    
    @Builder.Default
    private boolean enableMessageSequencing = true;
    
    @Builder.Default
    private Duration messageSequencingTimeout = Duration.ofSeconds(30);
    
    @Builder.Default
    private boolean enableStateLocking = true;
    
    @Builder.Default
    private Duration stateLockTimeout = Duration.ofSeconds(5);
    
    @Builder.Default
    private boolean autoResyncOnReconnect = true;
    
    @Builder.Default
    private Duration resyncLookback = Duration.ofMinutes(5);
    
    public static ConectorConfig defaultConfig() {
        return ConectorConfig.builder().build();
    }
    
    public void validate() {
        if (heartbeatInterval.isNegative() || heartbeatInterval.isZero()) {
            throw new IllegalArgumentException("Heartbeat interval must be positive");
        }
        
        if (connectionTimeout.isNegative() || connectionTimeout.isZero()) {
            throw new IllegalArgumentException("Connection timeout must be positive");
        }
        
        if (maxReconnectAttempts < -1 || maxReconnectAttempts == 0) {
            throw new IllegalArgumentException("Max reconnect attempts must be -1 or positive");
        }
        
        if (reconnectInitialDelay.isNegative()) {
            throw new IllegalArgumentException("Reconnect initial delay cannot be negative");
        }
        
        if (reconnectMaxDelay.compareTo(reconnectInitialDelay) < 0) {
            throw new IllegalArgumentException("Reconnect max delay must be >= initial delay");
        }
        
        if (reconnectBackoffMultiplier < 1.0) {
            throw new IllegalArgumentException("Reconnect backoff multiplier must be >= 1.0");
        }
        
        if (stateLockTimeout.isNegative() || stateLockTimeout.isZero()) {
            throw new IllegalArgumentException("State lock timeout must be positive");
        }
        
        if (messageSequencingTimeout.isNegative() || messageSequencingTimeout.isZero()) {
            throw new IllegalArgumentException("Message sequencing timeout must be positive");
        }
        
        if (resyncLookback.isNegative()) {
            throw new IllegalArgumentException("Resync lookback cannot be negative");
        }
        
        log.debug("Configuration validated successfully");
    }
}
