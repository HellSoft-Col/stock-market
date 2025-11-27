package tech.hellsoft.trading;

import static org.junit.jupiter.api.Assertions.*;

import org.junit.jupiter.api.Test;

import tech.hellsoft.trading.dto.server.*;
import tech.hellsoft.trading.enums.*;
import tech.hellsoft.trading.internal.serialization.JsonSerializer;

/** Quick validation of critical messages that were fixed. */
class QuickMessageValidationTest {

  @Test
  void testTickerMessage_AllFieldsDeserialize() {
    // EXACTLY what Go server sends (camelCase)
    String json =
        "{\"type\":\"TICKER\",\"product\":\"GUACA\","
            + "\"bestBid\":99.5,\"bestAsk\":100.5,\"mid\":100.0,"
            + "\"volume24h\":1000,\"serverTime\":\"2024-01-01T00:00:00Z\"}";

    TickerMessage msg = JsonSerializer.fromJson(json, TickerMessage.class);

    System.out.println("✅ TickerMessage: " + json);
    assertNotNull(msg, "Ticker message should deserialize");
    assertEquals(Product.GUACA, msg.getProduct());
    assertNotNull(msg.getBestBid(), "bestBid should NOT be null");
    assertNotNull(msg.getBestAsk(), "bestAsk should NOT be null");
    assertNotNull(msg.getMid(), "mid should NOT be null");
    assertNotNull(msg.getVolume24h(), "volume24h should NOT be null");
    assertNotNull(msg.getServerTime(), "serverTime should NOT be null");

    assertEquals(99.5, msg.getBestBid(), 0.01);
    assertEquals(100.5, msg.getBestAsk(), 0.01);
    assertEquals(100.0, msg.getMid(), 0.01);
    assertEquals(1000, msg.getVolume24h());
    assertEquals("2024-01-01T00:00:00Z", msg.getServerTime());
  }

  @Test
  void testTickerSerializationFormat() {
    TickerMessage ticker = new TickerMessage();
    ticker.setType(MessageType.TICKER);
    ticker.setProduct(Product.GUACA);
    ticker.setBestBid(99.5);
    ticker.setBestAsk(100.5);
    ticker.setMid(100.0);
    ticker.setVolume24h(1000);
    ticker.setServerTime("2024-01-01T00:00:00Z");

    String json = JsonSerializer.toJson(ticker);

    System.out.println("✅ Ticker serializes to: " + json);
    assertTrue(json.contains("\"bestBid\""), "Should use bestBid not best_bid");
    assertTrue(json.contains("\"bestAsk\""), "Should use bestAsk not best_ask");
    assertTrue(json.contains("\"volume24h\""), "Should use volume24h not volume_24h");
    assertTrue(json.contains("\"serverTime\""), "Should use serverTime not server_time");
  }

  @Test
  void testLoginOKMessage_Deserializes() {
    String json =
        "{\"type\":\"LOGIN_OK\",\"team\":\"TestTeam\",\"species\":\"Premium\","
            + "\"initialBalance\":10000.0,\"currentBalance\":9500.0,"
            + "\"inventory\":{\"GUACA\":50,\"SEBO\":25},"
            + "\"authorizedProducts\":[\"GUACA\",\"SEBO\"],"
            + "\"recipes\":{},"
            + "\"role\":{\"branches\":3,\"maxDepth\":2,\"decay\":0.5,\"budget\":1000.0,\"baseEnergy\":100.0,\"levelEnergy\":50.0},"
            + "\"serverTime\":\"2024-01-01T00:00:00Z\"}";

    LoginOKMessage msg = JsonSerializer.fromJson(json, LoginOKMessage.class);

    System.out.println("✅ LoginOKMessage deserialized");
    assertNotNull(msg);
    assertEquals("TestTeam", msg.getTeam());
    assertEquals(10000.0, msg.getInitialBalance(), 0.01);
    assertEquals(9500.0, msg.getCurrentBalance(), 0.01);
  }

  @Test
  void testFillMessage_Deserializes() {
    String json =
        "{\"type\":\"FILL\",\"clOrdID\":\"ORD-123\",\"fillQty\":10,\"fillPrice\":100.5,"
            + "\"side\":\"BUY\",\"product\":\"GUACA\",\"counterparty\":\"OtherTeam\","
            + "\"counterpartyMessage\":\"Thanks!\",\"serverTime\":\"2024-01-01T00:00:00Z\"}";

    FillMessage msg = JsonSerializer.fromJson(json, FillMessage.class);

    System.out.println("✅ FillMessage deserialized");
    assertNotNull(msg);
    assertEquals("ORD-123", msg.getClOrdID());
    assertEquals(10, msg.getFillQty());
    assertEquals(100.5, msg.getFillPrice(), 0.01);
  }

  @Test
  void testGlobalPerformanceReport_Deserializes() {
    String json =
        "{\"type\":\"GLOBAL_PERFORMANCE_REPORT\","
            + "\"duration\":\"1h\",\"totalTrades\":500,\"totalVolume\":250000.0,"
            + "\"topTraders\":[],\"productStats\":{},"
            + "\"serverTime\":\"2024-01-01T00:00:00Z\"}";

    GlobalPerformanceReportMessage msg =
        JsonSerializer.fromJson(json, GlobalPerformanceReportMessage.class);

    System.out.println("✅ GlobalPerformanceReportMessage deserialized");
    assertNotNull(msg);
    assertEquals("1h", msg.getDuration());
    assertEquals(500, msg.getTotalTrades());
  }
}
