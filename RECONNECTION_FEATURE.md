# WebSocket Reconnection Feature

## Overview
This implementation adds automatic reconnection support to the Pao game server. When a player's network connection is disrupted, the game will no longer immediately end. Instead, the server gives a 2-minute grace period for the player to reconnect.

## How It Works

### Server-Side Changes

1. **Session IDs**: Each player is assigned a unique session ID when they join a game. This ID is sent to the client and used to identify returning players.

2. **Disconnection Grace Period**: When a player disconnects:
   - The player is marked as "disconnected" rather than immediately losing
   - Other players are notified that the player disconnected
   - A 2-minute timer starts
   - If the player reconnects within 2 minutes, they rejoin seamlessly
   - If they don't reconnect, the game ends and the opponent wins

3. **Reconnection Process**: When a player tries to reconnect:
   - They send their session ID with the connection request
   - The server looks up the disconnected player by session ID
   - If found and within the timeout, the player's websocket is updated
   - The game continues from where it left off

### Client-Side Changes

1. **Session Persistence**: The React client stores the session ID in browser localStorage
   - Session IDs are saved as `pao_session_<gameId>`
   - Persists across page reloads and browser tab closures
   - Automatically cleared when the game ends

2. **Automatic Reconnection**: When the websocket closes:
   - The client automatically attempts to reconnect
   - Uses exponential backoff (1s, 1.5s, 2.25s, etc. up to 30s)
   - Makes up to 10 reconnection attempts
   - Shows status messages in the chat

3. **Page Reload Support**: When a player reloads the page or reopens a closed tab:
   - The client checks localStorage for an existing session ID
   - If found, automatically includes it in the connection request
   - Server recognizes the session and seamlessly reconnects the player
   - Shows "Attempting to rejoin game..." and "Successfully rejoined game!" messages

4. **User Experience**:
   - Players see "Connection lost. Reconnecting..." messages during network issues
   - "Attempting to rejoin game..." when loading with a saved session
   - "Successfully rejoined game!" when reconnection succeeds
   - The game state is restored immediately upon reconnection

## Configuration

### Server Settings
- **Reconnection timeout**: 2 minutes (configurable in `markPlayerDisconnected`)
- Located in: `game/game.go` line ~781

### Client Settings
- **Max reconnection attempts**: 10
- **Initial retry delay**: 1 second
- **Max retry delay**: 30 seconds
- Located in: `react-app/src/banqi/Game.jsx` constructor

## Technical Details

### Modified Files

**Server:**
- `game/player/player.go` - Added `SessionID` field to Player struct
- `game/game.go` - Added reconnection logic, session management, and grace period handling
- `game/command/command.go` - Added `SessionCommand` to send session IDs to clients
- `pao.go` - Updated to accept and pass session IDs from query parameters

**Client:**
- `react-app/src/banqi/Game.jsx` - Added reconnection logic with exponential backoff

### Key Functions

**Server:**
- `generateSessionID()` - Creates unique session IDs using crypto/rand
- `tryReconnectPlayer()` - Attempts to reconnect a player using their session ID
- `markPlayerDisconnected()` - Marks a player as disconnected with a grace period
- `sendSessionID()` - Sends session ID to the client

**Client:**
- `handleDisconnection()` - Manages reconnection attempts with exponential backoff
- `handleSession()` - Stores session ID received from server
- `tryConnect()` - Updated to include session ID in reconnection attempts

## Deployment

After making changes to the reconnection code:

1. **Rebuild the React app**:
   ```bash
   cd react-app
   npm run build
   ```

2. **Restart the Go server** to pick up the new build:
   ```bash
   cd ..
   ./pao  # or however you run your server
   ```

## Testing the Feature

### Test 1: Network Disruption
1. Start a game with two players
2. Simulate network disruption (disable WiFi/ethernet for a few seconds)
3. Re-enable network
4. The client should automatically reconnect and continue playing
5. Check browser console for reconnection logs

### Test 2: Page Reload ✅ **Recommended Test**
1. Start a game with two players
2. Reload the browser page (F5 or Cmd+R)
3. The page should automatically rejoin the game
4. You'll see in chat: "Attempting to rejoin game..." followed by "Successfully rejoined game!"
5. The game continues from where it left off
6. Check browser console - you should see the session ID being loaded from localStorage

### Test 3: Tab Closure and Reopening
1. Start a game with two players
2. Note the game URL (e.g., `http://localhost:2015/?name=YourName&id=123`)
3. Close the browser tab
4. Within 2 minutes, open a new tab and navigate to the same URL
5. The game should automatically reconnect you to your previous position
6. Check browser DevTools → Application → Local Storage to see `pao_session_<gameId>`

### Test 4: Timeout
1. Start a game with two players
2. Close the browser tab or disconnect
3. Wait more than 2 minutes
4. Try to reconnect - the game should have ended, and you'll need to start a new game

## Troubleshooting

### Reconnection Not Working After Page Reload

**Symptoms:**
- Server logs show: `sessionId:` (empty)
- Player becomes a kibitzer instead of reconnecting
- Browser console shows: "I don't know what to do with this..." for session command

**Solution:**
The React app needs to be rebuilt after code changes:
```bash
cd react-app
npm run build
cd ..
# Restart your Go server
```

### Session ID Not Found

**Symptoms:**
- Server logs show: `tryReconnectPlayer: session X not found in disconnectedPlayers`

**Possible causes:**
1. More than 2 minutes have passed since disconnection
2. The game ended while you were disconnected
3. The server was restarted (clears all sessions)

### Debugging Tips

**Check localStorage in browser:**
1. Open DevTools (F12)
2. Go to Application → Local Storage
3. Look for keys like `pao_session_<gameId>`

**Check browser console:**
- Should see: "Adding sessionId to connection params: <id>"
- Should see: "Received session ID: <id> for game: <gameId>"

**Check server logs:**
- Should see: "Received connection request - id: X, name: Y, sessionId: Z"
- Should see: "Player X reconnecting with session Y"

## Notes

- Bots and kibitzers do not use session IDs or reconnection
- If a game ends normally, reconnection attempts are cancelled
- Session IDs are only valid for the duration of a game
- The disconnection grace period applies only to human players, not AI opponents
- Session IDs are stored in browser localStorage as `pao_session_<gameId>`
- Sessions are automatically cleaned up from localStorage when a game ends
- Multiple games can have different session IDs stored simultaneously
- Clearing browser data/localStorage will prevent page reload reconnection (but network disruption reconnection still works within the same browser session)

