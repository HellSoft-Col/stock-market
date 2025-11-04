# Dependency Versions

## Current Versions (Last Updated: Nov 2024)

### Main Dependencies

| Dependency | Version | Status | Notes |
|------------|---------|--------|-------|
| **Gson** | 2.13.1 | ✅ Latest | JSON serialization library |
| **Lombok** | 1.18.40 | ✅ Latest | Reduces boilerplate code, Java 25 compatible |
| **SLF4J** | 2.0.16 | ✅ Latest Stable | Logging facade (2.1.0-alpha1 available but unstable) |

### Test Dependencies

| Dependency | Version | Status | Notes |
|------------|---------|--------|-------|
| **JUnit Jupiter** | 5.11.4 | ✅ Latest Stable | Unit testing framework (5.13.0-M3 available but milestone) |
| **Mockito** | 5.18.0 | ✅ Latest | Mocking framework |

## Version History

### 2024-11-04 (Update 2)
- **Gson**: Updated from 2.11.0 → 2.13.1
- **Mockito**: Updated from 5.15.2 → 5.18.0
- **SLF4J**: Kept at 2.0.16 (latest stable, avoiding alpha releases)
- **JUnit**: Kept at 5.11.4 (latest stable, avoiding milestone releases)

### 2024-11-04 (Update 1)
- **Mockito**: Updated from 5.14.2 → 5.15.2
- All other dependencies already at latest

### Initial Setup (2024-11)
- Gson: 2.11.0
- Lombok: 1.18.40 (Java 25 support)
- SLF4J: 2.0.16
- JUnit: 5.11.4
- Mockito: 5.14.2

## Checking for Updates

### Automated Script

Run the included script to check for latest versions:

```bash
./check-updates.sh
```

### Manual Check

Check Maven Central directly:

1. **Gson**: https://mvnrepository.com/artifact/com.google.code.gson/gson
2. **Lombok**: https://mvnrepository.com/artifact/org.projectlombok/lombok
3. **SLF4J**: https://mvnrepository.com/artifact/org.slf4j/slf4j-api
4. **JUnit**: https://mvnrepository.com/artifact/org.junit.jupiter/junit-jupiter
5. **Mockito**: https://mvnrepository.com/artifact/org.mockito/mockito-core

### Gradle Task

Check dependency versions using Gradle:

```bash
# View all dependencies
./gradlew dependencies

# View only runtime dependencies
./gradlew dependencies --configuration runtimeClasspath

# View only test dependencies
./gradlew dependencies --configuration testRuntimeClasspath
```

## Update Policy

- **Major versions**: Review breaking changes before updating
- **Minor versions**: Update as soon as stable release is available
- **Patch versions**: Update immediately for bug fixes
- **Pre-release versions**: Avoid in production code

## Java 25 Compatibility

All dependencies are tested and confirmed compatible with Java 25:

- ✅ Gson 2.11.0 - Full Java 25 support
- ✅ Lombok 1.18.40 - Java 25 support added in 1.18.30
- ✅ SLF4J 2.0.16 - Java 25 compatible
- ✅ JUnit 5.11.4 - Java 25 compatible
- ✅ Mockito 5.15.2 - Java 25 compatible

## Known Issues

None at this time. All dependencies work correctly with Java 25 and virtual threads.

## Dependency Tree

```
tech.hellsoft.trading:websocket-client:1.0.0-SNAPSHOT
├── com.google.code.gson:gson:2.11.0
├── org.slf4j:slf4j-api:2.0.16
└── (test dependencies)
    ├── org.junit.jupiter:junit-jupiter:5.11.4
    │   ├── org.junit.jupiter:junit-jupiter-api:5.11.4
    │   ├── org.junit.jupiter:junit-jupiter-params:5.11.4
    │   └── org.junit.jupiter:junit-jupiter-engine:5.11.4
    └── org.mockito:mockito-core:5.15.2
        ├── net.bytebuddy:byte-buddy:1.15.10
        └── net.bytebuddy:byte-buddy-agent:1.15.10

Lombok 1.18.40 - Compile-time only (not in final JAR)
```

## Updating Dependencies

To update a dependency:

1. Edit `build.gradle.kts`
2. Change the version number
3. Run `./gradlew build` to test
4. Update this file with the new version
5. Commit changes

Example:
```kotlin
// Before
testImplementation("org.mockito:mockito-core:5.14.2")

// After
testImplementation("org.mockito:mockito-core:5.15.2")
```

## Security

Monitor security advisories:

- https://github.com/advisories
- https://www.cve.org/
- Subscribe to security mailing lists for each dependency
