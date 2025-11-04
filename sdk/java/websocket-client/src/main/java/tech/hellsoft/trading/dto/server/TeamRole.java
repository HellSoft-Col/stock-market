package tech.hellsoft.trading.dto.server;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class TeamRole {
  private Integer branches;
  private Integer maxDepth;
  private Double decay;
  private Double budget;
  private Double baseEnergy;
  private Double levelEnergy;
}
