package tech.hellsoft.trading.internal.connection;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.mockito.ArgumentCaptor;

import java.net.http.WebSocket;
import java.nio.ByteBuffer;
import java.util.concurrent.CompletionStage;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.concurrent.atomic.AtomicReference;
import java.util.function.Consumer;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.Mockito.*;

class WebSocketHandlerTest {

    private WebSocketHandler handler;
    private Consumer<String> onMessage;
    private Consumer<Throwable> onError;
    private Runnable onClose;
    private WebSocket mockWebSocket;

    @SuppressWarnings("unchecked")
    @BeforeEach
    void setUp() {
        onMessage = mock(Consumer.class);
        onError = mock(Consumer.class);
        onClose = mock(Runnable.class);
        mockWebSocket = mock(WebSocket.class);
    }

    @Test
    void shouldCreateHandlerWithValidCallbacks() {
        handler = new WebSocketHandler(onMessage, onError, onClose);
        assertNotNull(handler);
    }

    @Test
    void shouldRejectNullOnMessage() {
        assertThrows(IllegalArgumentException.class, () ->
            new WebSocketHandler(null, onError, onClose)
        );
    }

    @Test
    void shouldRejectNullOnError() {
        assertThrows(IllegalArgumentException.class, () ->
            new WebSocketHandler(onMessage, null, onClose)
        );
    }

    @Test
    void shouldRejectNullOnClose() {
        assertThrows(IllegalArgumentException.class, () ->
            new WebSocketHandler(onMessage, onError, null)
        );
    }

    @Test
    void shouldRequestDataOnOpen() {
        handler = new WebSocketHandler(onMessage, onError, onClose);

        handler.onOpen(mockWebSocket);

        verify(mockWebSocket).request(1);
    }

    @Test
    void shouldProcessCompleteTextMessage() {
        handler = new WebSocketHandler(onMessage, onError, onClose);

        handler.onText(mockWebSocket, "{\"type\":\"TEST\"}", true);

        ArgumentCaptor<String> captor = ArgumentCaptor.forClass(String.class);
        verify(onMessage).accept(captor.capture());
        assertEquals("{\"type\":\"TEST\"}", captor.getValue());
        verify(mockWebSocket).request(1);
    }

    @Test
    void shouldBufferFragmentedTextMessages() {
        handler = new WebSocketHandler(onMessage, onError, onClose);

        handler.onText(mockWebSocket, "{\"type\":", false);
        handler.onText(mockWebSocket, "\"TEST\"}", true);

        ArgumentCaptor<String> captor = ArgumentCaptor.forClass(String.class);
        verify(onMessage).accept(captor.capture());
        assertEquals("{\"type\":\"TEST\"}", captor.getValue());
        verify(mockWebSocket, times(2)).request(1);
    }

    @Test
    void shouldHandleMultipleFragments() {
        handler = new WebSocketHandler(onMessage, onError, onClose);

        handler.onText(mockWebSocket, "{", false);
        handler.onText(mockWebSocket, "\"type\"", false);
        handler.onText(mockWebSocket, ":", false);
        handler.onText(mockWebSocket, "\"TEST\"", false);
        handler.onText(mockWebSocket, "}", true);

        ArgumentCaptor<String> captor = ArgumentCaptor.forClass(String.class);
        verify(onMessage).accept(captor.capture());
        assertEquals("{\"type\":\"TEST\"}", captor.getValue());
        verify(mockWebSocket, times(5)).request(1);
    }

    @Test
    void shouldClearBufferAfterCompleteMessage() {
        handler = new WebSocketHandler(onMessage, onError, onClose);

        handler.onText(mockWebSocket, "first", true);
        handler.onText(mockWebSocket, "second", true);

        ArgumentCaptor<String> captor = ArgumentCaptor.forClass(String.class);
        verify(onMessage, times(2)).accept(captor.capture());
        
        assertEquals("first", captor.getAllValues().get(0));
        assertEquals("second", captor.getAllValues().get(1));
    }

    @Test
    void shouldHandleExceptionInMessageCallback() {
        doThrow(new RuntimeException("Callback error"))
            .when(onMessage).accept(anyString());

        handler = new WebSocketHandler(onMessage, onError, onClose);

        assertDoesNotThrow(() ->
            handler.onText(mockWebSocket, "{\"type\":\"TEST\"}", true)
        );

        verify(onMessage).accept(anyString());
        verify(mockWebSocket).request(1);
    }

