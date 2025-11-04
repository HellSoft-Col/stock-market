package tech.hellsoft.trading.enums;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.MethodSource;

import java.util.stream.Stream;

import static org.junit.jupiter.api.Assertions.*;

class RecipeTypeTest {

    @Test
    void shouldHaveCorrectNumberOfTypes() {
        assertEquals(2, RecipeType.values().length);
    }

    @ParameterizedTest
    @MethodSource("allTypes")
    void shouldHaveNonNullValue(RecipeType type) {
        assertNotNull(type.getValue());
        assertFalse(type.getValue().isEmpty());
    }

    @Test
    void shouldHaveBasicType() {
        assertEquals("BASIC", RecipeType.BASIC.getValue());
    }

    @Test
    void shouldHavePremiumType() {
        assertEquals("PREMIUM", RecipeType.PREMIUM.getValue());
    }

    @ParameterizedTest
    @MethodSource("allTypes")
    void shouldSerializeToJson(RecipeType type) {
        String json = type.getValue();
        assertNotNull(json);
        assertEquals(type.getValue(), json);
    }

    @Test
    void shouldDeserializeBasicFromJson() {
        RecipeType result = RecipeType.fromJson("BASIC");
        assertEquals(RecipeType.BASIC, result);
    }

    @Test
    void shouldDeserializePremiumFromJson() {
        RecipeType result = RecipeType.fromJson("PREMIUM");
        assertEquals(RecipeType.PREMIUM, result);
    }

    @Test
    void shouldThrowExceptionForInvalidType() {
        assertThrows(IllegalArgumentException.class, () ->
            RecipeType.fromJson("INVALID")
        );
    }

    @Test
    void shouldThrowExceptionForNullType() {
        assertThrows(IllegalArgumentException.class, () ->
            RecipeType.fromJson(null)
        );
    }

    @Test
    void shouldThrowExceptionForEmptyType() {
        assertThrows(IllegalArgumentException.class, () ->
            RecipeType.fromJson("")
        );
    }

    @ParameterizedTest
    @MethodSource("allTypes")
    void shouldRoundTripThroughJson(RecipeType type) {
        String json = type.getValue();
        RecipeType result = RecipeType.fromJson(json);
        assertEquals(type, result);
    }

    static Stream<RecipeType> allTypes() {
        return Stream.of(RecipeType.values());
    }
}
