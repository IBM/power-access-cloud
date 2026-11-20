// Must match utils.WebSocketAuthSubprotocol in the API server.
const WS_AUTH_SUBPROTOCOL = "pac.v1";

// openAuthedWebSocket opens a WebSocket authenticated with a bearer token.
// Browsers cannot set an Authorization header on a WebSocket handshake, and
// putting the token in the URL would leak it into proxy and access logs, so it
// is offered in the Sec-WebSocket-Protocol header instead. The server only
// ever echoes back WS_AUTH_SUBPROTOCOL, never the token.
export function openAuthedWebSocket(url, token) {
  return new WebSocket(url, [WS_AUTH_SUBPROTOCOL, token]);
}
