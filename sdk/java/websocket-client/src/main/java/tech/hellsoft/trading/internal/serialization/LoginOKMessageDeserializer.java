package tech.hellsoft.trading.internal.serialization;

import java.lang.reflect.Type;
import java.util.HashMap;
import java.util.Map;

import com.google.gson.*;

import tech.hellsoft.trading.dto.server.LoginOKMessage;
import tech.hellsoft.trading.dto.server.Recipe;
import tech.hellsoft.trading.dto.server.TeamRole;
import tech.hellsoft.trading.enums.MessageType;
import tech.hellsoft.trading.enums.Product;

import lombok.extern.slf4j.Slf4j;

/**
 * Custom deserializer for LoginOKMessage that handles null keys in the recipes map.
 *
 * <p>The server sometimes sends recipes with null product keys, which Gson cannot handle by
 * default. This deserializer filters out null keys.
 */
@Slf4j
public class LoginOKMessageDeserializer implements JsonDeserializer<LoginOKMessage> {

  private static final Gson GSON = new GsonBuilder().create();

  @Override
  public LoginOKMessage deserialize(
      JsonElement json, Type typeOfT, JsonDeserializationContext context)
      throws JsonParseException {
    JsonObject obj = json.getAsJsonObject();

    LoginOKMessage message = new LoginOKMessage();

    // Deserialize simple fields
    message.setType(context.deserialize(obj.get("type"), MessageType.class));
    message.setTeam(obj.has("team") ? obj.get("team").getAsString() : null);
    message.setSpecies(obj.has("species") ? obj.get("species").getAsString() : null);
    message.setInitialBalance(
        obj.has("initialBalance") ? obj.get("initialBalance").getAsDouble() : null);
    message.setCurrentBalance(
        obj.has("currentBalance") ? obj.get("currentBalance").getAsDouble() : null);
    message.setServerTime(obj.has("serverTime") ? obj.get("serverTime").getAsString() : null);

    // Deserialize inventory map
    if (obj.has("inventory")) {
      Map<Product, Integer> inventory = new HashMap<>();
      JsonObject invObj = obj.getAsJsonObject("inventory");
      for (String key : invObj.keySet()) {
        try {
          Product product = Product.fromJson(key);
          inventory.put(product, invObj.get(key).getAsInt());
        } catch (IllegalArgumentException e) {
          // Skip unknown products
          log.debug("Skipping unknown product in inventory: {}", key);
        }
      }
      message.setInventory(inventory);
    }

    // Deserialize authorized products list
    if (obj.has("authorizedProducts")) {
      JsonArray authProducts = obj.getAsJsonArray("authorizedProducts");
      java.util.List<Product> products = new java.util.ArrayList<>();
      for (JsonElement elem : authProducts) {
        try {
          Product product = Product.fromJson(elem.getAsString());
          products.add(product);
        } catch (IllegalArgumentException e) {
          // Skip unknown products
          log.debug("Skipping unknown authorized product: {}", elem.getAsString());
        }
      }
      message.setAuthorizedProducts(products);
    }

    // Deserialize recipes map - filtering out null keys
    if (obj.has("recipes") && !obj.get("recipes").isJsonNull()) {
      Map<Product, Recipe> recipes = new HashMap<>();
      JsonObject recipesObj = obj.getAsJsonObject("recipes");
      for (String key : recipesObj.keySet()) {
        if (key == null || "null".equals(key)) {
          log.debug("Skipping null key in recipes");
          continue;
        }
        try {
          Product product = Product.fromJson(key);
          Recipe recipe = context.deserialize(recipesObj.get(key), Recipe.class);
          recipes.put(product, recipe);
        } catch (IllegalArgumentException e) {
          // Skip unknown products
          log.debug("Skipping unknown product in recipes: {}", key);
        }
      }
      message.setRecipes(recipes);
    }

    // Deserialize role
    if (obj.has("role") && !obj.get("role").isJsonNull()) {
      message.setRole(context.deserialize(obj.get("role"), TeamRole.class));
    }

    return message;
  }
}
