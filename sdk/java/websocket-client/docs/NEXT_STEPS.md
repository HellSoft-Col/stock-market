# Next Steps - Java SDK Implementation

**Last Session End:** 2025-01-04  
**Current Progress:** 32% Complete (21 of ~65 files)  
**Build Status:** ✅ PASSING

---

## 🎯 IMMEDIATE PRIORITY: Complete Server DTOs

### Files to Create (11 remaining)

Create these in `src/main/java/tech/hellsoft/trading/dto/server/`:

#### 1. LoginOKMessage.java ⚠️ COMPLEX
```java
package tech.hellsoft.trading.dto.server;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import tech.hellsoft.trading.enums.MessageType;
import tech.hellsoft.trading.enums.Product;

import java.util.Collections;
import java.util.List;
import java.util.Map;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class LoginOKMessage {
    private MessageType type;
    private String team;
    private String species;
    private Double initialBalance;
    private Double currentBalance;
    private Map<Product, Integer> inventory;
    private List<Product> authorizedProducts;
    private Map<Product, Recipe> recipes;
    private TeamRole role;
    private String serverTime;
    
    // IMPORTANT: Return unmodifiable collections
    public Map<Product, Integer> getInventory() {
        return inventory == null ? Map.of() : Collections.unmodifiableMap(inventory);
    }
    
    public List<Product> getAuthorizedProducts() {
        return authorizedProducts == null ? List.of() : Collections.unmodifiableList(authorizedProducts);
    }
    
    public Map<Product, Recipe> getRecipes() {
        return recipes == null ? Map.of() : Collections.unmodifiableMap(recipes);
    }
}
```

#### 2. FillMessage.java
```java
package tech.hellsoft.trading.dto.server;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import tech.hellsoft.trading.enums.MessageType;
import tech.hellsoft.trading.enums.OrderSide;
import tech.hellsoft.trading.enums.Product;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class FillMessage {
    private MessageType type;
    private String clOrdID;
    private Integer fillQty;
    private Double fillPrice;
    private OrderSide side;
    private Product product;
    private String counterparty;
    private String counterpartyMessage;
    private String serverTime;
    private Integer remainingQty;
    private Integer totalQty;
}
```

#### 3. TickerMessage.java
```java
package tech.hellsoft.trading.dto.server;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import tech.hellsoft.trading.enums.MessageType;
import tech.hellsoft.trading.enums.Product;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class TickerMessage {
    private MessageType type;
    private Product product;
    private Double bestBid;
    private Double bestAsk;
    private Double mid;
    private Integer volume24h;
    private String serverTime;
}
```

#### 4. OfferMessage.java
```java
package tech.hellsoft.trading.dto.server;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import tech.hellsoft.trading.enums.MessageType;
import tech.hellsoft.trading.enums.Product;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class OfferMessage {
    private MessageType type;
    private String offerId;
    private String buyer;
    private Product product;
    private Integer quantityRequested;
    private Double maxPrice;
    private Integer expiresIn;
    private String timestamp;
}
```

#### 5. ErrorMessage.java
```java
package tech.hellsoft.trading.dto.server;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import tech.hellsoft.trading.enums.ErrorCode;
import tech.hellsoft.trading.enums.MessageType;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class ErrorMessage {
    private MessageType type;
    private ErrorCode code;
    private String reason;
    private String clOrdID;
    private String timestamp;
}
```

#### 6. OrderAckMessage.java
```java
package tech.hellsoft.trading.dto.server;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import tech.hellsoft.trading.enums.MessageType;
import tech.hellsoft.trading.enums.OrderStatus;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class OrderAckMessage {
    private MessageType type;
    private String clOrdID;
    private OrderStatus status;
    private String serverTime;
}
```

#### 7. InventoryUpdateMessage.java
```java
package tech.hellsoft.trading.dto.server;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import tech.hellsoft.trading.enums.MessageType;
import tech.hellsoft.trading.enums.Product;

import java.util.Map;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class InventoryUpdateMessage {
    private MessageType type;
    private Map<Product, Integer> inventory;
    private String serverTime;
}
```

#### 8. BalanceUpdateMessage.java
```java
package tech.hellsoft.trading.dto.server;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import tech.hellsoft.trading.enums.MessageType;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class BalanceUpdateMessage {
    private MessageType type;
    private Double balance;
    private String serverTime;
}
```

#### 9. EventDeltaMessage.java
```java
package tech.hellsoft.trading.dto.server;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import tech.hellsoft.trading.enums.MessageType;

import java.util.List;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class EventDeltaMessage {
    private MessageType type;
    private List<FillMessage> events;
    private String serverTime;
}
```

#### 10. BroadcastNotificationMessage.java
```java
package tech.hellsoft.trading.dto.server;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import tech.hellsoft.trading.enums.MessageType;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class BroadcastNotificationMessage {
    private MessageType type;
    private String message;
    private String sender;
    private String serverTime;
}
```

#### 11. PongMessage.java
```java
package tech.hellsoft.trading.dto.server;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import tech.hellsoft.trading.enums.MessageType;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class PongMessage {
    private MessageType type;
    private String timestamp;
}
```