    @Test
    void shouldProcessMultipleConsecutiveMessages() {
        handler = new WebSocketHandler(onMessage, onError, onClose);

        handler.onText(mockWebSocket, "message1", true);
        handler.onText(mockWebSocket, "message2", true);
        handler.onText(mockWebSocket, "message3", true);

        verify(onMessage, times(3)).accept(anyString());
        verify(mockWebSocket, times(3)).request(1);
    }

    @Test
    void shouldHandleBinaryData() {
        handler = new WebSocketHandler(onMessage, onError, onClose);
        ByteBuffer buffer = ByteBuffer.wrap(new byte[]{1, 2, 3});

        CompletionStage<?> result = handler.onBinary(mockWebSocket, buffer, true);

        assertNull(result);
        verify(mockWebSocket).request(1);
        verifyNoInteractions(onMessage);
    }

    @Test
    void shouldHandlePing() {
        handler = new WebSocketHandler(onMessage, onError, onClose);
        ByteBuffer buffer = ByteBuffer.allocate(0);

        CompletionStage<?> result = handler.onPing(mockWebSocket, buffer);

        assertNull(result);
        verify(mockWebSocket).request(1);
        verifyNoInteractions(onMessage);
    }

    @Test
    void shouldHandlePong() {
        handler = new WebSocketHandler(onMessage, onError, onClose);
        ByteBuffer buffer = ByteBuffer.allocate(0);

        CompletionStage<?> result = handler.onPong(mockWebSocket, buffer);

        assertNull(result);
        verify(mockWebSocket).request(1);
        verifyNoInteractions(onMessage);
    }

    @Test
    void shouldInvokeCloseCallback() {
        handler = new WebSocketHandler(onMessage, onError, onClose);

        handler.onClose(mockWebSocket, 1000, "Normal closure");

        verify(onClose).run();
    }

    @Test
    void shouldHandleExceptionInCloseCallback() {
        doThrow(new RuntimeException("Close callback error"))
            .when(onClose).run();

        handler = new WebSocketHandler(onMessage, onError, onClose);

        assertDoesNotThrow(() ->
            handler.onClose(mockWebSocket, 1000, "Normal closure")
        );

        verify(onClose).run();
    }

    @Test
    void shouldInvokeErrorCallback() {
        handler = new WebSocketHandler(onMessage, onError, onClose);
        Throwable testError = new RuntimeException("Test error");

        handler.onError(mockWebSocket, testError);

        ArgumentCaptor<Throwable> captor = ArgumentCaptor.forClass(Throwable.class);
        verify(onError).accept(captor.capture());
        assertEquals(testError, captor.getValue());
    }

    @Test
    void shouldHandleExceptionInErrorCallback() {
        doThrow(new RuntimeException("Error callback error"))
            .when(onError).accept(any());

        handler = new WebSocketHandler(onMessage, onError, onClose);
        Throwable testError = new RuntimeException("Test error");

        assertDoesNotThrow(() ->
            handler.onError(mockWebSocket, testError)
        );

        verify(onError).accept(any());
    }

    @Test
    void shouldHandleEmptyMessages() {
        handler = new WebSocketHandler(onMessage, onError, onClose);

        handler.onText(mockWebSocket, "", true);

        ArgumentCaptor<String> captor = ArgumentCaptor.forClass(String.class);
        verify(onMessage).accept(captor.capture());
        assertEquals("", captor.getValue());
    }

    @Test
    void shouldHandleVeryLongMessages() {
        handler = new WebSocketHandler(onMessage, onError, onClose);
        String longMessage = "x".repeat(10000);

        handler.onText(mockWebSocket, longMessage, true);

        ArgumentCaptor<String> captor = ArgumentCaptor.forClass(String.class);
        verify(onMessage).accept(captor.capture());
        assertEquals(longMessage, captor.getValue());
    }

    @Test
    void shouldHandleVeryLongFragmentedMessages() {
        handler = new WebSocketHandler(onMessage, onError, onClose);
        
        handler.onText(mockWebSocket, "x".repeat(5000), false);
        handler.onText(mockWebSocket, "y".repeat(5000), true);

        ArgumentCaptor<String> captor = ArgumentCaptor.forClass(String.class);
        verify(onMessage).accept(captor.capture());
        
        String result = captor.getValue();
        assertTrue(result.startsWith("xxx"));
        assertTrue(result.endsWith("yyy"));
        assertEquals(10000, result.length());
    }

