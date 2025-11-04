package tech.hellsoft.trading.dto.server;

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
public class TickerMessage {
    private MessageType type;
    private Product product;
    private Double bestBid;
    private Double bestAsk;
    private Double mid;
    private Integer volume24h;
    private String serverTime;
}
