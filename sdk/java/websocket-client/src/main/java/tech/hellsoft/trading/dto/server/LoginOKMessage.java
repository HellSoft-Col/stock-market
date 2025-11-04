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

/**
 * Message received upon successful authentication with the server.
 *
 * <p>This message contains the complete session information including team details, initial
 * balance, inventory, and authorized products. The client transitions to AUTHENTICATED state after
 * receiving this message.
 *
 * <p>This message is delivered via {@link
 * tech.hellsoft.trading.EventListener#onLoginOk(LoginOKMessage)}.
 *
 * @see tech.hellsoft.trading.EventListener#onLoginOk(LoginOKMessage) for receiving this message
 * @see tech.hellsoft.trading.dto.client.LoginMessage for the authentication request
 */
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class LoginOKMessage {
  /** The message type (LOGIN_OK). */
  private MessageType type;

  /** The team name for this session. */
  private String team;

  /** The species/team type. */
  private String species;

  /** The initial account balance. */
  private Double initialBalance;

  /** The current account balance. */
  private Double currentBalance;

  /**
   * Current inventory levels for each product.
   *
   * <p>Returns an unmodifiable view to prevent external modification.
   */
  private Map<Product, Integer> inventory;

  /**
   * List of products this team is authorized to trade.
   *
   * <p>Returns an unmodifiable view to prevent external modification.
   */
  private List<Product> authorizedProducts;

  /**
   * Available production recipes for this team.
   *
   * <p>Returns an unmodifiable view to prevent external modification.
   */
  private Map<Product, Recipe> recipes;

  /** The team's role in the market. */
  private TeamRole role;

  /** Current server time timestamp. */
  private String serverTime;

  /**
   * Gets the current inventory as an unmodifiable map.
   *
   * @return unmodifiable map of product quantities
   */
  public Map<Product, Integer> getInventory() {
    return inventory == null ? Map.of() : Collections.unmodifiableMap(inventory);
  }

  /**
   * Gets the authorized products as an unmodifiable list.
   *
   * @return unmodifiable list of authorized products
   */
  public List<Product> getAuthorizedProducts() {
    return authorizedProducts == null
        ? List.of()
        : Collections.unmodifiableList(authorizedProducts);
  }

  /**
   * Gets the available recipes as an unmodifiable map.
   *
   * @return unmodifiable map of product recipes
   */
  public Map<Product, Recipe> getRecipes() {
    return recipes == null ? Map.of() : Collections.unmodifiableMap(recipes);
  }
}
