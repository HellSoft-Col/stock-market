package tech.hellsoft.trading.dto.client;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import tech.hellsoft.trading.enums.MessageType;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class CancelMessage {
    private MessageType type;
    private String clOrdID;
}
