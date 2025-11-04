package tech.hellsoft.trading.dto.client;

import tech.hellsoft.trading.enums.MessageType;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class AcceptOfferMessage {
  private MessageType type;
  private String offerId;
  private Boolean accept;
  private Integer quantityOffered;
  private Double priceOffered;
}
