package tech.hellsoft.trading.config;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.ValueSource;

import java.time.Duration;

import static org.junit.jupiter.api.Assertions.*;

class ConectorConfigTest {

    @Test
    void shouldCreateDefaultConfig() {
        ConectorConfig config = ConectorConfig.defaultConfig();
        
        assertNotNull(config);
        assertEquals(Duration.ofSeconds(30), config.getHeartbeatInterval());
        assertEquals(Duration.ofSeconds(10), config.getConnectionTimeout());
        assertTrue(config.isAutoReconnect());
        assertEquals(5, config.getMaxReconnectAttempts());
    }

    @Test
    void shouldCreateConfigWithBuilder() {
        ConectorConfig config = ConectorConfig.builder()
            .heartbeatInterval(Duration.ofSeconds(15))
            .connectionTimeout(Duration.ofSeconds(5))
            .autoReconnect(false)
            .maxReconnectAttempts(3)
            .build();

        assertEquals(Duration.ofSeconds(15), config.getHeartbeatInterval());
        assertEquals(Duration.ofSeconds(5), config.getConnectionTimeout());
        assertFalse(config.isAutoReconnect());
        assertEquals(3, config.getMaxReconnectAttempts());
    }

    @Test
    void shouldHaveCorrectDefaults() {
        ConectorConfig config = ConectorConfig.defaultConfig();
        
        assertEquals(Duration.ofSeconds(30), config.getHeartbeatInterval());
        assertEquals(Duration.ofSeconds(10), config.getConnectionTimeout());
        assertTrue(config.isAutoReconnect());
        assertEquals(5, config.getMaxReconnectAttempts());
        assertEquals(Duration.ofSeconds(1), config.getReconnectInitialDelay());
        assertEquals(Duration.ofSeconds(30), config.getReconnectMaxDelay());
        assertEquals(2.0, config.getReconnectBackoffMultiplier());
        assertTrue(config.isEnableMessageSequencing());
        assertEquals(Duration.ofSeconds(30), config.getMessageSequencingTimeout());
        assertTrue(config.isEnableStateLocking());
        assertEquals(Duration.ofSeconds(5), config.getStateLockTimeout());
        assertTrue(config.isAutoResyncOnReconnect());
        assertEquals(Duration.ofMinutes(5), config.getResyncLookback());
    }

    @Test
    void shouldValidateSuccessfully() {
        ConectorConfig config = ConectorConfig.defaultConfig();
        assertDoesNotThrow(() -> config.validate());
    }

    @Test
    void shouldRejectZeroMaxReconnectAttempts() {
        ConectorConfig config = ConectorConfig.defaultConfig();
        config.setMaxReconnectAttempts(0);
        
        assertThrows(IllegalArgumentException.class, () -> config.validate());
    }

    @Test
    void shouldRejectInvalidMaxReconnectAttempts() {
        ConectorConfig config = ConectorConfig.defaultConfig();
        config.setMaxReconnectAttempts(-2);
        
        assertThrows(IllegalArgumentException.class, () -> config.validate());
    }

    @Test
    void shouldAllowUnlimitedReconnects() {
        ConectorConfig config = ConectorConfig.defaultConfig();
        config.setMaxReconnectAttempts(-1);
        
        assertDoesNotThrow(() -> config.validate());
    }

    @Test
    void shouldRejectNegativeHeartbeatInterval() {
        ConectorConfig config = ConectorConfig.defaultConfig();
        config.setHeartbeatInterval(Duration.ofSeconds(-1));
        
        assertThrows(IllegalArgumentException.class, () -> config.validate());
    }

    @Test
    void shouldRejectZeroHeartbeatInterval() {
        ConectorConfig config = ConectorConfig.defaultConfig();
        config.setHeartbeatInterval(Duration.ZERO);
        
        assertThrows(IllegalArgumentException.class, () -> config.validate());
    }

    @Test
    void shouldRejectNegativeConnectionTimeout() {
        ConectorConfig config = ConectorConfig.defaultConfig();
        config.setConnectionTimeout(Duration.ofSeconds(-1));
        
        assertThrows(IllegalArgumentException.class, () -> config.validate());
    }

    @Test
    void shouldRejectZeroConnectionTimeout() {
        ConectorConfig config = ConectorConfig.defaultConfig();
        config.setConnectionTimeout(Duration.ZERO);
        
        assertThrows(IllegalArgumentException.class, () -> config.validate());
    }

    @Test
    void shouldRejectNegativeReconnectInitialDelay() {
        ConectorConfig config = ConectorConfig.defaultConfig();
        config.setReconnectInitialDelay(Duration.ofSeconds(-1));
        
        assertThrows(IllegalArgumentException.class, () -> config.validate());
    }

    @Test
    void shouldRejectReconnectMaxDelayLessThanInitial() {
        ConectorConfig config = ConectorConfig.defaultConfig();
        config.setReconnectInitialDelay(Duration.ofSeconds(10));
        config.setReconnectMaxDelay(Duration.ofSeconds(5));
        
        assertThrows(IllegalArgumentException.class, () -> config.validate());
    }

