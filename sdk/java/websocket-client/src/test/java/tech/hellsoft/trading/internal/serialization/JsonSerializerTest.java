package tech.hellsoft.trading.internal.serialization;

import com.google.gson.JsonObject;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.NullAndEmptySource;
import tech.hellsoft.trading.dto.client.OrderMessage;
import tech.hellsoft.trading.enums.MessageType;
import tech.hellsoft.trading.enums.OrderMode;
import tech.hellsoft.trading.enums.OrderSide;
import tech.hellsoft.trading.enums.Product;

import static org.junit.jupiter.api.Assertions.*;

class JsonSerializerTest {

    @Test
    void shouldSerializeObjectToJson() {
        OrderMessage order = OrderMessage.builder()
            .type(MessageType.ORDER)
            .clOrdID("ORDER-123")
            .side(OrderSide.BUY)
            .mode(OrderMode.LIMIT)
            .product(Product.GUACA)
            .qty(10)
            .limitPrice(100.0)
            .build();

        String json = JsonSerializer.toJson(order);

        assertNotNull(json);
        assertTrue(json.contains("\"cl_ord_id\":\"ORDER-123\""));
        assertTrue(json.contains("\"type\":\"ORDER\""));
        assertTrue(json.contains("\"side\":\"BUY\""));
    }

    @Test
    void shouldDeserializeJsonToObject() {
        String json = "{\"type\":\"ORDER\",\"cl_ord_id\":\"ORDER-123\",\"side\":\"BUY\",\"mode\":\"LIMIT\",\"product\":\"GUACA\",\"qty\":10,\"limit_price\":100.0}";

        OrderMessage result = JsonSerializer.fromJson(json, OrderMessage.class);

        assertNotNull(result);
        assertEquals("ORDER-123", result.getClOrdID());
        assertEquals(OrderSide.BUY, result.getSide());
        assertEquals(Product.GUACA, result.getProduct());
        assertEquals(10, result.getQty());
    }

    @Test
    void shouldHandleNullSerialization() {
        String result = JsonSerializer.toJson(null);
        assertEquals("null", result);
    }

    @ParameterizedTest
    @NullAndEmptySource
    void shouldThrowExceptionForInvalidJsonDeserialization(String json) {
        assertThrows(RuntimeException.class, () ->
            JsonSerializer.fromJson(json, OrderMessage.class)
        );
    }

    @Test
    void shouldThrowExceptionForNullClass() {
        assertThrows(IllegalArgumentException.class, () ->
            JsonSerializer.fromJson("{}", null)
        );
    }

    @Test
    void shouldParseJsonObject() {
        String json = "{\"type\":\"ORDER\",\"qty\":10}";
        
        JsonObject obj = JsonSerializer.parseObject(json);
        
        assertNotNull(obj);
        assertTrue(obj.has("type"));
        assertEquals("ORDER", obj.get("type").getAsString());
    }

    @ParameterizedTest
    @NullAndEmptySource
    void shouldThrowExceptionForInvalidJsonParsing(String json) {
        assertThrows(RuntimeException.class, () ->
            JsonSerializer.parseObject(json)
        );
    }
}
