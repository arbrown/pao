package game_test

import (
	"testing"
	"time"

	"github.com/arbrown/pao/game"
	"github.com/arbrown/pao/game/command"
	"github.com/arbrown/pao/game/player"
	// "github.com/gorilla/websocket" // Not strictly needed for mock if not using websocket constants
)

// mockConn is a mock WebSocket connection for testing purposes.
// It implements the interface implicitly expected by player.Ws, specifically WriteJSON.
type mockConn struct {
	received chan interface{}
}

// newMockConn creates a new mockConn with a buffered channel.
func newMockConn(bufferSize int) *mockConn {
	return &mockConn{
		received: make(chan interface{}, bufferSize),
	}
}

// WriteJSON sends the message to the received channel.
func (m *mockConn) WriteJSON(v interface{}) error {
	if m.received == nil {
		// This case should ideally not happen if mockConn is always initialized with newMockConn
		// For robustness, especially if tests evolve, good to have.
		panic("mockConn.received channel is nil")
	}
	// Non-blocking send to prevent test hangs if channel is full,
	// though buffered channel should prevent this for expected message counts.
	select {
	case m.received <- v:
		return nil
	default:
		// This would indicate the channel buffer is full, which might be a test setup issue.
		// For now, let's assume the buffer is sufficient.
		// If tests become flaky, this might need a different strategy (e.g., error or larger buffer).
		panic("mockConn.received channel is full or closed")
	}
}

// ReadJSON is a no-op for these tests, as we are not testing incoming messages to the server via this mock.
func (m *mockConn) ReadJSON(v interface{}) error {
	return nil
}

// Close closes the received channel.
func (m *mockConn) Close() error {
	if m.received != nil {
		// Safely close the channel only if it's not already closed.
		// This defer-recover is a common Go idiom for safe closing.
		defer func() {
			recover() // Recover from panic if closing an already closed channel.
		}()
		close(m.received)
	}
	return nil
}

// NextReader is a no-op, not used in the tested code path.
func (m *mockConn) NextReader() (int, error) {
	return 0, nil
}

