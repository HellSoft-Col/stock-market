# Student Setup Guide

## Step 1: Configure GitHub Token

Edit `gradle.properties` with your GitHub credentials:

```properties
# Replace these values:
gpr.user=your-github-username
gpr.token=ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx

# Keep these settings as-is:
org.gradle.jvmargs=-Xmx2g
org.gradle.daemon=true
org.gradle.parallel=true
org.gradle.caching=true
```

**Where to get these values:**
- `gpr.user`: Your GitHub username (e.g., `santiago123`)
- `gpr.token`: GitHub Personal Access Token provided by instructor
  - Should start with `ghp_`
  - Needs `read:packages` permission
  - If you need to generate one: GitHub → Settings → Developer settings → Personal access tokens

## Step 2: Configure Trading Server

Edit `TradingBot.java` and replace these values:

```java
// Connection settings (from instructor)
private String serverHost = "wss://trading.hellsoft.tech";  
private int serverPort = 443;            
private String token = "TK-your-team-token-here";  // From instructor
```

**Where to get these values:**
- `serverHost`: WebSocket server URL (ask instructor)
- `serverPort`: Server port (usually 443 for wss://)
- `token`: Your team's trading token (starts with `TK-`)

## Step 3: Open in IntelliJ IDEA

1. **Open Project**
   - File → Open
   - Navigate to and select the `student-starter-project` folder
   - Click "Open"

2. **Configure JDK** (if prompted)
   - File → Project Structure (Cmd+; on Mac)
   - Project → SDK: Select Java 25
   - Language level: 25
   - Click OK

3. **Wait for Gradle Sync**
   - IntelliJ will automatically download dependencies
   - Look for "BUILD SUCCESSFUL" in the Build panel
   - If sync fails, check your `gradle.properties` values

## Step 4: Compile and Run in IntelliJ

### Method 1: Using the Run Button (Easiest)

1. Open `src/main/java/com/myteam/TradingBot.java`
2. Right-click anywhere in the file
3. Select **"Run 'TradingBot.main()'"**
4. Your bot will compile and start!

### Method 2: Using Gradle Panel

1. Open the **Gradle** panel (right side of window)
2. Expand: **student-starter-project → Tasks → build**
3. Double-click **"build"** to compile
4. Then expand **Tasks → application**
5. Double-click **"run"** to execute

### Method 3: Using Terminal

```bash
# Compile
./gradlew build

# Run
./gradlew run
```

## Troubleshooting

### "Could not resolve tech.hellsoft.trading:websocket-client"

**Problem:** Can't download the SDK from GitHub Packages.

**Solution:**
1. Check `gradle.properties` has correct values
2. Verify `gpr.token` starts with `ghp_`
3. In IntelliJ: Gradle panel → Reload All Gradle Projects (🔄)
4. Ask instructor for a new token if needed

### "Connection refused" or "WebSocket error"

**Problem:** Can't connect to trading server.

**Solution:**
1. Verify `serverHost` is correct (should be wss:// URL)
2. Check `serverPort` (usually 443)
3. Make sure server is running (ask instructor)

### "Authentication failed" or "Invalid token"

**Problem:** Server rejected your trading token.

**Solution:**
1. Check `token` in TradingBot.java
2. Verify it starts with `TK-`
3. Ask instructor for correct team token

### "Unsupported class file major version" or "Java 25 not supported"

**Problem:** Wrong Java version installed.

**Solution:**
1. Install Java 25 from https://jdk.java.net/25/
2. In IntelliJ: File → Project Structure → SDK: Java 25
3. Restart IntelliJ

### IntelliJ doesn't show SDK classes

**Problem:** Can't import ConectorBolsa or other SDK classes.

**Solution:**
1. File → Invalidate Caches → Invalidate and Restart
2. Wait for IntelliJ to rebuild index
3. Check Gradle sync completed successfully

## Quick Compilation Guide

### To compile in IntelliJ:
```
Right-click TradingBot.java → Run 'TradingBot.main()'
```

### To compile via terminal:
```bash
./gradlew build      # Compile only
./gradlew run        # Compile and run
```

### Before submitting code:
```bash
./gradlew spotlessApply    # Format code
./gradlew build            # Compile and test
```

## Next Steps

1. ✅ Complete Steps 1-4 above
2. 🏃 Run the bot to see it connect
3. 📝 Look for `// TODO:` comments in TradingBot.java
4. 🤖 Implement your trading strategy!
5. 🎯 Test and refine your bot

Good luck! 🚀
