import React from 'react';
import Board from './Board.jsx'
import DeadPieces from './Dead.jsx'
import Chat from './Chat.jsx'

export default class Game extends React.Component {
    constructor(props) {
        super(props)
        this.state = {
            chats: [],
            board:
                [['.', '.', '.', '.', '.', '.', '.', '.',],
                ['.', '.', '.', '.', '.', '.', '.', '.',],
                ['.', '.', '.', '.', '.', '.', '.', '.',],
                ['.', '.', '.', '.', '.', '.', '.', '.',]],
            myTurn: false,
            myColor: null,
            dead: [],
            numPlayers: 0,
            whoseTurn: null,
            turnColor: null,
            gameOverReason: "",
        };
        this.sessionId = null;
        this.gameId = null;
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 10;
        this.reconnectDelay = 1000; // Start with 1 second
        this.isReconnecting = false;
        this.connectionPort = null;
        this.connectionParams = null;
        this.wasAttemptingRejoin = false; // Track if we loaded session from storage
    }
    render() {
        return (
            <div>
                <GameState myTurn={this.state.myTurn}
                    gameOver={this.state.gameOver}
                    gameOverMessage={this.state.gameOverMessage}
                    gameOverReason={this.state.gameOverReason}
                    won={this.state.won}
                    myColor={this.state.myColor}
                    players={this.state.players}
                    player={this.state.player} />
                <Board
                    board={this.state.board}
                    myTurn={this.state.myTurn}
                    sendMove={this.sendMove.bind(this)}
                    myColor={this.state.myColor}
                    lastMove={this.state.lastMove}
                    firstMove={this.state.firstMove} />
                <DeadPieces dead={this.state.dead}
                    lastDead={this.state.lastDead}
                    board={this.state.board} />
                <Chat submitChat={this.submitChat.bind(this)} chats={this.state.chats} />
                <button className="goBackButton"><a href="/">Go back to lobby</a></button>
                {!this.state.gameOver ? <button className="resignButton" onClick={(e) => this.resign(e)}>Resign</button> : null}
            </div>
        )
    }
    resign() {
        if (this.ws) {
            var command = { Action: "resign" };
            this.ws.send(JSON.stringify(command));
        }
    }
    sendMove(move) {
        // sends a ban qi formatted move
        // game will update us if it was valid
        if (this.ws) {
            var command = { Action: "move", Argument: move };
            this.ws.send(JSON.stringify(command));
        }
    }
    submitChat(text) {
        if (this.ws) {
            var chat = { Action: "chat", Argument: text };
            this.ws.send(JSON.stringify(chat));
        }
    }
    componentDidMount() {
        this.connect();
    }
    connect() {
        if (this.props.ai) {
            var xhttp = new XMLHttpRequest();
            var game = this;
            xhttp.onreadystatechange = function () {
                if (this.readyState === 4 && this.status === 200) {
                    var r = JSON.parse(this.response);
                    params = { name: game.props.name, id: r.ID }
                    
                    // Check localStorage for existing session
                    game.loadSessionFromStorage(r.ID);
                    
                    game.tryConnect(document.location.port, params);
                }
            };
            var url = "/playAi?ai=" + this.props.ai;
            xhttp.open("GET", url, true);
            xhttp.send();
        } else {
            var params = { name: this.props.name, id: this.props.id }
            
            // Check localStorage for existing session before connecting
            this.loadSessionFromStorage(this.props.id);
            
            this.tryConnect(document.location.port, params)
        }
    }
    
