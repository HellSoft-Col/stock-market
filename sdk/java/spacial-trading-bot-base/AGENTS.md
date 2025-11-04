# AGENTS.md

## 1. Short Introduction

This document provides coding guidelines for building a CLI trading bot client for the Bolsa Interestelar de Aguacates Andorianos. It establishes principles for writing clean, maintainable Java 25 code with strict adherence to "no `else`" style. This is a **guide on how to think and write code**, not an implementation plan. Students should adapt these principles to their own solutions.

---

## 2. Design Principles (Java 25, Style, and "No Else")

### Naming Conventions
- **Classes**: `UpperCamelCase` (e.g., `ClienteBolsa`, `EstadoCliente`)
- **Methods & variables**: `lowerCamelCase` (e.g., `calcularPL`, `saldoActual`)
- **Constants**: `UPPER_SNAKE_CASE` (e.g., `MAX_RETRIES`, `DEFAULT_PORT`)
- **Packages**: lowercase (e.g., `tech.hellsoft.trading`)

### The "No Else" Rule
Ban `else` keywords entirely. Use these alternatives:

1. **Guard clauses** (early return/throw)
2. **Strategy pattern** or polymorphism
3. **`switch` expressions** (Java 25)

**Example: Guard Clauses**
```java
// ❌ Bad (nested if-else)
public void procesarOrden(String producto, int cantidad) {
    if (producto != null) {
        if (cantidad > 0) {
            if (saldo >= costo) {
                ejecutar();
            } else {
                throw new SaldoInsuficienteException();
            }
        } else {
            throw new CantidadInvalidaException();
        }
    } else {
        throw new ProductoNuloException();
    }
}

// ✅ Good (guard clauses, no else)
public void procesarOrden(String producto, int cantidad) {
    if (producto == null) {
        throw new ProductoNuloException();
    }
    if (cantidad <= 0) {
        throw new CantidadInvalidaException();
    }
    if (saldo < costo) {
        throw new SaldoInsuficienteException();
    }
    // All validations passed → execute
    ejecutar();
}
```

**Example: Switch Expressions (Java 25)**
```java
// ❌ Bad (if-else chain)
public String procesarComando(String cmd) {
    if (cmd.equals("comprar")) {
        return handleComprar();
    } else if (cmd.equals("vender")) {
        return handleVender();
    } else if (cmd.equals("producir")) {
        return handleProducir();
    } else {
        return "Unknown";
    }
}

// ✅ Good (switch expression)
public String procesarComando(String cmd) {
    return switch (cmd) {
        case "comprar" -> handleComprar();
        case "vender" -> handleVender();
        case "producir" -> handleProducir();
        default -> "Unknown";
    };
}
```

### General Guidelines
- **Cohesion**: One responsibility per class; methods ≤ 20–30 lines
- **Null safety**: Use `Objects.requireNonNull` at entry points
- **Fail fast**: Validate early; throw exceptions immediately
- **Minimal stdout**: Log only essential user feedback

---

## 3. Package & Naming Structure

Recommended structure based on project requirements:

```
tech.hellsoft.trading
├── ClienteBolsa.java              // Main orchestrator (implements EventListener)
├── EstadoCliente.java             // Runtime state (Serializable)
├── CalculadoraProduccion.java     // Recursive production algorithm
├── RecetaValidator.java           // Ingredient validation logic
├── SnapshotManager.java           // Binary serialization utilities
├── ConfigLoader.java              // JSON configuration reader
├── ConsolaInteractiva.java        // CLI command parser
├── dto/
│   ├── Rol.java
│   ├── Receta.java
│   └── Config.java
├── exception/
│   ├── TradingException.java      // Base class (optional)
│   ├── SaldoInsuficienteException.java
│   ├── InventarioInsuficienteException.java
│   ├── ProductoNoAutorizadoException.java
│   ├── IngredientesInsuficientesException.java
│   ├── RecetaNoEncontradaException.java
│   ├── ConfiguracionInvalidaException.java
│   └── SnapshotCorruptoException.java
└── util/                          // Optional helpers
```

