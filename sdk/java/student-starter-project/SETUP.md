# Student Setup Guide

## Step 1: Configure GitHub Token

Edit `gradle.properties`:
```properties
gpr.user=YOUR_GITHUB_USERNAME
gpr.token=ghp_xxxx  # Token from instructor
```

## Step 2: Configure Trading Server

Edit `TradingBot.java`:
```java
private String serverHost = "localhost";  # From instructor
private int serverPort = 8080;            # From instructor
private String token = "YOUR_TOKEN_HERE"; # From instructor
```

## Step 3: Open in IntelliJ

1. File → Open → Select `student-starter-project` folder
2. Wait for Gradle sync (downloads SDK)
3. Done!

## Step 4: Run

Right-click `TradingBot.java` → Run

Or terminal: `./gradlew run`

## Troubleshooting

**"Could not resolve websocket-client"** → Check GitHub token in gradle.properties

**"Connection refused"** → Check server host/port

**"Authentication failed"** → Check trading token in TradingBot.java

**"Java 25 not supported"** → Install Java 25 from https://jdk.java.net/25/

## Next Steps

1. Run the bot to see it connect
2. Look for `// TODO:` comments in TradingBot.java
3. Implement your trading strategy!

Good luck! 🚀
