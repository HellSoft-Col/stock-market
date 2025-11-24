package tech.hellsoft.trading.dto.server;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class ProductStatsMessage {
  private String product;
  private Integer totalTrades;
  private Double totalVolume;
  private Double avgPrice;
  private Double minPrice;
  private Double maxPrice;
  private Double lastPrice;
}