---

## 4. Core Responsibilities & Patterns

### 4.1 ClienteBolsa — Main Orchestrator
**Role**: Implements `EventListener`; coordinates SDK and business logic.

**Key Pattern**: Guard clauses for validation before SDK calls.

```java
public void comprar(String producto, int cantidad, String mensaje) {
    // Guard clauses validate BEFORE calling SDK
    if (producto == null || producto.isBlank()) {
        throw new ValidacionOrdenException("Product required");
    }
    if (cantidad <= 0) {
        throw new ValidacionOrdenException("Quantity must be positive");
    }
    
    double costoEstimado = estimarCosto(producto, cantidad);
    if (estado.getSaldo() < costoEstimado) {
        throw new SaldoInsuficienteException(estado.getSaldo(), costoEstimado);
    }
    
    // All validations passed → send to SDK
    conector.enviarOrden(crearOrden("BUY", producto, cantidad, mensaje));
}
```

**Callbacks Pattern**: Handle each SDK event separately; avoid nested logic.

```java
@Override
public void onFill(Fill fill) {
    if (fill == null) {
        return;
    }
    
    if ("BUY".equals(fill.getSide())) {
        estado.actualizarSaldo(-fill.getTotal());
        estado.actualizarInventario(fill.getProducto(), fill.getCantidad());
        return;
    }
    
    if ("SELL".equals(fill.getSide())) {
        estado.actualizarSaldo(fill.getTotal());
        estado.actualizarInventario(fill.getProducto(), -fill.getCantidad());
    }
}
```

---

### 4.2 EstadoCliente — State Management
**Role**: Track balance, inventory, prices. Must be `Serializable`.

**Key Pattern**: Stream operations for calculations.

```java
public double calcularPL() {
    double valorInventario = inventario.entrySet().stream()
        .mapToDouble(e -> {
            double precio = preciosActuales.getOrDefault(e.getKey(), 0.0);
            return e.getValue() * precio;
        })
        .sum();
    
    double patrimonioNeto = saldo + valorInventario;
    return ((patrimonioNeto - saldoInicial) / saldoInicial) * 100.0;
}
```

---

### 4.3 CalculadoraProduccion — Recursive Algorithm
**Role**: Calculate production units using recursion.

**Key Pattern**: Clear base case + recursive case.

```java
public static int calcularUnidades(Rol rol) {
    return calcularRecursivo(0, rol);
}

private static int calcularRecursivo(int nivel, Rol rol) {
    // Base case: depth limit reached
    if (nivel > rol.getMaxDepth()) {
        return 0;
    }
    
    // Calculate contribution at this level
    double energia = rol.getBaseEnergy() + rol.getLevelEnergy() * nivel;
    double factor = Math.pow(rol.getDecay(), nivel) * Math.pow(rol.getBranches(), nivel);
    int contribucion = (int) Math.round(energia * factor);
    
    // Recursive case: add contributions from deeper levels
    return contribucion + calcularRecursivo(nivel + 1, rol);
}
```

---

### 4.4 RecetaValidator — Ingredient Logic
**Role**: Check and consume ingredients.

**Key Pattern**: Guard clauses for each ingredient.

```java
public static boolean puedeProducir(Receta receta, Map<String, Integer> inventario) {
    if (receta.getIngredientes() == null) {
        return true; // Basic production (no ingredients)
    }
    
    for (Map.Entry<String, Integer> ingrediente : receta.getIngredientes().entrySet()) {
        int disponible = inventario.getOrDefault(ingrediente.getKey(), 0);
        if (disponible < ingrediente.getValue()) {
            return false;
        }
    }
    
    return true;
}
```

---

