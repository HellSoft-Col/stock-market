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
public class PongMessage {
    private MessageType type;
    private String timestamp;
}
