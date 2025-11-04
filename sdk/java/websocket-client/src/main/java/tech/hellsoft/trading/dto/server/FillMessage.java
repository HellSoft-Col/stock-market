package tech.hellsoft.trading.dto.server;

import tech.hellsoft.trading.enums.MessageType;
import tech.hellsoft.trading.enums.OrderSide;
import tech.hellsoft.trading.enums.Product;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class FillMessage {
  private MessageType type;
  private String clOrdID;
  private Integer fillQty;
  private Double fillPrice;
  private OrderSide side;
  private Product product;
  private String counterparty;
  private String counterpartyMessage;
  private String serverTime;
  private Integer remainingQty;
  private Integer totalQty;
}
