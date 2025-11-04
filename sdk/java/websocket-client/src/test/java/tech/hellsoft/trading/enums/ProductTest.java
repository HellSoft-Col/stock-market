package tech.hellsoft.trading.enums;

import static org.junit.jupiter.api.Assertions.*;

import java.util.stream.Stream;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.MethodSource;

class ProductTest {

  static Stream<Arguments> productValues() {
    return Stream.of(
        Arguments.of(Product.GUACA, "GUACA"),
        Arguments.of(Product.SEBO, "SEBO"),
        Arguments.of(Product.FOSFO, "FOSFO"),
        Arguments.of(Product.PALTA_OIL, "PALTA-OIL"),
        Arguments.of(Product.CASCAR_ALLOY, "CASCAR-ALLOY"));
  }

  @ParameterizedTest
  @MethodSource("productValues")
  void shouldReturnCorrectValue(Product product, String expectedValue) {
    assertEquals(expectedValue, product.getValue());
  }

  @ParameterizedTest
  @MethodSource("productValues")
  void shouldDeserializeFromJson(Product expectedProduct, String jsonValue) {
    Product result = Product.fromJson(jsonValue);
    assertEquals(expectedProduct, result);
  }

  @Test
  void shouldHandleHyphenatedProductNames() {
    assertEquals("PALTA-OIL", Product.PALTA_OIL.getValue());
    assertEquals("CASCAR-ALLOY", Product.CASCAR_ALLOY.getValue());
  }

  @Test
  void shouldThrowExceptionForInvalidValue() {
    assertThrows(IllegalArgumentException.class, () -> Product.fromJson("INVALID_PRODUCT"));
  }

  @Test
  void shouldThrowExceptionForNullValue() {
    assertThrows(IllegalArgumentException.class, () -> Product.fromJson(null));
  }
}
