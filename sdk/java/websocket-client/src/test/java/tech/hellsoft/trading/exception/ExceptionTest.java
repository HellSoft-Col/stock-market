package tech.hellsoft.trading.exception;

import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.*;

class ExceptionTest {

    @Test
    void shouldCreateConexionFallidaException() {
        ConexionFallidaException ex = new ConexionFallidaException(
            "Connection failed", 
            "localhost", 
            8080
        );
        
        assertNotNull(ex);
        assertEquals("localhost", ex.getHost());
        assertEquals(8080, ex.getPort());
        assertTrue(ex.getMessage().contains("Connection failed"));
    }

    @Test
    void shouldCreateConexionFallidaExceptionWithCause() {
        Throwable cause = new RuntimeException("Network error");
        ConexionFallidaException ex = new ConexionFallidaException(
            "Connection failed",
            "localhost",
            8080,
            cause
        );
        
        assertNotNull(ex);
        assertEquals(cause, ex.getCause());
        assertEquals("localhost", ex.getHost());
        assertEquals(8080, ex.getPort());
    }

    @Test
    void shouldCreateValidationException() {
        ValidationException ex = new ValidationException("Invalid input");
        
        assertNotNull(ex);
        assertEquals("Invalid input", ex.getMessage());
    }

    @Test
    void shouldCreateValidationExceptionWithCause() {
        Throwable cause = new IllegalArgumentException("Bad value");
        ValidationException ex = new ValidationException("Validation failed", cause);
        
        assertNotNull(ex);
        assertEquals(cause, ex.getCause());
        assertTrue(ex.getMessage().contains("Validation failed"));
    }

    @Test
    void shouldCreateStateLockExceptionWithMessage() {
        StateLockException ex = new StateLockException("Lock failed");
        
        assertNotNull(ex);
        assertEquals("Lock failed", ex.getMessage());
        assertNull(ex.getActionName());
        assertEquals(0, ex.getTimeoutMillis());
    }

    @Test
    void shouldCreateStateLockExceptionWithDetails() {
        StateLockException ex = new StateLockException("sendMessage", 5000);
        
        assertNotNull(ex);
        assertEquals("sendMessage", ex.getActionName());
        assertEquals(5000, ex.getTimeoutMillis());
        assertTrue(ex.getMessage().contains("sendMessage"));
        assertTrue(ex.getMessage().contains("5000"));
    }

    @Test
    void shouldHaveSerialVersionUID() {
        assertDoesNotThrow(() -> {
            ConexionFallidaException.class.getDeclaredField("serialVersionUID");
            ValidationException.class.getDeclaredField("serialVersionUID");
            StateLockException.class.getDeclaredField("serialVersionUID");
        });
    }
}