### 4.5 SnapshotManager — Persistence
**Role**: Save/load state for crash recovery.

**Key Pattern**: Guard clauses for file operations.

```java
public static EstadoCliente cargar(String archivo) throws IOException {
    Path path = Path.of(archivo);
    
    if (!Files.exists(path)) {
        throw new SnapshotCorruptoException("File not found");
    }
    
    try (ObjectInputStream ois = new ObjectInputStream(Files.newInputStream(path))) {
        return (EstadoCliente) ois.readObject();
    } catch (ClassNotFoundException e) {
        throw new SnapshotCorruptoException("Invalid format", e);
    }
}
```

---

### 4.6 ConfigLoader — JSON Parsing
**Role**: Load config and recipes at startup.

**Key Pattern**: Validate required fields after parsing.

```java
public static Config cargarConfig(String archivo) throws IOException {
    String json = Files.readString(Path.of(archivo));
    
    if (json == null || json.isBlank()) {
        throw new ConfiguracionInvalidaException("Empty config file");
    }
    
    Config config = gson.fromJson(json, Config.class);
    
    if (config.getApiKey() == null || config.getApiKey().isBlank()) {
        throw new ConfiguracionInvalidaException("Missing apiKey");
    }
    
    return config;
}
```

---

### 4.7 ConsolaInteractiva — Command Parser
**Role**: Parse user input and delegate to `ClienteBolsa`.

**Key Pattern**: Switch expressions for command routing.

```java
public void processCommand(String input) {
    if (input == null || input.isBlank()) {
        return;
    }
    
    String[] tokens = input.trim().split("\\s+");
    String comando = tokens[0].toLowerCase();
    
    try {
        switch (comando) {
            case "comprar" -> handleComprar(tokens);
            case "vender" -> handleVender(tokens);
            case "producir" -> handleProducir(tokens);
            case "status" -> handleStatus();
            case "exit" -> System.exit(0);
            default -> System.out.println("Unknown command: " + comando);
        }
    } catch (Exception e) {
        System.out.println("❌ Error: " + e.getMessage());
    }
}
```

---

## 5. Exception Handling Strategy

### Required Business Exceptions (7 minimum)

| Exception | When to Throw | Example Context |
|-----------|---------------|-----------------|
| `SaldoInsuficienteException` | Balance < cost | Before `comprar()` |
| `InventarioInsuficienteException` | Inventory < quantity | Before `vender()` |
| `ProductoNoAutorizadoException` | Product not in whitelist | In `producir()` |
| `IngredientesInsuficientesException` | Missing recipe items | In `producir()` premium |
| `RecetaNoEncontradaException` | Unknown recipe name | In `producir()` |
| `ConfiguracionInvalidaException` | Bad JSON or missing fields | In `ConfigLoader` |
| `SnapshotCorruptoException` | Damaged snapshot file | In `SnapshotManager.cargar()` |

### Exception Pattern: Information-Rich Messages

```java
// ❌ Bad (no context)
throw new SaldoInsuficienteException("Insufficient balance");

// ✅ Good (includes context)
throw new SaldoInsuficienteException(
    String.format("Insufficient balance. Required: %.2f, Available: %.2f", 
        costoRequerido, saldoActual)
);
```

### Validation Strategy
1. **Validate locally FIRST** (client-side)
2. **Then call SDK** (server-side)
3. **Handle server errors** via `onError()` callback

---

## 6. Testing Approach

### Unit Test Priorities
Focus on **pure logic** (no SDK, no I/O):

1. **Recursive algorithm**: Test `CalculadoraProduccion` with known inputs
2. **Exception throwing**: Verify guard clauses throw correct exceptions
3. **P&L calculation**: Test `EstadoCliente.calcularPL()` with mock data
4. **Ingredient checks**: Test `RecetaValidator.puedeProducir()`