    loadSessionFromStorage(gameId) {
        if (!gameId) return;
        
        const sessionKey = 'pao_session_' + gameId;
        const storedSessionId = localStorage.getItem(sessionKey);
        
        if (storedSessionId) {
            console.log("Found existing session in localStorage:", storedSessionId);
            this.sessionId = storedSessionId;
            this.gameId = gameId;
            this.wasAttemptingRejoin = true; // Mark that we're attempting to rejoin
            
            // Add a chat message indicating we're trying to reconnect
            var chats = this.state.chats;
            chats.push({ 
                player: "System", 
                text: "Attempting to rejoin game...", 
                color: "blue", 
                timestamp: new Date() 
            });
            this.setState({ chats });
        }
    }
    tryConnect(port, params) {
        // Store connection parameters for reconnection
        this.connectionPort = port;
        this.connectionParams = Object.assign({}, params); // Make a copy
        
        // Add sessionId if we have one (for reconnection)
        if (this.sessionId) {
            console.log("Adding sessionId to connection params:", this.sessionId);
            params.sessionId = this.sessionId;
        } else {
            console.log("No sessionId to add to connection params");
        }

        var addr = "ws://" +
            document.location.hostname
            + ':'
            + port
            + "/game?";
        for (var key in params) {
            if (params.hasOwnProperty(key) && params[key]) {
                addr += key + "=" + params[key] + "&"
            }
        }
        
        console.log("Connecting to:", addr);
        console.log("Connection params:", params);
        var ws = new WebSocket(addr);

        var comp = this;
        ws.onopen = function () {
            console.log("WebSocket connected");
            comp.ws = ws;
            comp.reconnectAttempts = 0; // Reset reconnection attempts on successful connection
            comp.reconnectDelay = 1000; // Reset delay
            comp.isReconnecting = false;
            
            this.onmessage = (p1) => comp.handleMessage(p1)
            
            // Set up onclose handler for reconnection
            this.onclose = function(event) {
                console.log("WebSocket closed", event);
                comp.handleDisconnection();
            }
            
            this.onerror = function(error) {
                console.log("WebSocket error", error);
            }
            
            comp.askForBoard()
        }
        
        // Handle connection failure
        ws.onerror = function(error) {
            console.log("WebSocket connection error", error);
        }
    }
    
