package tech.hellsoft.trading.dto.client;

import com.google.gson.annotations.SerializedName;

import tech.hellsoft.trading.enums.MessageType;
import tech.hellsoft.trading.enums.OrderMode;
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
public class OrderMessage {
  private MessageType type;

  @SerializedName("cl_ord_id")
  private String clOrdID;

  private OrderSide side;
  private OrderMode mode;
  private Product product;
  private Integer qty;

  @SerializedName("limit_price")
  private Double limitPrice;

  @SerializedName("expires_at")
  private String expiresAt;

  private String message;

  @SerializedName("debug_mode")
  private String debugMode;
}
