package tech.hellsoft.trading.model;

import java.io.Serializable;
import lombok.Data;
import lombok.NoArgsConstructor;
import lombok.AllArgsConstructor;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class Role implements Serializable {
  private int branches;
  private int maxDepth;
  private double decay;
  private double baseEnergy;
  private double levelEnergy;
}