    @Test
    void shouldAllowReconnectMaxDelayEqualToInitial() {
        ConectorConfig config = ConectorConfig.defaultConfig();
        config.setReconnectInitialDelay(Duration.ofSeconds(10));
        config.setReconnectMaxDelay(Duration.ofSeconds(10));
        
        assertDoesNotThrow(() -> config.validate());
    }

    @Test
    void shouldRejectBackoffMultiplierLessThanOne() {
        ConectorConfig config = ConectorConfig.defaultConfig();
        config.setReconnectBackoffMultiplier(0.5);
        
        assertThrows(IllegalArgumentException.class, () -> config.validate());
    }

    @Test
    void shouldAllowBackoffMultiplierOfOne() {
        ConectorConfig config = ConectorConfig.defaultConfig();
        config.setReconnectBackoffMultiplier(1.0);
        
        assertDoesNotThrow(() -> config.validate());
    }

    @Test
    void shouldRejectNegativeStateLockTimeout() {
        ConectorConfig config = ConectorConfig.defaultConfig();
        config.setStateLockTimeout(Duration.ofSeconds(-1));
        
        assertThrows(IllegalArgumentException.class, () -> config.validate());
    }

    @Test
    void shouldRejectZeroStateLockTimeout() {
        ConectorConfig config = ConectorConfig.defaultConfig();
        config.setStateLockTimeout(Duration.ZERO);
        
        assertThrows(IllegalArgumentException.class, () -> config.validate());
    }

    @Test
    void shouldRejectNegativeMessageSequencingTimeout() {
        ConectorConfig config = ConectorConfig.defaultConfig();
        config.setMessageSequencingTimeout(Duration.ofSeconds(-1));
        
        assertThrows(IllegalArgumentException.class, () -> config.validate());
    }

    @Test
    void shouldRejectZeroMessageSequencingTimeout() {
        ConectorConfig config = ConectorConfig.defaultConfig();
        config.setMessageSequencingTimeout(Duration.ZERO);
        
        assertThrows(IllegalArgumentException.class, () -> config.validate());
    }

    @Test
    void shouldRejectNegativeResyncLookback() {
        ConectorConfig config = ConectorConfig.defaultConfig();
        config.setResyncLookback(Duration.ofMinutes(-1));
        
        assertThrows(IllegalArgumentException.class, () -> config.validate());
    }

    @Test
    void shouldAllowZeroResyncLookback() {
        ConectorConfig config = ConectorConfig.defaultConfig();
        config.setResyncLookback(Duration.ZERO);
        
        assertDoesNotThrow(() -> config.validate());
    }

    @Test
    void shouldAllowDisablingAutoReconnect() {
        ConectorConfig config = ConectorConfig.builder()
            .autoReconnect(false)
            .build();

        assertFalse(config.isAutoReconnect());
        assertDoesNotThrow(() -> config.validate());
    }

    @Test
    void shouldAllowDisablingMessageSequencing() {
        ConectorConfig config = ConectorConfig.builder()
            .enableMessageSequencing(false)
            .build();

        assertFalse(config.isEnableMessageSequencing());
        assertDoesNotThrow(() -> config.validate());
    }

    @Test
    void shouldAllowDisablingStateLocking() {
        ConectorConfig config = ConectorConfig.builder()
            .enableStateLocking(false)
            .build();

        assertFalse(config.isEnableStateLocking());
        assertDoesNotThrow(() -> config.validate());
    }

    @Test
    void shouldAllowDisablingAutoResync() {
        ConectorConfig config = ConectorConfig.builder()
            .autoResyncOnReconnect(false)
            .build();

        assertFalse(config.isAutoResyncOnReconnect());
        assertDoesNotThrow(() -> config.validate());
    }

    @Test
    void shouldValidateComplexConfiguration() {
        ConectorConfig config = ConectorConfig.builder()
            .heartbeatInterval(Duration.ofSeconds(20))
            .connectionTimeout(Duration.ofSeconds(15))
            .autoReconnect(true)
            .maxReconnectAttempts(10)
            .reconnectInitialDelay(Duration.ofMillis(500))
            .reconnectMaxDelay(Duration.ofSeconds(60))
            .reconnectBackoffMultiplier(1.5)
            .enableMessageSequencing(true)
            .messageSequencingTimeout(Duration.ofSeconds(45))
            .enableStateLocking(true)
            .stateLockTimeout(Duration.ofSeconds(3))
            .autoResyncOnReconnect(true)
            .resyncLookback(Duration.ofMinutes(10))
            .build();

        assertDoesNotThrow(() -> config.validate());
    }

    @Test
    void shouldHandleExtremeValues() {
        ConectorConfig config = ConectorConfig.builder()
            .heartbeatInterval(Duration.ofMillis(1))
            .connectionTimeout(Duration.ofMillis(1))
            .maxReconnectAttempts(1000)
            .reconnectInitialDelay(Duration.ZERO)
            .reconnectMaxDelay(Duration.ofHours(24))
            .reconnectBackoffMultiplier(10.0)
            .stateLockTimeout(Duration.ofMillis(1))
            .messageSequencingTimeout(Duration.ofMillis(1))
            .resyncLookback(Duration.ofHours(48))
            .build();

        assertDoesNotThrow(() -> config.validate());
    }
}
