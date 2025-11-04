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
public class BroadcastNotificationMessage {
  private MessageType type;
  private String message;
  private String sender;
  private String serverTime;
}
