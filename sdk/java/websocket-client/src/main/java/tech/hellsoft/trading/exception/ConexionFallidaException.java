package tech.hellsoft.trading.exception;

import lombok.Getter;

@Getter
public class ConexionFallidaException extends Exception {
    private static final long serialVersionUID = 1L;
    
    private final String host;
    private final int port;
    
    public ConexionFallidaException(String message, String host, int port) {
        super(message);
        this.host = host;
        this.port = port;
    }
    
    public ConexionFallidaException(String message, String host, int port, Throwable cause) {
        super(message, cause);
        this.host = host;
        this.port = port;
    }
}
