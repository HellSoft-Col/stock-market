package tech.hellsoft.trading.dto.server;

import static org.junit.jupiter.api.Assertions.*;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;

import tech.hellsoft.trading.internal.serialization.JsonSerializer;

/**
 * Tests for TeamRole JSON deserialization to ensure camelCase fields like baseEnergy and
 * levelEnergy are correctly deserialized.
 *
 * <p>This test verifies the fix for the issue where Gson was configured with
 * LOWER_CASE_WITH_UNDERSCORES naming policy, causing baseEnergy and levelEnergy to be null.
 */
class TeamRoleDeserializationTest {

  @Test
  @DisplayName("Should deserialize TeamRole with all fields including baseEnergy and levelEnergy")
  void testTeamRoleDeserialization() {
    // JSON from Go server with camelCase field names
    String json =
        "{"
            + "\"branches\": 2,"
            + "\"maxDepth\": 4,"
            + "\"decay\": 0.75,"
            + "\"budget\": 25.0,"
            + "\"baseEnergy\": 3.0,"
            + "\"levelEnergy\": 2.0"
            + "}";

    TeamRole role = JsonSerializer.fromJson(json, TeamRole.class);

    assertNotNull(role, "TeamRole should not be null");
    assertEquals(2, role.getBranches(), "branches should be 2");
    assertEquals(4, role.getMaxDepth(), "maxDepth should be 4");
    assertEquals(0.75, role.getDecay(), 0.001, "decay should be 0.75");
    assertEquals(25.0, role.getBudget(), 0.001, "budget should be 25.0");

    // These are the critical fields that were returning null before the fix
    assertNotNull(role.getBaseEnergy(), "baseEnergy should not be null");
    assertNotNull(role.getLevelEnergy(), "levelEnergy should not be null");
    assertEquals(3.0, role.getBaseEnergy(), 0.001, "baseEnergy should be 3.0");
    assertEquals(2.0, role.getLevelEnergy(), 0.001, "levelEnergy should be 2.0");
  }

  @Test
  @DisplayName("Should deserialize TeamRole within LoginOKMessage correctly")
  void testTeamRoleInLoginOKMessage() {
    // Simplified LOGIN_OK message JSON from Go server
    String json =
        "{"
            + "\"type\": \"LOGIN_OK\","
            + "\"team\": \"TestTeam\","
            + "\"species\": \"Premium\","
            + "\"initialBalance\": 100000.0,"
            + "\"currentBalance\": 100000.0,"
            + "\"inventory\": {},"
            + "\"authorizedProducts\": [\"FOSFO\"],"
            + "\"recipes\": {"
            + "  \"FOSFO\": {"
            + "    \"type\": \"BASIC\","
            + "    \"ingredients\": {},"
            + "    \"premiumBonus\": 1.0"
            + "  }"
            + "},"
            + "\"role\": {"
            + "  \"branches\": 2,"
            + "  \"maxDepth\": 4,"
            + "  \"decay\": 0.75,"
            + "  \"budget\": 25.0,"
            + "  \"baseEnergy\": 3.0,"
            + "  \"levelEnergy\": 2.0"
            + "},"
            + "\"serverTime\": \"2024-01-01T00:00:00Z\""
            + "}";

    LoginOKMessage message = JsonSerializer.fromJson(json, LoginOKMessage.class);

    assertNotNull(message, "LoginOKMessage should not be null");
    assertNotNull(message.getRole(), "Role should not be null");

    TeamRole role = message.getRole();

    // Verify all role fields are correctly deserialized
    assertEquals(2, role.getBranches(), "branches should be 2");
    assertEquals(4, role.getMaxDepth(), "maxDepth should be 4");
    assertEquals(0.75, role.getDecay(), 0.001, "decay should be 0.75");
    assertEquals(25.0, role.getBudget(), 0.001, "budget should be 25.0");

    // Critical assertion: energy fields should not be null
    assertNotNull(role.getBaseEnergy(), "baseEnergy should not be null in LoginOK context");
    assertNotNull(role.getLevelEnergy(), "levelEnergy should not be null in LoginOK context");
    assertEquals(3.0, role.getBaseEnergy(), 0.001, "baseEnergy should be 3.0");
    assertEquals(2.0, role.getLevelEnergy(), 0.001, "levelEnergy should be 2.0");
  }

  @Test
  @DisplayName("Should handle TeamRole with zero energy values (not null)")
  void testTeamRoleWithZeroEnergyValues() {
    String json =
        "{"
            + "\"branches\": 2,"
            + "\"maxDepth\": 4,"
            + "\"decay\": 0.75,"
            + "\"budget\": 25.0,"
            + "\"baseEnergy\": 0.0,"
            + "\"levelEnergy\": 0.0"
            + "}";

    TeamRole role = JsonSerializer.fromJson(json, TeamRole.class);

    assertNotNull(role, "TeamRole should not be null");

    // Even with 0.0 values, the fields should not be null
    assertNotNull(role.getBaseEnergy(), "baseEnergy should not be null even when 0");
    assertNotNull(role.getLevelEnergy(), "levelEnergy should not be null even when 0");
    assertEquals(0.0, role.getBaseEnergy(), 0.001, "baseEnergy should be 0.0");
    assertEquals(0.0, role.getLevelEnergy(), 0.001, "levelEnergy should be 0.0");
  }

  @Test
  @DisplayName("Should fail with old snake_case naming (regression test)")
  void testOldSnakeCaseNamingWouldFail() {
    // This JSON uses snake_case which the OLD configuration expected
    // With the new IDENTITY naming policy, this should NOT deserialize correctly
    String json =
        "{"
            + "\"branches\": 2,"
            + "\"max_depth\": 4,"
            + "\"decay\": 0.75,"
            + "\"budget\": 25.0,"
            + "\"base_energy\": 3.0,"
            + "\"level_energy\": 2.0"
            + "}";

    TeamRole role = JsonSerializer.fromJson(json, TeamRole.class);

    assertNotNull(role, "TeamRole should not be null");

    // With IDENTITY naming policy, snake_case fields won't map to camelCase Java fields
    // So these should be null (because max_depth != maxDepth, base_energy != baseEnergy)
    assertNull(role.getMaxDepth(), "maxDepth should be null with snake_case JSON");
    assertNull(role.getBaseEnergy(), "baseEnergy should be null with snake_case JSON");
    assertNull(role.getLevelEnergy(), "levelEnergy should be null with snake_case JSON");

    // But fields with exact matches should work
    assertEquals(2, role.getBranches(), "branches should still work");
    assertEquals(0.75, role.getDecay(), 0.001, "decay should still work");
    assertEquals(25.0, role.getBudget(), 0.001, "budget should still work");
  }
}
