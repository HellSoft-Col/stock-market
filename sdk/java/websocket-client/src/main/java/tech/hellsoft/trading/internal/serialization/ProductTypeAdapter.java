package tech.hellsoft.trading.internal.serialization;

import java.io.IOException;

import com.google.gson.TypeAdapter;
import com.google.gson.stream.JsonReader;
import com.google.gson.stream.JsonToken;
import com.google.gson.stream.JsonWriter;

import tech.hellsoft.trading.enums.Product;

import lombok.extern.slf4j.Slf4j;

/**
 * Custom TypeAdapter for Product enum that handles unknown products gracefully.
 *
 * <p>When deserializing, if an unknown product is encountered, it returns null instead of throwing
 * an exception. This allows the application to continue working even when the server adds new
 * products that the client doesn't know about yet.
 */
@Slf4j
public class ProductTypeAdapter extends TypeAdapter<Product> {

  @Override
  public void write(JsonWriter out, Product value) throws IOException {
    if (value == null) {
      out.nullValue();
    } else {
      out.value(value.getValue());
    }
  }

  @Override
  public Product read(JsonReader in) throws IOException {
    if (in.peek() == JsonToken.NULL) {
      in.nextNull();
      return null;
    }

    String value = in.nextString();

    if (value == null || value.isEmpty()) {
      return null;
    }

    try {
      return Product.fromJson(value);
    } catch (IllegalArgumentException e) {
      // Unknown product - log and return null instead of failing
      log.debug("Unknown product: {} - skipping", value);
      return null;
    }
  }
}
