package tech.hellsoft.trading.model;

import java.io.Serializable;
import java.util.HashMap;
import java.util.Map;
import lombok.Data;
import lombok.NoArgsConstructor;
import lombok.AllArgsConstructor;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class Recipe implements Serializable {
  private String product;
  private Map<String, Integer> ingredients = new HashMap<>();
  private double premiumBonus;

  public boolean isBasic() {
    return ingredients == null || ingredients.isEmpty();
  }
}
