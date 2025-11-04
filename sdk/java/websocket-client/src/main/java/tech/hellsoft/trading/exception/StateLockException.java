package tech.hellsoft.trading.exception;

import lombok.Getter;

@Getter
public class StateLockException extends RuntimeException {
    private final String actionName;
    private final long timeoutMillis;
    
    public StateLockException(String message) {
        super(message);
        this.actionName = null;
        this.timeoutMillis = 0;
    }
    
    public StateLockException(String actionName, long timeoutMillis) {
        super(String.format("Could not acquire lock for action '%s' within %d ms", 
            actionName, timeoutMillis));
        this.actionName = actionName;
        this.timeoutMillis = timeoutMillis;
    }
}
