package tech.hellsoft.trading.internal.routing;

import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

import lombok.extern.slf4j.Slf4j;

@Slf4j
public class MessageSequencer {
  private final ExecutorService sequentialExecutor =
      Executors.newSingleThreadExecutor(Thread.ofVirtual().factory());

  public void submit(Runnable task) {
    if (sequentialExecutor.isShutdown()) {
      log.warn("Cannot submit task, sequencer is shut down");
      return;
    }

    sequentialExecutor.execute(
        () -> {
          try {
            task.run();
          } catch (Exception e) {
            log.error("Error processing sequenced task", e);
          }
        });
  }

  public void shutdown() {
    log.debug("Shutting down message sequencer");
    sequentialExecutor.shutdown();
  }

  public boolean isShutdown() {
    return sequentialExecutor.isShutdown();
  }
}
