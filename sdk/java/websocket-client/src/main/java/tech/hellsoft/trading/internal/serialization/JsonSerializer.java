package tech.hellsoft.trading.internal.serialization;

import java.lang.reflect.Type;

import com.google.gson.FieldNamingPolicy;
import com.google.gson.Gson;
import com.google.gson.GsonBuilder;
import com.google.gson.JsonIOException;
import com.google.gson.JsonObject;
import com.google.gson.JsonParser;
import com.google.gson.JsonSyntaxException;

import tech.hellsoft.trading.dto.server.InventoryUpdateMessage;
import tech.hellsoft.trading.dto.server.LoginOKMessage;
import tech.hellsoft.trading.enums.Product;

import lombok.extern.slf4j.Slf4j;

@Slf4j
public class JsonSerializer {
  private static final Gson GSON =
      new GsonBuilder()
          .setFieldNamingPolicy(FieldNamingPolicy.LOWER_CASE_WITH_UNDERSCORES)
          .serializeNulls()
          // Custom deserializers for messages with Product maps that may have null keys
          .registerTypeAdapter(LoginOKMessage.class, new LoginOKMessageDeserializer())
          .registerTypeAdapter(
              InventoryUpdateMessage.class, new InventoryUpdateMessageDeserializer())
          // Custom type adapter for Product enum to handle unknown products gracefully
          .registerTypeAdapter(Product.class, new ProductTypeAdapter())
          .create();

  public static String toJson(Object obj) {
    if (obj == null) {
      return "null";
    }

    try {
      return GSON.toJson(obj);
    } catch (JsonIOException e) {
      log.error("Failed to serialize object to JSON: {}", obj.getClass().getName(), e);
      throw new RuntimeException("JSON serialization failed", e);
    }
  }

  public static <T> T fromJson(String json, Class<T> clazz) {
    if (json == null || json.isBlank()) {
      throw new IllegalArgumentException("JSON string cannot be null or blank");
    }

    if (clazz == null) {
      throw new IllegalArgumentException("Target class cannot be null");
    }

    try {
      return GSON.fromJson(json, clazz);
    } catch (JsonSyntaxException e) {
      log.error("Failed to deserialize JSON to {}: {}", clazz.getName(), json, e);
      throw new RuntimeException("JSON deserialization failed", e);
    }
  }

  public static <T> T fromJson(String json, Type type) {
    if (json == null || json.isBlank()) {
      throw new IllegalArgumentException("JSON string cannot be null or blank");
    }

    if (type == null) {
      throw new IllegalArgumentException("Target type cannot be null");
    }

    try {
      return GSON.fromJson(json, type);
    } catch (JsonSyntaxException e) {
      log.error("Failed to deserialize JSON to {}: {}", type.getTypeName(), json, e);
      throw new RuntimeException("JSON deserialization failed", e);
    }
  }

  public static JsonObject parseObject(String json) {
    if (json == null || json.isBlank()) {
      throw new IllegalArgumentException("JSON string cannot be null or blank");
    }

    try {
      return JsonParser.parseString(json).getAsJsonObject();
    } catch (JsonSyntaxException e) {
      log.error("Failed to parse JSON: {}", json, e);
      throw new RuntimeException("JSON parsing failed", e);
    }
  }
}
