# Trading Bot Starter Project

Ready-to-use starter template for building trading bots.

## Quick Start

### 1. Configure GitHub Access
Edit `gradle.properties`:
```properties
gpr.user=your-github-username
gpr.token=ghp_xxxxxxxxxxxx  # From instructor
```

### 2. Configure Trading Server
Edit `TradingBot.java`:
```java
private String serverHost = "wss://trading.hellsoft.tech";
private String token = "TK-your-team-token";  // From instructor
```

### 3. Open in IntelliJ
- File → Open → Select this folder
- Wait for Gradle sync to complete

### 4. Compile and Run
**In IntelliJ:** Right-click `TradingBot.java` → Run 'TradingBot.main()'

**Or via terminal:** `./gradlew run`

See `SETUP.md` for detailed instructions with examples.

## Files

- `src/main/java/com/myteam/TradingBot.java` - Your bot code (edit this!)
- `build.gradle.kts` - Dependencies and build config
- `gradle.properties` - GitHub token config
- `SETUP.md` - Detailed setup instructions

## Your Tasks

Look for `// TODO:` comments in `TradingBot.java` - that's where you implement your strategy!

Good luck! 🚀
