package tech.hellsoft.trading.internal.serialization;

import com.google.gson.JsonObject;
import com.google.gson.reflect.TypeToken;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.NullAndEmptySource;
import org.junit.jupiter.params.provider.ValueSource;
import tech.hellsoft.trading.dto.client.OrderMessage;
import tech.hellsoft.trading.enums.MessageType;
import tech.hellsoft.trading.enums.OrderMode;
import tech.hellsoft.trading.enums.OrderSide;
import tech.hellsoft.trading.enums.Product;

import java.lang.reflect.Type;
import java.util.List;
import java.util.Map;

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
        assertThrows(IllegalArgumentException.class, () ->
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
        assertThrows(IllegalArgumentException.class, () ->
            JsonSerializer.parseObject(json)
        );
    }

    @Test
    void shouldDeserializeWithGenericType() {
        String json = "[\"GUACA\",\"SEBO\",\"FOSFO\"]";
        Type listType = new TypeToken<List<String>>(){}.getType();

        List<String> result = JsonSerializer.fromJson(json, listType);

        assertNotNull(result);
        assertEquals(3, result.size());
        assertEquals("GUACA", result.get(0));
        assertEquals("SEBO", result.get(1));
        assertEquals("FOSFO", result.get(2));
    }

    @Test
    void shouldDeserializeMapWithType() {
        String json = "{\"GUACA\":100,\"SEBO\":50}";
        Type mapType = new TypeToken<Map<String, Integer>>(){}.getType();

        Map<String, Integer> result = JsonSerializer.fromJson(json, mapType);

        assertNotNull(result);
        assertEquals(2, result.size());
        assertEquals(100, result.get("GUACA"));
        assertEquals(50, result.get("SEBO"));
    }

    @Test
    void shouldThrowExceptionForNullType() {
        assertThrows(IllegalArgumentException.class, () ->
            JsonSerializer.fromJson("{}", (Type) null)
        );
    }

    @ParameterizedTest
    @NullAndEmptySource
    void shouldThrowExceptionForInvalidJsonWithType(String json) {
        Type listType = new TypeToken<List<String>>(){}.getType();
        
        assertThrows(IllegalArgumentException.class, () ->
            JsonSerializer.fromJson(json, listType)
        );
    }

    @Test
    void shouldThrowExceptionForMalformedJsonWithType() {
        Type listType = new TypeToken<List<String>>(){}.getType();
        
        assertThrows(RuntimeException.class, () ->
            JsonSerializer.fromJson("{invalid json}", listType)
        );
    }

    @Test
    void shouldSerializeNullFieldsInObject() {
        OrderMessage order = OrderMessage.builder()
            .type(MessageType.ORDER)
            .clOrdID("ORDER-123")
            .build();

        String json = JsonSerializer.toJson(order);

        assertTrue(json.contains("\"side\":null"));
        assertTrue(json.contains("\"mode\":null"));
        assertTrue(json.contains("\"product\":null"));
    }

    @Test
    void shouldHandleSnakeCaseFieldNaming() {
        String json = "{\"cl_ord_id\":\"TEST-123\",\"limit_price\":99.5}";

        OrderMessage result = JsonSerializer.fromJson(json, OrderMessage.class);

        assertEquals("TEST-123", result.getClOrdID());
        assertEquals(99.5, result.getLimitPrice());
    }

    @Test
    void shouldThrowExceptionForMalformedJson() {
        String malformedJson = "{invalid}";

        assertThrows(RuntimeException.class, () ->
            JsonSerializer.fromJson(malformedJson, OrderMessage.class)
        );
    }

    @Test
    void shouldThrowExceptionForMalformedJsonObject() {
        String malformedJson = "[1,2,3]";

        assertThrows(RuntimeException.class, () ->
            JsonSerializer.parseObject(malformedJson)
        );
    }

    @ParameterizedTest
    @ValueSource(strings = {" ", "\t", "\n", "   "})
    void shouldRejectWhitespaceOnlyJson(String json) {
        assertThrows(IllegalArgumentException.class, () ->
            JsonSerializer.fromJson(json, OrderMessage.class)
        );
    }

    @ParameterizedTest
    @ValueSource(strings = {" ", "\t", "\n", "   "})
    void shouldRejectWhitespaceOnlyJsonForType(String json) {
        Type listType = new TypeToken<List<String>>(){}.getType();
        
        assertThrows(IllegalArgumentException.class, () ->
            JsonSerializer.fromJson(json, listType)
        );
    }

    @ParameterizedTest
    @ValueSource(strings = {" ", "\t", "\n", "   "})
    void shouldRejectWhitespaceOnlyJsonForParsing(String json) {
        assertThrows(IllegalArgumentException.class, () ->
            JsonSerializer.parseObject(json)
        );
    }

    @Test
    void shouldHandleComplexFieldValues() {
        String json = "{\"type\":\"ORDER\",\"cl_ord_id\":\"ORD-1\",\"side\":\"BUY\",\"product\":\"GUACA\",\"qty\":15}";

        OrderMessage result = JsonSerializer.fromJson(json, OrderMessage.class);

        assertNotNull(result);
        assertEquals("ORD-1", result.getClOrdID());
        assertEquals(Product.GUACA, result.getProduct());
        assertEquals(15, result.getQty());
    }

    @Test
    void shouldSerializeAndDeserializeRoundTrip() {
        OrderMessage original = OrderMessage.builder()
            .type(MessageType.ORDER)
            .clOrdID("ROUND-TRIP")
            .side(OrderSide.SELL)
            .mode(OrderMode.MARKET)
            .product(Product.SEBO)
            .qty(25)
            .build();

        String json = JsonSerializer.toJson(original);
        OrderMessage deserialized = JsonSerializer.fromJson(json, OrderMessage.class);

        assertEquals(original.getClOrdID(), deserialized.getClOrdID());
        assertEquals(original.getSide(), deserialized.getSide());
        assertEquals(original.getMode(), deserialized.getMode());
        assertEquals(original.getProduct(), deserialized.getProduct());
        assertEquals(original.getQty(), deserialized.getQty());
    }

    @Test
    void shouldHandleEmptyJsonObject() {
        String json = "{}";

        OrderMessage result = JsonSerializer.fromJson(json, OrderMessage.class);

        assertNotNull(result);
        assertNull(result.getClOrdID());
        assertNull(result.getSide());
    }

    @Test
    void shouldParseJsonObjectWithNestedObjects() {
        String json = "{\"outer\":{\"inner\":\"value\"},\"array\":[1,2,3]}";

        JsonObject result = JsonSerializer.parseObject(json);

        assertTrue(result.has("outer"));
        assertTrue(result.get("outer").isJsonObject());
        assertTrue(result.has("array"));
        assertTrue(result.get("array").isJsonArray());
    }

    @Test
    void shouldDeserializeListOfComplexObjects() {
        String json = "[{\"type\":\"ORDER\",\"qty\":10},{\"type\":\"ORDER\",\"qty\":20}]";
        Type listType = new TypeToken<List<OrderMessage>>(){}.getType();

        List<OrderMessage> result = JsonSerializer.fromJson(json, listType);

        assertNotNull(result);
        assertEquals(2, result.size());
        assertEquals(10, result.get(0).getQty());
        assertEquals(20, result.get(1).getQty());
    }

    @Test
    void shouldHandleNumericFieldsInJson() {
        String json = "{\"qty\":42,\"limit_price\":123.456}";

        OrderMessage result = JsonSerializer.fromJson(json, OrderMessage.class);

        assertEquals(42, result.getQty());
        assertEquals(123.456, result.getLimitPrice());
    }

    @Test
    void shouldSerializePrimitiveTypes() {
        String intJson = JsonSerializer.toJson(42);
        String stringJson = JsonSerializer.toJson("test");
        String boolJson = JsonSerializer.toJson(true);

        assertEquals("42", intJson);
        assertEquals("\"test\"", stringJson);
        assertEquals("true", boolJson);
    }
}

