# Automatic Tasks (TareaAutomatica)

## Overview

The `TareaAutomatica` system provides a powerful way to create and manage automatic, recurring tasks that run in the background using virtual threads. This is perfect for implementing trading strategies, monitoring systems, and automated market operations.

## Key Features

- **Virtual Thread Execution**: All tasks run on lightweight virtual threads (Java 21+)
- **Key-Based Locking**: Prevents concurrent execution of tasks with the same key
- **Flexible Scheduling**: Support for interval-based or continuous execution
- **Integrated Management**: Automatically managed by `ConectorBolsa`
- **Graceful Shutdown**: Clean resource cleanup on stop or shutdown

## Basic Usage

### 1. Creating a Simple Task

```java
public class PriceMonitorTask extends TareaAutomatica {
    private final ConectorBolsa connector;
    private final String product;

    public PriceMonitorTask(ConectorBolsa connector, String product) {
        super("monitor-" + product); // Unique key
        this.connector = connector;
        this.product = product;
    }

    @Override
    protected void ejecutar() {
        // Your task logic here
        log.info("Monitoring prices for {}", product);
        
        // You can send orders, check market conditions, etc.
        // if (shouldBuy()) {
        //     connector.enviarOrden(createBuyOrder());
        // }
    }

    @Override
    protected Duration intervalo() {
        return Duration.ofSeconds(5); // Run every 5 seconds
    }
}
```

### 2. Registering Tasks with ConectorBolsa

```java
ConectorBolsa connector = new ConectorBolsa();
connector.conectar("wss://market.example.com/ws", "your-token");

// Register automatic tasks
connector.registrarTarea(new PriceMonitorTask(connector, "GUACA"));
connector.registrarTarea(new PriceMonitorTask(connector, "CAFE"));

// Tasks run automatically in the background
// ...

// Stop a specific task
connector.detenerTarea("monitor-GUACA");

// Clean shutdown stops all tasks
connector.shutdown();
```

## Execution Modes

### Interval-Based Execution (Default)

Tasks execute periodically with a configurable delay between executions:

```java
@Override
protected Duration intervalo() {
    return Duration.ofMillis(100);  // 100ms - fast execution
    // return Duration.ofSeconds(5);   // 5 seconds
    // return Duration.ofMinutes(1);   // 1 minute
}
```

### Continuous Execution

For high-frequency tasks that need to run as fast as possible:

```java
@Override
protected boolean ejecucionContinua() {
    return true; // Run continuously with minimal delay
}

@Override
protected void ejecutar() {
    // This will run in a tight loop
    // Be careful with CPU usage!
}
```

**Warning**: Continuous execution can consume significant CPU resources. Use only when necessary.

## Advanced Features

### Task Cleanup

Override `onDetener()` to perform cleanup when the task stops:

```java
private Connection dbConnection;

@Override
protected void onDetener() {
    log.info("Cleaning up task resources");
    if (dbConnection != null) {
        dbConnection.close();
    }
}
```

### Key-Based Locking

Tasks with the same key cannot run concurrently:

```java
// These two tasks will never execute at the same time
TareaAutomatica task1 = new MyTask("shared-key");
TareaAutomatica task2 = new MyTask("shared-key");

// If you register task2, it will replace task1
connector.registrarTarea(task1);
connector.registrarTarea(task2); // task1 is stopped, task2 starts
```

### Error Handling

Uncaught exceptions in `ejecutar()` are logged but don't stop the task:

```java
@Override
protected void ejecutar() {
    try {
        // Your logic
        riskyOperation();
    } catch (SpecificException e) {
        // Handle specific errors
        log.error("Error in task", e);
    }
    // Uncaught exceptions are automatically logged
    // Task continues running on next interval
}
```

## Complete Example: Arbitrage Strategy

```java
public class ArbitrageStrategy extends TareaAutomatica {
    private final ConectorBolsa connector;
    private final String productA;
    private final String productB;
    private volatile double lastPriceA = 0;
    private volatile double lastPriceB = 0;

    public ArbitrageStrategy(ConectorBolsa connector, String productA, String productB) {
        super("arbitrage-" + productA + "-" + productB);
        this.connector = connector;
        this.productA = productA;
        this.productB = productB;
    }

    @Override
    protected void ejecutar() {
        // Update prices from market data
        updatePrices();
        
        // Calculate spread
        double spread = Math.abs(lastPriceA - lastPriceB);
        
        // Execute arbitrage if spread is favorable
        if (spread > 5.0) { // threshold
            executeArbitrage();
        }
    }

    @Override
    protected Duration intervalo() {
        return Duration.ofMillis(100); // Fast monitoring
    }

    private void updatePrices() {
        // Update from ticker messages via EventListener
        // Store in instance variables
    }

    private void executeArbitrage() {
        log.info("Executing arbitrage: {} vs {}", productA, productB);
        
        // Send buy order for cheaper product
        OrderMessage buyOrder = OrderMessage.builder()
            .clOrdID("arb-buy-" + System.currentTimeMillis())
            .side(OrderSide.BUY)
            .product(getCheaperProduct())
            .qty(100)
            .mode(OrderMode.MARKET)
            .build();
        
        connector.enviarOrden(buyOrder);
        
        // Send sell order for expensive product
        OrderMessage sellOrder = OrderMessage.builder()
            .clOrdID("arb-sell-" + System.currentTimeMillis())
            .side(OrderSide.SELL)
            .product(getExpensiveProduct())
            .qty(100)
            .mode(OrderMode.MARKET)
            .build();
        
        connector.enviarOrden(sellOrder);
    }

    @Override
    protected void onDetener() {
        log.info("Stopping arbitrage strategy for {} vs {}", productA, productB);
    }

    private Product getCheaperProduct() {
        return lastPriceA < lastPriceB ? 
            Product.valueOf(productA) : Product.valueOf(productB);
    }

    private Product getExpensiveProduct() {
        return lastPriceA > lastPriceB ? 
            Product.valueOf(productA) : Product.valueOf(productB);
    }
}

// Usage
connector.registrarTarea(new ArbitrageStrategy(connector, "GUACA", "CAFE"));
```

