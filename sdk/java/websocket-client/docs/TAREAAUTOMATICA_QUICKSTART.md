# Quick Start: TareaAutomatica

## Installation

The `TareaAutomatica` system is built into the WebSocket SDK. No additional dependencies required.

## 30-Second Example

```java
import tech.hellsoft.trading.ConectorBolsa;
import tech.hellsoft.trading.tasks.TareaAutomatica;
import java.time.Duration;

public class QuickStart {
    public static void main(String[] args) {
        // 1. Create and connect
        ConectorBolsa connector = new ConectorBolsa();
        connector.conectar("wss://market.example.com/ws", "your-token");
        
        // 2. Create a task
        TareaAutomatica monitorTask = new TareaAutomatica("my-monitor") {
            @Override
            protected void ejecutar() {
                System.out.println("Monitoring market...");
                // Your trading logic here
            }
            
            @Override
            protected Duration intervalo() {
                return Duration.ofSeconds(5); // Run every 5 seconds
            }
        };
        
        // 3. Register the task (starts automatically)
        connector.registrarTarea(monitorTask);
        
        // Task runs in background...
        
        // 4. Stop when done
        connector.detenerTarea("my-monitor");
        connector.shutdown();
    }
}
```

## Common Patterns

### Pattern 1: Price Monitor

```java
public class PriceMonitor extends TareaAutomatica {
    private final ConectorBolsa connector;
    private final String product;
    
    public PriceMonitor(ConectorBolsa connector, String product) {
        super("monitor-" + product);
        this.connector = connector;
        this.product = product;
    }
    
    @Override
    protected void ejecutar() {
        // Check prices and send orders
    }
    
    @Override
    protected Duration intervalo() {
        return Duration.ofSeconds(1);
    }
}

// Usage
connector.registrarTarea(new PriceMonitor(connector, "GUACA"));
```

### Pattern 2: Trading Strategy

```java
public class SimpleStrategy extends TareaAutomatica {
    private final ConectorBolsa connector;
    
    public SimpleStrategy(ConectorBolsa connector) {
        super("simple-strategy");
        this.connector = connector;
    }
    
    @Override
    protected void ejecutar() {
        if (shouldBuy()) {
            connector.enviarOrden(createBuyOrder());
        }
        if (shouldSell()) {
            connector.enviarOrden(createSellOrder());
        }
    }
    
    @Override
    protected Duration intervalo() {
        return Duration.ofMillis(100); // Fast execution
    }
    
    private boolean shouldBuy() {
        // Your buy logic
        return false;
    }
    
    private boolean shouldSell() {
        // Your sell logic
        return false;
    }
    
    private OrderMessage createBuyOrder() {
        return OrderMessage.builder()
            .clOrdID("buy-" + System.currentTimeMillis())
            .side(OrderSide.BUY)
            .product(Product.GUACA)
            .qty(100)
            .mode(OrderMode.MARKET)
            .build();
    }
    
    private OrderMessage createSellOrder() {
        return OrderMessage.builder()
            .clOrdID("sell-" + System.currentTimeMillis())
            .side(OrderSide.SELL)
            .product(Product.GUACA)
            .qty(100)
            .mode(OrderMode.MARKET)
            .build();
    }
}

// Usage
connector.registrarTarea(new SimpleStrategy(connector));
```

### Pattern 3: Continuous High-Frequency

```java
public class HighFrequencyTask extends TareaAutomatica {
    private final ConectorBolsa connector;
    
    public HighFrequencyTask(ConectorBolsa connector) {
        super("hft-task");
        this.connector = connector;
    }
    
    @Override
    protected void ejecutar() {
        // Ultra-fast execution
        processMarketData();
    }
    
    @Override
    protected boolean ejecucionContinua() {
        return true; // No delay between executions
    }
}

// Usage
connector.registrarTarea(new HighFrequencyTask(connector));
```

## Configuration Options

### Execution Intervals

```java
// Milliseconds
Duration.ofMillis(100)   // 100ms - Very fast
Duration.ofMillis(500)   // 500ms - Fast

// Seconds
Duration.ofSeconds(1)    // 1 second - Default speed
Duration.ofSeconds(5)    // 5 seconds - Normal
Duration.ofSeconds(30)   // 30 seconds - Slow

// Minutes
Duration.ofMinutes(1)    // 1 minute - Very slow
Duration.ofMinutes(5)    // 5 minutes - Periodic

// Continuous (no delay)
@Override
protected boolean ejecucionContinua() {
    return true;
}
```

## Lifecycle Management

```java
// Start a task
connector.registrarTarea(task);

// Stop a specific task
connector.detenerTarea("task-key");

// Stop all tasks and cleanup
connector.shutdown();
```

## Error Handling

```java
@Override
protected void ejecutar() {
    try {
        riskyOperation();
    } catch (Exception e) {
        // Log and continue
        log.error("Task error: {}", e.getMessage());
        // Task will continue on next interval
    }
}
```

## Cleanup Resources

```java
private DatabaseConnection db;

@Override
protected void onDetener() {
    // Called when task stops
    if (db != null) {
        db.close();
    }
}
```

## Multiple Tasks

```java
// Run multiple tasks concurrently
connector.registrarTarea(new PriceMonitor(connector, "GUACA"));
connector.registrarTarea(new PriceMonitor(connector, "CAFE"));
connector.registrarTarea(new SimpleStrategy(connector));
connector.registrarTarea(new RiskManager(connector));

// All tasks run independently
```

## Best Practices

✅ **DO**:
- Use descriptive task keys: `"monitor-GUACA"`, `"strategy-arbitrage"`
- Keep `ejecutar()` fast and non-blocking
- Handle exceptions gracefully
- Clean up resources in `onDetener()`
- Use appropriate intervals for your use case

❌ **DON'T**:
- Don't use generic keys: `"task1"`, `"task"`
- Don't perform long-running operations in `ejecutar()`
- Don't ignore exceptions (log them)
- Don't use continuous mode unless necessary
- Don't forget to call `shutdown()` on exit

## Full Example

See [AUTOMATIC_TASKS.md](./AUTOMATIC_TASKS.md) for complete documentation and advanced examples.

## Need Help?

- Check the [documentation](./AUTOMATIC_TASKS.md)
- Review the [test examples](../src/test/java/tech/hellsoft/trading/tasks/)
- See the [example implementation](../src/main/java/tech/hellsoft/trading/tasks/MonitorPreciosEjemplo.java)

