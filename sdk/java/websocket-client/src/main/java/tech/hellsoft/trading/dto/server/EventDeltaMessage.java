package tech.hellsoft.trading.dto.server;

import java.util.List;

import tech.hellsoft.trading.enums.MessageType;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class EventDeltaMessage {
  private MessageType type;
  private List<FillMessage> events;
  private String serverTime;
}
