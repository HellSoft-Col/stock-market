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
    void shouldValidateSuccessfully() {
        ConectorConfig config = ConectorConfig.defaultConfig();
        assertDoesNotThrow(() -> config.validate());
    }

    @ParameterizedTest
    @ValueSource(ints = {-1, 0})
    void shouldRejectInvalidMaxReconnectAttempts(int attempts) {
        ConectorConfig config = ConectorConfig.defaultConfig();
        config.setMaxReconnectAttempts(attempts);
        
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
}