    handleDisconnection() {
        // Don't reconnect if the game is over
        if (this.state.gameOver) {
            console.log("Game is over, not reconnecting");
            return;
        }
        
        // Don't start multiple reconnection attempts
        if (this.isReconnecting) {
            console.log("Already attempting to reconnect");
            return;
        }
        
        this.isReconnecting = true;
        
        // Check if we've exceeded max reconnection attempts
        if (this.reconnectAttempts >= this.maxReconnectAttempts) {
            console.log("Max reconnection attempts reached");
            var chats = this.state.chats;
            chats.push({ 
                player: "System", 
                text: "Connection lost. Unable to reconnect after " + this.maxReconnectAttempts + " attempts.", 
                color: "red", 
                timestamp: new Date() 
            });
            this.setState({ chats });
            return;
        }
        
        this.reconnectAttempts++;
        console.log("Attempting to reconnect (attempt " + this.reconnectAttempts + ")");
        
        // Add system message about reconnection
        var reconnectChats = this.state.chats;
        reconnectChats.push({ 
            player: "System", 
            text: "Connection lost. Reconnecting... (attempt " + this.reconnectAttempts + ")", 
            color: "orange", 
            timestamp: new Date() 
        });
        this.setState({ chats: reconnectChats });
        
        // Wait with exponential backoff before reconnecting
        setTimeout(() => {
            console.log("Reconnecting now...");
            // Use stored connection parameters
            if (this.connectionPort && this.connectionParams) {
                this.tryConnect(this.connectionPort, this.connectionParams);
            }
            // Exponential backoff with max of 30 seconds
            this.reconnectDelay = Math.min(this.reconnectDelay * 1.5, 30000);
        }, this.reconnectDelay);
    }
    askForBoard() {
        if (this.ws) {
            var command = { Action: 'board?' };
            this.ws.send(JSON.stringify(command));
        }
    }
    handleMessage(wsMsg) {
        if (!wsMsg || !wsMsg.data) {
            return;
        }
        var data = JSON.parse(wsMsg.data)
        switch (data.Action) {
            case 'session':
                this.handleSession(data);
                break;
            case 'chat':
                this.handleChat(data);
                break;
            case 'board':
                this.handleBoard(data);
                break;
            case 'color':
                this.handleColor(data);
                break;
            case 'gameover':
                this.handleGameOver(data);
                break;
            default:
                console.log("I don't know what to do with this...");
                console.log(data);
        }
    }
    handleSession(sessionCommand) {
        // Store session ID for reconnection
        this.sessionId = sessionCommand.SessionID;
        this.gameId = sessionCommand.GameID;
        console.log("Received session ID:", this.sessionId, "for game:", this.gameId);
        
        // Persist session ID to localStorage so it survives page reloads
        if (this.sessionId && this.gameId) {
            const sessionKey = 'pao_session_' + this.gameId;
            localStorage.setItem(sessionKey, this.sessionId);
            console.log("Saved session to localStorage:", sessionKey);
            
            // If we were attempting to rejoin and received session confirmation, show success
            if (this.wasAttemptingRejoin) {
                var chats = this.state.chats;
                chats.push({ 
                    player: "System", 
                    text: "Successfully rejoined game!", 
                    color: "green", 
                    timestamp: new Date() 
                });
                this.setState({ chats });
                this.wasAttemptingRejoin = false; // Reset flag
            }
        }
    }
    handleBoard(boardCommand) {
        this.setState({
            board: boardCommand.Board,
            myTurn: boardCommand.YourTurn,
            dead: boardCommand.Dead,
            lastMove: boardCommand.LastMove,
            lastDead: boardCommand.LastDead,
            firstMove: boardCommand.FirstMove,
            players: boardCommand.Players,
            player: boardCommand.Player,
        })
    }
    handleChat(chatCommand) {
        var chats = this.state.chats;
        chats.push({ player: chatCommand.Player, text: chatCommand.Message, color: chatCommand.Color, auth: chatCommand.Auth, timestamp: new Date() })
        this.setState({ chats });
    }
    handleColor(colorCommand) {
        var myColor = colorCommand.Color;
        this.setState({ myColor })
    }
    handleGameOver(gameOverCommand) {
        // Ignore spurious gameOverCommand
        if(this.state.gameOver) {
            return;
        }
        this.setState({ myTurn: false, gameOver: true, won: gameOverCommand.YouWin, gameOverMessage: gameOverCommand.Message, gameOverReason: gameOverCommand.Reason });
        
        // Clear session from localStorage when game ends
        if (this.gameId) {
            const sessionKey = 'pao_session_' + this.gameId;
            localStorage.removeItem(sessionKey);
            console.log("Cleared session from localStorage:", sessionKey);
        }
    }

}

class GameState extends React.Component {
    render() {
        let cannons = {
            "red": <div className="banner-piece banqi-square red-cannon"/>,
            "black": <div className="banner-piece banqi-square black-cannon"/>,
            "green": <div className="banner-piece banqi-square unflipped-piece"/>,
        }

        let turnIndicator = <div class="turn-indicator">🔼 TURN 🔼</div>
        let turnPlaceholder = <div class="turn-indicator placeholder">&nbsp;</div>

        let playerHeaders = []
        for (let i=0; i<this.props.players?.length; ++i) {
            let player = this.props.players[i]
            let name = <span>{player.Name}</span>
            let h =
                <h2><div class="middle-valign-container"> 
                    {cannons[player.Color]}
                    {name}
                    {cannons[player.Color]}
                </div>{(player.IsTheirTurn && !this.props.gameOver) ? turnIndicator : turnPlaceholder}</h2>
            playerHeaders.push(h);
        }
        if (this.props.players?.length < 2) {
            let h =
                <h2><div class="middle-valign-container">
                    {cannons["green"]}
                    <span>Waiting For Opponent</span>
                    {cannons["green"]}
                </div>{turnPlaceholder}</h2>
            playerHeaders.push(h);
        }

        let headers = []
        if (this.props.gameOver) {
            headers.push(<h2 className="game-info-header">Game Over</h2>);
            headers.push(<h2 className="game-info-header">{this.props.gameOverMessage} -- {this.props.gameOverReason}</h2>);
        }
        headers = headers.concat(playerHeaders)

        return (
            <div className="game-state-banner">
                <div className="headers">
                    {headers}
                </div>
            </div>
        )
    }
}
