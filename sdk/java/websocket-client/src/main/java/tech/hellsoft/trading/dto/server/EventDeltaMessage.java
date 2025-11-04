package tech.hellsoft.trading.dto.server;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import tech.hellsoft.trading.enums.MessageType;

import java.util.List;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class EventDeltaMessage {
    private MessageType type;
    private List<FillMessage> events;
    private String serverTime;
}
