package tech.hellsoft.trading.dto.server;

import java.util.Map;

import tech.hellsoft.trading.enums.MessageType;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class PerformanceReportMessage {
  private MessageType messageType;
  private String teamName;
  private Double startBalance;
  private Double finalBalance;
  private Double profitLoss;
  private Double roi;
  private Integer totalTrades;
  private Double totalVolume;
  private Double avgTradeSize;
  private Integer buyTrades;
  private Integer sellTrades;
  private Map<String, Integer> products;
  private Map<String, Integer> finalInventory;
  private Integer rank;
  private Integer totalTeams;
  private String severTime;
}
