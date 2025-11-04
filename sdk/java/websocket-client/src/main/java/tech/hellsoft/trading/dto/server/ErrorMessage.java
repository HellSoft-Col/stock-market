package tech.hellsoft.trading.dto.server;

import tech.hellsoft.trading.enums.ErrorCode;
import tech.hellsoft.trading.enums.MessageType;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class ErrorMessage {
  private MessageType type;
  private ErrorCode code;
  private String reason;
  private String clOrdID;
  private String timestamp;
}
