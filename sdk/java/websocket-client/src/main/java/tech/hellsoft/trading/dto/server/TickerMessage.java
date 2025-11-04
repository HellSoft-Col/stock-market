package tech.hellsoft.trading.dto.server;

import com.google.gson.annotations.SerializedName;

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
public class TickerMessage {
  private MessageType type;
  private Product product;

  @SerializedName("best_bid")
  private Double bestBid;

  @SerializedName("best_ask")
  private Double bestAsk;

  private Double mid;

  @SerializedName("volume_24h")
  private Integer volume24h;

  @SerializedName("server_time")
  private String serverTime;
}
