package tech.hellsoft.trading.internal.serialization;

import java.lang.reflect.Type;
import java.util.HashMap;
import java.util.Map;

import com.google.gson.*;

import tech.hellsoft.trading.dto.server.InventoryUpdateMessage;
import tech.hellsoft.trading.enums.MessageType;
import tech.hellsoft.trading.enums.Product;

import lombok.extern.slf4j.Slf4j;

/**
 * Custom deserializer for InventoryUpdateMessage that handles unknown products and null keys.
 *
 * <p>The server may send inventory updates with unknown products or null keys. This deserializer
 * filters them out gracefully instead of failing.
 */
@Slf4j
public class InventoryUpdateMessageDeserializer
    implements JsonDeserializer<InventoryUpdateMessage> {

  @Override
  public InventoryUpdateMessage deserialize(
      JsonElement json, Type typeOfT, JsonDeserializationContext context)
      throws JsonParseException {
    JsonObject obj = json.getAsJsonObject();

    InventoryUpdateMessage message = new InventoryUpdateMessage();

    // Deserialize type
    if (obj.has("type")) {
      message.setType(context.deserialize(obj.get("type"), MessageType.class));
    }

    // Deserialize serverTime
    if (obj.has("serverTime")) {
      message.setServerTime(obj.get("serverTime").getAsString());
    }

    // Deserialize inventory map - filtering out null keys and unknown products
    if (obj.has("inventory")) {
      Map<Product, Integer> inventory = new HashMap<>();
      JsonObject invObj = obj.getAsJsonObject("inventory");
      for (String key : invObj.keySet()) {
        if (key == null || "null".equals(key)) {
          log.debug("Skipping null key in inventory");
          continue;
        }
        try {
          Product product = Product.fromJson(key);
          inventory.put(product, invObj.get(key).getAsInt());
        } catch (IllegalArgumentException e) {
          // Skip unknown products
          log.debug("Skipping unknown product: {}", key);
        }
      }
      message.setInventory(inventory);
    }

    return message;
  }
}
