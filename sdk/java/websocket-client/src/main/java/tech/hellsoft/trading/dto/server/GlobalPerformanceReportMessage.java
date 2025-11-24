package tech.hellsoft.trading.dto.server;

import java.util.List;
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
public class GlobalPerformanceReportMessage {
  private MessageType type;
  private String duration;
  private Integer totalTrades;
  private Double totalVolume;
  private List<PerformanceReportMessage> topTraders;
  private Map<String, ProductStatsMessage> productStats;
  private String serverTime;
}
