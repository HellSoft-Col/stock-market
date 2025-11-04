package tech.hellsoft.trading.dto.client;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import tech.hellsoft.trading.enums.MessageType;
import tech.hellsoft.trading.enums.Product;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class ProductionUpdateMessage {
    private MessageType type;
    private Product product;
    private Integer quantity;
}