## Best Practices

### 1. Choose Appropriate Intervals

- **Fast monitoring** (100-500ms): For price-sensitive strategies
- **Regular updates** (1-5s): For general monitoring
- **Periodic checks** (30s-1m): For less critical tasks

### 2. Keep Execution Fast

Tasks should complete quickly to avoid blocking:

```java
@Override
protected void ejecutar() {
    // ✅ Good: Quick operations
    checkPrices();
    updateMetrics();
    
    // ❌ Bad: Long-running operations
    // downloadLargeFile();
    // complexCalculation();
}
```

### 3. Use Meaningful Keys

```java
// ✅ Good: Descriptive and unique
super("monitor-prices-GUACA");
super("arbitrage-GUACA-CAFE");

// ❌ Bad: Generic or unclear
super("task1");
super("monitor");
```

### 4. Handle Exceptions Gracefully

```java
@Override
protected void ejecutar() {
    try {
        performOperation();
    } catch (Exception e) {
        log.error("Task failed, will retry: {}", e.getMessage());
        // Don't rethrow - let task continue
    }
}
```

### 5. Clean Up Resources

```java
@Override
protected void onDetener() {
    // Release connections, files, etc.
    closeConnections();
    saveState();
}
```

## Threading Model

- All tasks run on **virtual threads** (lightweight)
- Tasks with the same key use a **shared semaphore** for mutual exclusion
- The `ConectorBolsa` manages a single **TareaAutomaticaManager**
- Tasks are automatically stopped on `shutdown()`

## Lifecycle

1. **Registration**: `connector.registrarTarea(task)` → Task starts immediately
2. **Execution**: Task runs according to `intervalo()` or `ejecucionContinua()`
3. **Stopping**: `connector.detenerTarea(key)` → Current execution completes, then stops
4. **Cleanup**: `onDetener()` is called exactly once
5. **Shutdown**: All tasks stop when `connector.shutdown()` is called

## Performance Considerations

- Virtual threads are very lightweight (thousands can run concurrently)
- Key-based locking ensures tasks with the same key never overlap
- Continuous execution mode can consume CPU - use sparingly
- Keep `ejecutar()` fast to maintain responsiveness

## Troubleshooting

### Task Not Executing

- Check that the task is registered: Look for log message "Registered automatic task: {key}"
- Verify `intervalo()` returns a valid duration (not null or negative)
- Check for exceptions in your `ejecutar()` method

### High CPU Usage

- Consider increasing the interval in `intervalo()`
- Check if `ejecucionContinua()` is returning `true`
- Profile your `ejecutar()` method for performance bottlenecks

### Task Not Stopping

- Ensure you're using the correct task key in `detenerTarea()`
- Check logs for "Stopping automatic task: {key}"
- Verify `onDetener()` doesn't throw exceptions

## API Reference

### TareaAutomatica

**Constructor**:
- `protected TareaAutomatica(String taskKey)` - Create task with unique key

**Abstract Methods**:
- `protected abstract void ejecutar()` - Task execution logic

**Override Methods**:
- `protected Duration intervalo()` - Return delay between executions (default: 1 second)
- `protected boolean ejecucionContinua()` - Return true for continuous execution (default: false)
- `protected void onDetener()` - Cleanup callback when task stops (default: no-op)

**Getters**:
- `String getTaskKey()` - Get the task's unique key

### ConectorBolsa

**Task Management**:
- `void registrarTarea(TareaAutomatica tarea)` - Register and start a task
- `void detenerTarea(String taskKey)` - Stop a specific task
- `void shutdown()` - Stop all tasks and release resources

## See Also

- [ConectorBolsa Documentation](../README.md)
- [Event Listener Guide](./EVENT_LISTENERS.md)
- [Trading Strategies Examples](./STRATEGIES.md)

