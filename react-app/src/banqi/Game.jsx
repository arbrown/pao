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
                    game.tryConnect(document.location.port, params);
                }
            };
            var url = "/playAi?ai=" + this.props.ai;
            xhttp.open("GET", url, true);
            xhttp.send();
        } else {
            var params = { name: this.props.name, id: this.props.id }
            this.tryConnect(document.location.port, params)
        }
    }
    tryConnect(port, params) {

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
        var ws = new WebSocket(addr);

        var comp = this;
        ws.onopen = function () {
            comp.ws = ws;
            this.onmessage = (p1) => comp.handleMessage(p1)
            comp.askForBoard()
        }
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
        chats.push({ player: chatCommand.Player, text: chatCommand.Message, color: chatCommand.Color, auth: chatCommand.Auth })
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
                </div>{player.IsTheirTurn && ! this.props.gameOver && turnIndicator || turnPlaceholder}</h2>
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
