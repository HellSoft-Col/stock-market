package tech.hellsoft.trading.dto.client;

import tech.hellsoft.trading.enums.MessageType;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

/**
 * Message for authenticating with the Stock Market server.
 *
 * <p>This message is used to authenticate the client session using a token. The token is typically
 * provided by the server administrator or obtained through a separate authentication process.
 *
 * <p>Login messages are typically sent automatically by ConectorBolsa.conectar(), but can be sent
 * manually for re-authentication scenarios.
 *
 * @see tech.hellsoft.trading.ConectorBolsa
 * @see tech.hellsoft.trading.dto.server.LoginOKMessage for successful authentication response
 */
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class LoginMessage {
  /** The message type (automatically set to LOGIN). */
  private MessageType type;

  /**
   * Authentication token for the session.
   *
   * <p>This token must be valid and not expired.
   */
  private String token;

  /** Timezone for the session (typically "UTC"). */
  private String tz;
}