**Example (JUnit 5)**:
```java
@Test
void shouldThrowWhenBalanceInsufficient() {
    EstadoCliente estado = new EstadoCliente();
    estado.setSaldo(50.0);
    
    ClienteBolsa cliente = new ClienteBolsa(conector, estado);
    
    assertThrows(SaldoInsuficienteException.class, () -> {
        cliente.comprar("FOSFO", 100, "test"); // Costs > 50
    });
}

@Test
void shouldCalculateProductionRecursively() {
    Rol rol = new Rol(2, 4, 0.7651, 3.0, 2.0);
    int unidades = CalculadoraProduccion.calcularUnidades(rol);
    
    assertTrue(unidades > 0, "Should produce some units");
    assertTrue(unidades < 200, "Should not produce unreasonably many");
}
```

---

## 7. Dependency Guidelines

### Philosophy: JDK-First
**"First look for solutions in what we already have."**

### Allowed (if truly needed)
- **JSON**: Gson OR Jackson (choose one; prefer Gson for simplicity)
  ```gradle
  implementation("com.google.code.gson:gson:2.10.1")
  ```

### Forbidden
- Spring Framework, Guice, Dagger
- Lombok (use Java 25 records)
- Apache Commons Lang
- Heavy logging (SLF4J, Logback)

### JDK Alternatives
- Date/time → `java.time.*`
- Collections → `java.util.*`
- Streams → `java.util.stream.*`
- I/O → `java.nio.file.*`

---

## 8. Code Quality Tools (Gradle)

The project includes Spotless, Checkstyle, and PMD.

### Essential Commands

```bash
# Format code (run before committing)
./gradlew spotlessApply

# Check formatting (CI/validation)
./gradlew spotlessCheck

# Run all linters
./gradlew checkstyleMain pmdMain

# Run tests
./gradlew test

# Full build (format + lint + test)
./gradlew build
```

### Typical Workflow
1. Write code
2. Format: `./gradlew spotlessApply`
3. Validate: `./gradlew check`
4. Commit

---

## 9. SDK Integration Patterns

The project uses `tech.hellsoft.trading:websocket-client:1.0.3` from the GitHub repository `HellSoft-Col/stock-market`.

### Connection Pattern
```java
ConectorBolsa conector = new ConectorBolsa();
ClienteBolsa cliente = new ClienteBolsa(conector);

conector.conectar("localhost", 9000);
conector.login("TK-YOUR-API-KEY", cliente); // Pass EventListener implementation
```

### EventListener Implementation
Your `ClienteBolsa` must implement all callbacks. Use guard clauses:

```java
@Override
public void onLoginOk(LoginOk msg) {
    if (msg == null) {
        return;
    }
    // Initialize estado from msg
}

@Override
public void onError(ErrorMessage error) {
    if (error == null) {
        return;
    }
    
    switch (error.getCodigo()) {
        case "INVALID_TOKEN" -> {
            System.err.println("Invalid API key");
            System.exit(1);
        }
        case "INSUFFICIENT_BALANCE" -> System.err.println("Balance too low (server)");
        default -> System.err.println("Error: " + error.getMensaje());
    }
}
```

**Consult SDK Javadoc** for exact method signatures and DTO structures.

---

## 10. Key Concepts Glossary

| Term | Definition |
|------|------------|
| **Guard Clause** | Early return/throw that validates preconditions (eliminates `else`) |
| **Fill** | Confirmation that an order executed (transaction complete) |
| **Ticker** | Periodic price update (bestBid, bestAsk, mid) |
| **P&L** | Profit & Loss percentage: `((current - initial) / initial) × 100` |
| **Snapshot** | Serialized state for crash recovery |
| **Recursion** | Function calling itself with simpler inputs until base case |
| **Switch Expression** | Java 25 feature returning a value (no `break`, no `else`) |

---

**End of AGENTS.md**

*Remember: This is a guide on **how to think**, not **what to code**. Adapt these patterns to your solution. Trust your intelligence.*

