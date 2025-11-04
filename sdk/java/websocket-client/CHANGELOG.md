# Changelog

All notable changes to the Stock Market Java SDK will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0-SNAPSHOT] - 2024-11-04

### Added
- Initial release of Java 25 WebSocket SDK
- Complete message type system with 8 enumerations
- 20 DTOs for client-server communication
- WebSocket connection management with built-in `java.net.http.WebSocket`
- Automatic heartbeat/ping-pong management
- Virtual thread support for all concurrent operations
- Sequential message processing with `MessageSequencer`
- Type-safe message routing with `MessageRouter`
- Event-driven architecture with `EventListener` interface
- Configuration system with `ConectorConfig` (16 parameters)
- JSON serialization with Gson
- Thread-safe message sending with Semaphore
- Comprehensive error handling with custom exceptions
- Builder pattern for all DTOs using Lombok
- Automated dependency version checking with `check-updates.sh`
- Complete README with usage examples
- Dependency version tracking in `DEPENDENCY_VERSIONS.md`

### Dependencies
- Gson 2.13.1
- Lombok 1.18.40 (Java 25 support)
- SLF4J 2.0.16
- JUnit Jupiter 5.11.4
- Mockito 5.18.0

### Changed
- Updated Gson from 2.11.0 to 2.13.1
- Updated Mockito from 5.14.2 to 5.18.0

### Technical Details
- Java 25 with virtual threads
- No else statements (guard clause pattern)
- Functional programming style (streams, lambdas)
- Lombok for boilerplate reduction
- Immutable collections for thread safety
- CopyOnWriteArrayList for listener management

## [Unreleased]

### Planned
- Unit tests for all components
- Integration tests
- Auto-reconnect implementation
- Message validation before sending
- JavaDoc generation
- Maven Central publishing

---

## Version History

### [1.0.0-SNAPSHOT] - Current Development
- First complete implementation
- All core features working
- Build successful
- Ready for testing