func TestTauntCommand(t *testing.T) {
	// Subtest for a regular player (non-kibitzer) sending a taunt
	t.Run("RegularPlayerTaunts", func(t *testing.T) {
		removeGameChan := make(chan *game.Game, 1) // Buffered to prevent blocking
		g := game.NewGame("testgame_regular", removeGameChan, nil)

		// Setup Player 1 (Sender)
		p1Conn := newMockConn(2) // Buffer for 1 message (sender also receives broadcast)
		p1 := player.NewPlayer(nil, "P1_Taunter", nil, false, false) // false for Kibitzer
		p1.Ws = p1Conn // Assign our mock connection

		// Setup Player 2 (Receiver)
		p2Conn := newMockConn(1) // Buffer for 1 message
		p2 := player.NewPlayer(nil, "P2_Regular", nil, false, false)
		p2.Ws = p2Conn

		g.Players = []*player.Player{p1, p2}
		g.CurrentPlayerIndex = 0 // p1 is the current player

		// Color expectation:
		// The color logic in game's handleSlashCommand for /taunt is:
		//   color := "black"
		//   if c.P == g.red { color = "red" }       // g.red is an unexported field in game.Game
		//   if c.P.Kibitzer == true { color = "teal" }
		// Since p1 is not a kibitzer and g.red (unexported) cannot be directly set to p1
		// from this game_test package, c.P == g.red will be false.
		// Thus, the color is expected to be the default "black".
		expectedColor := "black"
		expectedSenderName := "P1_Taunter"

		cmd := command.PlayerCommand{
			C: command.Command{Action: "chat", Argument: "/taunt"},
			P: p1, // Player 1 (regular player) sends the taunt
		}
		g.HandleCommand(cmd) // This function processes the command

		// Assert messages for both players (sender and receiver)
		playersToTest := []struct {
			name   string
			conn   *mockConn
		}{
			{"Sender_P1", p1Conn},
			{"Receiver_P2", p2Conn},
		}

		for _, pt := range playersToTest {
			t.Run(pt.name, func(t *testing.T) { // Sub-subtest for each player
				select {
				case receivedMsg := <-pt.conn.received:
					chatCmd, ok := receivedMsg.(command.ChatCommand)
					if !ok {
						t.Fatalf("Expected ChatCommand, got %T", receivedMsg)
					}
					if chatCmd.Player != expectedSenderName {
						t.Errorf("Expected sender name '%s', got '%s'", expectedSenderName, chatCmd.Player)
					}
					if chatCmd.Color != expectedColor {
						t.Errorf("Expected color '%s', got '%s'", expectedColor, chatCmd.Color)
					}
					foundTaunt := false
					for _, taunt := range game.Taunts { // game.Taunts was exported in a previous step
						if chatCmd.Message == taunt {
							foundTaunt = true
							break
						}
					}
					if !foundTaunt {
						t.Errorf("Received message '%s' is not a known taunt from game.Taunts", chatCmd.Message)
					}
				case <-time.After(200 * time.Millisecond): // Timeout for receiving message
					t.Errorf("Timed out waiting for message")
				}
			})
		}
	})

	// Subtest for a kibitzer player sending a taunt
	t.Run("KibitzerPlayerTaunts", func(t *testing.T) {
		removeGameChan := make(chan *game.Game, 1)
		g := game.NewGame("testgame_kibitzer", removeGameChan, nil)

		// Setup Kibitzer (Sender)
		kibitzerConn := newMockConn(1) // Kibitzer sender does not receive their own message unless in g.kibitzers list
		kibitzerSender := player.NewPlayer(nil, "K1_Taunter", nil, true, false) // true for Kibitzer
		kibitzerSender.Ws = kibitzerConn

		// Setup Regular Players (Receivers)
		p1RegularConn := newMockConn(1)
		p1Regular := player.NewPlayer(nil, "P1_Regular", nil, false, false)
		p1Regular.Ws = p1RegularConn

		p2RegularConn := newMockConn(1)
		p2Regular := player.NewPlayer(nil, "P2_Regular", nil, false, false)
		p2Regular.Ws = p2RegularConn
		
		g.Players = []*player.Player{p1Regular, p2Regular} // Active players in the game
		g.CurrentPlayerIndex = 0
		// Note: kibitzerSender is NOT added to g.Kibitzers (unexported list) via g.JoinKibitz here.
		// We are unit-testing HandleCommand with a command from a player marked as Kibitzer.
		// The broadcast logic will send to g.Players and g.kibitzers.

		expectedColor := "teal" // Kibitzers' taunts should be "teal"
		expectedSenderName := "K1_Taunter"

		cmd := command.PlayerCommand{
			C: command.Command{Action: "chat", Argument: "/taunt"},
			P: kibitzerSender, // Kibitzer sends the taunt
		}
		g.HandleCommand(cmd)

		// Assert messages for regular players (receivers)
		receiversToTest := []struct {
			name   string
			conn   *mockConn
		}{
			{"Receiver_P1_Regular", p1RegularConn},
			{"Receiver_P2_Regular", p2RegularConn},
		}

		for _, pt := range receiversToTest {
			t.Run(pt.name, func(t *testing.T) {
				select {
				case receivedMsg := <-pt.conn.received:
					chatCmd, ok := receivedMsg.(command.ChatCommand)
					if !ok {
						t.Fatalf("Expected ChatCommand, got %T", receivedMsg)
					}
					if chatCmd.Player != expectedSenderName {
						t.Errorf("Expected sender name '%s', got '%s'", expectedSenderName, chatCmd.Player)
					}
					if chatCmd.Color != expectedColor {
						t.Errorf("Expected color '%s', got '%s'", expectedColor, chatCmd.Color)
					}
					// Taunt message content check
					foundTaunt := false
					for _, taunt := range game.Taunts {
						if chatCmd.Message == taunt {
							foundTaunt = true
							break
						}
					}
					if !foundTaunt {
						t.Errorf("Received message '%s' is not a known taunt from game.Taunts", chatCmd.Message)
					}
				case <-time.After(200 * time.Millisecond):
					t.Errorf("Timed out waiting for message from kibitzer")
				}
			})
		}
		
		// Assert that the kibitzer sender does NOT receive their own message
		// (because they were not added to g.kibitzers list through g.JoinKibitz for this specific test structure)
		t.Run("KibitzerSender_SelfReceiptCheck", func(t *testing.T) {
			select {
			case receivedMsgKib := <-kibitzerConn.received:
				t.Errorf("Kibitzer sender unexpectedly received their own message: %+v. This implies they were in g.kibitzers or broadcast logic includes sender by default.", receivedMsgKib)
			case <-time.After(50 * time.Millisecond): // Short timeout, expecting no message
				// This is the expected behavior for this test setup.
			}
		})
	})
}
