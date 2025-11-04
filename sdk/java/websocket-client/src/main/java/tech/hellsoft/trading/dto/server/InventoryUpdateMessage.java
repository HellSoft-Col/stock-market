package tech.hellsoft.trading.dto.server;

import java.util.Map;

import tech.hellsoft.trading.enums.MessageType;
import tech.hellsoft.trading.enums.Product;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class InventoryUpdateMessage {
  private MessageType type;
  private Map<Product, Integer> inventory;
  private String serverTime;
}
