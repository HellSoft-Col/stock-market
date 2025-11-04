package tech.hellsoft.trading.dto.server;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import tech.hellsoft.trading.enums.Product;
import tech.hellsoft.trading.enums.RecipeType;

import java.util.Map;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class Recipe {
    private RecipeType type;
    private Map<Product, Integer> ingredients;
    private Double premiumBonus;
}
