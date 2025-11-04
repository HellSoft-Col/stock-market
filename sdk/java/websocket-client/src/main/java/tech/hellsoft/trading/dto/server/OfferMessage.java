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
public class OfferMessage {
    private MessageType type;
    private String offerId;
    private String buyer;
    private Product product;
    private Integer quantityRequested;
    private Double maxPrice;
    private Integer expiresIn;
    private String timestamp;
}
