package tech.hellsoft.trading.dto.server;

import java.util.Collections;
import java.util.List;
import java.util.Map;

import tech.hellsoft.trading.enums.MessageType;
import tech.hellsoft.trading.enums.Product;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class LoginOKMessage {
  private MessageType type;
  private String team;
  private String species;
  private Double initialBalance;
  private Double currentBalance;
  private Map<Product, Integer> inventory;
  private List<Product> authorizedProducts;
  private Map<Product, Recipe> recipes;
  private TeamRole role;
  private String serverTime;

  public Map<Product, Integer> getInventory() {
    return inventory == null ? Map.of() : Collections.unmodifiableMap(inventory);
  }

  public List<Product> getAuthorizedProducts() {
    return authorizedProducts == null
        ? List.of()
        : Collections.unmodifiableList(authorizedProducts);
  }

  public Map<Product, Recipe> getRecipes() {
    return recipes == null ? Map.of() : Collections.unmodifiableMap(recipes);
  }
}
