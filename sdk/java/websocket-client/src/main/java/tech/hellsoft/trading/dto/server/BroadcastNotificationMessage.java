package tech.hellsoft.trading.dto.server;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import tech.hellsoft.trading.enums.MessageType;

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
