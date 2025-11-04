package tech.hellsoft.trading.dto.client;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import tech.hellsoft.trading.enums.MessageType;
import tech.hellsoft.trading.enums.OrderMode;
import tech.hellsoft.trading.enums.OrderSide;
import tech.hellsoft.trading.enums.Product;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class OrderMessage {
    private MessageType type;
    private String clOrdID;
    private OrderSide side;
    private OrderMode mode;
    private Product product;
    private Integer qty;
    private Double limitPrice;
    private String expiresAt;
    private String message;
    private String debugMode;
}