    @Test
    void shouldHandleSpecialCharacters() {
        handler = new WebSocketHandler(onMessage, onError, onClose);
        String specialChars = "{\"message\":\"Hello 世界 🌍\"}";

        handler.onText(mockWebSocket, specialChars, true);

        ArgumentCaptor<String> captor = ArgumentCaptor.forClass(String.class);
        verify(onMessage).accept(captor.capture());
        assertEquals(specialChars, captor.getValue());
    }

    @Test
    void shouldHandleRapidMessageSequence() {
        handler = new WebSocketHandler(onMessage, onError, onClose);

        for (int i = 0; i < 100; i++) {
            handler.onText(mockWebSocket, "message" + i, true);
        }

        verify(onMessage, times(100)).accept(anyString());
        verify(mockWebSocket, times(100)).request(1);
    }

    @Test
    void shouldHandleCloseWithDifferentStatusCodes() {
        handler = new WebSocketHandler(onMessage, onError, onClose);

        handler.onClose(mockWebSocket, 1000, "Normal");
        handler.onClose(mockWebSocket, 1001, "Going away");
        handler.onClose(mockWebSocket, 1002, "Protocol error");

        verify(onClose, times(3)).run();
    }

    @Test
    void shouldNotInterfereBetweenMessages() {
        AtomicInteger messageCount = new AtomicInteger(0);
        AtomicReference<String> lastMessage = new AtomicReference<>();

        Consumer<String> trackingCallback = msg -> {
            messageCount.incrementAndGet();
            lastMessage.set(msg);
        };

        handler = new WebSocketHandler(trackingCallback, onError, onClose);

        handler.onText(mockWebSocket, "first", true);
        assertEquals("first", lastMessage.get());
        assertEquals(1, messageCount.get());

        handler.onText(mockWebSocket, "second", true);
        assertEquals("second", lastMessage.get());
        assertEquals(2, messageCount.get());
    }

    @Test
    void shouldReturnNullCompletionStage() {
        handler = new WebSocketHandler(onMessage, onError, onClose);

        assertNull(handler.onText(mockWebSocket, "test", true));
        assertNull(handler.onBinary(mockWebSocket, ByteBuffer.allocate(0), true));
        assertNull(handler.onPing(mockWebSocket, ByteBuffer.allocate(0)));
        assertNull(handler.onPong(mockWebSocket, ByteBuffer.allocate(0)));
        assertNull(handler.onClose(mockWebSocket, 1000, "Normal"));
    }

    @Test
    void shouldHandleMixedOperations() {
        handler = new WebSocketHandler(onMessage, onError, onClose);

        handler.onOpen(mockWebSocket);
        handler.onText(mockWebSocket, "message", true);
        handler.onPing(mockWebSocket, ByteBuffer.allocate(0));
        handler.onPong(mockWebSocket, ByteBuffer.allocate(0));
        handler.onBinary(mockWebSocket, ByteBuffer.allocate(0), true);
        handler.onClose(mockWebSocket, 1000, "Normal");

        verify(mockWebSocket, times(5)).request(1);
        verify(onMessage).accept("message");
        verify(onClose).run();
    }

    @Test
    void shouldHandleConcurrentCallbacks() throws InterruptedException {
        CountDownLatch latch = new CountDownLatch(10);
        AtomicInteger callbackCount = new AtomicInteger(0);

        Consumer<String> concurrentCallback = msg -> {
            callbackCount.incrementAndGet();
            latch.countDown();
        };

        handler = new WebSocketHandler(concurrentCallback, onError, onClose);

        for (int i = 0; i < 10; i++) {
            final int messageNum = i;
            Thread.startVirtualThread(() ->
                handler.onText(mockWebSocket, "message" + messageNum, true)
            );
        }

        assertTrue(latch.await(2, TimeUnit.SECONDS));
        assertEquals(10, callbackCount.get());
    }

    @Test
    void shouldPreserveMessageOrderWithFragments() {
        handler = new WebSocketHandler(onMessage, onError, onClose);

        handler.onText(mockWebSocket, "part1-", false);
        handler.onText(mockWebSocket, "part2-", false);
        handler.onText(mockWebSocket, "part3", true);

        ArgumentCaptor<String> captor = ArgumentCaptor.forClass(String.class);
        verify(onMessage).accept(captor.capture());
        assertEquals("part1-part2-part3", captor.getValue());
    }
}
