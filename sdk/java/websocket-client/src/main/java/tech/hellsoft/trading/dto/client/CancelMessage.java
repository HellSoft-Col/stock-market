package tech.hellsoft.trading.dto.client;

import tech.hellsoft.trading.enums.MessageType;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class CancelMessage {
  private MessageType type;
  private String clOrdID;
}
