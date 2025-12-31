import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { GameService } from "../gen/game/v1/game_connect";

// Create transport
const transport = createConnectTransport({
  baseUrl: import.meta.env.VITE_API_URL?.trim() || window.location.origin,
});

// Create typed client
export const gameClient = createClient(GameService, transport);
