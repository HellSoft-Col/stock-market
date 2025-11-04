package tech.hellsoft.trading.dto.server;

import tech.hellsoft.trading.enums.MessageType;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class BalanceUpdateMessage {
  private MessageType type;
  private Double balance;
  private String serverTime;
}
