pao
===

A [Banqi](http://en.wikipedia.org/wiki/Banqi) Game Server

Installing Pao
--------------
### Prerequisites
1. [Install Go](https://golang.org/doc/install) (version 1.21 or later)
   * Modern Go uses modules, so you don't need to set up `$GOPATH` or `$GOROOT` anymore
2. [Install Node.js](https://nodejs.org/) (for building the React frontend)
   * Includes npm, which is needed for the web UI

### Running from source
1. Clone this repository:
   ```bash
   git clone https://github.com/arbrown/pao.git
   cd pao
   ```
2. Install Go dependencies:
   ```bash
   go mod download
   ```
3. Create the configuration file:
   ```bash
   cp conf/paoSettings.json.sample conf/paoSettings.json
   ```
   This creates a basic SQLite configuration. You can edit `conf/paoSettings.json` to customize database settings and encryption keys.

4. Build the React frontend:
   ```bash
   cd react-app
   npm install
   npm run build
   cd ..
   ```

5. Build and run the server:
   ```bash
   go run .
   ```
   Or build a binary:
   ```bash
   go build -o pao .
   ./pao
   ```
   The server will start on port 2015. Go to http://localhost:2015/ in a web browser to see the lobby.

### Installing as a command
You can install Pao globally using:
```bash
go install github.com/arbrown/pao@latest
```
Then run it with:
```bash
pao
```

### Development
To contribute changes:
1. Fork this repository on GitHub
2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR_USERNAME/pao.git
   cd pao
   ```
3. Make your changes and test them
4. Submit a pull request to the upstream repository

Joining a game
--------------
If no games are currently running, you can join a new game by clicking the button:
![join](./screenshots/join-game.png)  
Then, another player can join your game and it will begin.
![join](./screenshots/join-existing.png)  
Have fun!
![game](./screenshots/game2.png)  
Banqi Game Notation
---------------------
The game code uses a character-based notation to store and transmit piece information.

### Pieces
Pieces are represented by the letters in the following table

| Piece | Red | Black |
|:------|:---:|:-----:|
| King  | k   | K     |
| Guard | g   | G     |
| Elephant| e | E     |
| Cart  | c   | C     |
| Horse | h   | H     |
| Pawn  | p   | P     |
| Canon | q   | Q     |

Uncovered pieces are represented by '?' and empty squares are '.'.

### Board
The board is represented as an 8x4 (8 columns, 4 rows) board with coordinates as follows:

| .   | A | B | C | D | E | F | G | H |
|----:|:-:|:-:|:-:|:-:|:-:|:-:|:-:|:-:|
|**1**|   |   |   |   |   |   |   |   |
|**2**|   |   |   |   |   |   |   |   |
|**3**|   |   |   |   |   |   |   |   |
|**4**|   |   |   |   |   |   |   |   |

A location is referred to by its letter/number coordinates, for example, A1 or F3.

### Plies
A ply consists of a piece identifier, a location, an action (moves '>', or becomes '=') followed by a piece identifier, and a location.  If a piece was killed in this ply, there is then an 'x' followed by the piece identifier and location of the killed piece.

For example, if a black guard at C2 killed a red pawn at D2, the ply notation would be:

    GC2>GD2xpD2

If a player turned up a red cannon at B4, the notation would be:

    ?B4=qB4

#### Theoretical variants
In theory, in some game variants a ply could involve more than one piece (on either side) or multiple moves (as in double-move Banqi.)  In this case, multiple pieces (and their locations) are separated by commas (,) and multiple moves are separated by semicolons (;).
