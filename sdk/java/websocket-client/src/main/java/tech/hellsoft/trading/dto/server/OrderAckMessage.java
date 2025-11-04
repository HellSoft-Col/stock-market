package tech.hellsoft.trading.dto.server;

import tech.hellsoft.trading.enums.MessageType;
import tech.hellsoft.trading.enums.OrderStatus;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class OrderAckMessage {
  private MessageType type;
  private String clOrdID;
  private OrderStatus status;
  private String serverTime;
}