### Verification After Creating Server DTOs
```bash
cd /path/to/websocket-client
./gradlew build
# Should still show: BUILD SUCCESSFUL
```

---

## 📋 SUBSEQUENT PHASES (In Order)

### Phase 5: Internal Infrastructure - State Management
Create in `src/main/java/tech/hellsoft/trading/internal/routing/`:

1. **StateLocker.java** - Thread-safe state mutation locking
   - Uses ReentrantLock per action
   - Timeout configuration
   - See IMPLEMENTATION_PLAN.md lines 1354-1436

2. **MessageSequencer.java** - Sequential message processing by type
   - Single-thread executor per MessageType
   - Virtual threads
   - See IMPLEMENTATION_PLAN.md lines 1438-1595

### Phase 6: Internal Infrastructure - Message Processing
Create in `src/main/java/tech/hellsoft/trading/internal/`:

1. **serialization/JsonSerializer.java** - JSON serialization utilities
   - Static Gson instance
   - Custom InstantAdapter
   - See IMPLEMENTATION_PLAN.md lines 1604-1679

2. **routing/MessageRouter.java** - Message routing logic
   - Switch expression for routing
   - Integrates StateLocker and MessageSequencer
   - See IMPLEMENTATION_PLAN.md lines 1681-1845

### Phase 7: Connection Management
Create in `src/main/java/tech/hellsoft/trading/internal/connection/`:

1. **HeartbeatManager.java** - Scheduled heartbeat pings
   - Virtual thread scheduler
   - See IMPLEMENTATION_PLAN.md lines 1853-1957

2. **ConnectionManager.java** - Connection lifecycle management
   - Auto-reconnect with exponential backoff
   - State management
   - See IMPLEMENTATION_PLAN.md lines 1959-2176

3. **WebSocketHandler.java** - WebSocket.Listener implementation
   - Message buffering
   - Error handling
   - See IMPLEMENTATION_PLAN.md lines 2179-2248

### Phase 8: Public API
Create in `src/main/java/tech/hellsoft/trading/`:

1. **EventListener.java** (interface) - Callback interface
   - 7 callback methods
   - See IMPLEMENTATION_PLAN.md lines 2256-2313

2. **ConectorBolsa.java** - Main SDK class
   - Public methods: conectar(), login(), enviarOrden(), etc.
   - Integration of all components
   - See IMPLEMENTATION_PLAN.md lines 2315-2600+

### Phase 9: Testing
Create in `src/test/java/tech/hellsoft/trading/`:
- Unit tests for each major component
- JUnit 5 + Mockito
- Test coverage for validation, serialization, connection management

### Phase 10: Documentation
- Javadoc for all public APIs
- README.md with usage examples
- Configuration guide

---

## ⏱️ ESTIMATED TIME

| Phase | Files | Estimated Time |
|-------|-------|----------------|
| Server DTOs (remaining) | 11 | 30-45 minutes |
| State Management | 2 | 1-1.5 hours |
| Message Processing | 2 | 1-1.5 hours |
| Connection Management | 3 | 1.5-2 hours |
| Public API | 2 | 1.5-2 hours |
| Testing | 15+ | 2-3 hours |
| Documentation | - | 1 hour |
| **TOTAL** | **35+** | **8-11 hours** |

---

## 🔑 KEY REMINDERS

### Code Style (from AGENTS.md)
- ❌ **NO** else statements - use guard clauses
- ✅ Functional programming (streams, lambdas, Optional)
- ✅ Virtual threads for concurrency (`Executors.newVirtualThreadPerTaskExecutor()`)
- ✅ Lombok (@Data, @Builder, @Slf4j, @Getter)
- ✅ Validate early, fail fast

### Patterns Established
- All DTOs use: `@Data @Builder @NoArgsConstructor @AllArgsConstructor`
- All enums have: `getValue()` and `fromJson(String)` methods
- Config has: static `defaultConfig()` and `validate()` method
- Exceptions use: `@Getter` for fields

---

## 🚀 QUICK START COMMANDS

```bash
# Navigate to project
cd /Users/santiago.chaustregladly.com/Git/Personal/stock-market/sdk/java/websocket-client

# Verify current build
./gradlew clean build

# Create server DTOs (use templates above)
# Then verify again
./gradlew build

# View progress
cat docs/PHASE_PROGRESS.md
```

---

## 📚 REFERENCE DOCUMENTS

- **IMPLEMENTATION_PLAN.md** - Complete specification with all code
- **AGENTS.md** - Code style guide and patterns
- **PHASE_PROGRESS.md** - Detailed progress tracking
- **SESSION_SUMMARY.md** - Summary of what was completed
- **build.gradle.kts** - Working build configuration

---

## ✅ SUCCESS CRITERIA

After completing Server DTOs:
- [ ] All 11 server DTO files created
- [ ] Build passes: `./gradlew build`
- [ ] Update PHASE_PROGRESS.md to mark Phase 4 as complete
- [ ] Ready to start Phase 5 (State Management)

---

**Last Updated:** 2025-01-04  
**Next Session Goal:** Complete all server DTOs and start Phase 5

---

**END OF NEXT STEPS**
