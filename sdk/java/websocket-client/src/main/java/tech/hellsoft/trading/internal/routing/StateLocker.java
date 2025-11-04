package tech.hellsoft.trading.internal.routing;

import java.time.Duration;
import java.util.concurrent.Semaphore;
import java.util.concurrent.TimeUnit;
import java.util.function.Supplier;

import tech.hellsoft.trading.exception.StateLockException;

import lombok.extern.slf4j.Slf4j;

@Slf4j
public class StateLocker {
  private final Semaphore lock = new Semaphore(1);
  private final Duration lockTimeout;

  public StateLocker(Duration lockTimeout) {
    if (lockTimeout == null || lockTimeout.isNegative() || lockTimeout.isZero()) {
      throw new IllegalArgumentException("lockTimeout must be positive");
    }
    this.lockTimeout = lockTimeout;
  }

  public <T> T withLock(Supplier<T> action) throws StateLockException {
    if (!acquireLock()) {
      throw new StateLockException("Failed to acquire state lock within " + lockTimeout);
    }

    try {
      return action.get();
    } finally {
      lock.release();
    }
  }

  public void withLockVoid(Runnable action) throws StateLockException {
    if (!acquireLock()) {
      throw new StateLockException("Failed to acquire state lock within " + lockTimeout);
    }

    try {
      action.run();
    } finally {
      lock.release();
    }
  }

  private boolean acquireLock() {
    try {
      return lock.tryAcquire(lockTimeout.toMillis(), TimeUnit.MILLISECONDS);
    } catch (InterruptedException e) {
      Thread.currentThread().interrupt();
      log.warn("Lock acquisition interrupted", e);
      return false;
    }
  }
}
