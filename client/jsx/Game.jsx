var Game = React.createClass({
    render: function () {
        return (
            <div>
                <GameState myTurn={this.state.myTurn}
                    gameOver={this.state.gameOver}
                    won={this.state.won}
                    myColor={this.state.myColor}
                    whoseTurn={this.state.whoseTurn}
                    turnColor={this.state.turnColor}
                    numPlayers={this.state.numPlayers} />
                <Board
                    board={this.state.board}
                    myTurn={this.state.myTurn}
                    sendMove={this.sendMove}
                    myColor={this.state.myColor}
                    lastMove={this.state.lastMove} />
                <DeadPieces dead={this.state.dead}
                    lastDead={this.state.lastDead}
                    board={this.state.board} />
                <Chat submitChat={this.submitChat} chats={this.state.chats} />
                <button className="goBackButton"><a href="/">Go back to lobby</a></button>
                {!this.state.gameOver ? <button className="resignButton" onClick={this.resign}>Resign</button> : null}
            </div>
        )
    },
    resign: function () {
        if (this.ws) {
            var command = { Action: "resign" };
            this.ws.send(JSON.stringify(command));
        }
    },
    sendMove: function (move) {
        // sends a ban qi formatted move
        // game will update us if it was valid
        if (this.ws) {
            var command = { Action: "move", Argument: move };
            this.ws.send(JSON.stringify(command));
        }
    },
    submitChat: function (text) {
        if (this.ws) {
            var chat = { Action: "chat", Argument: text };
            this.ws.send(JSON.stringify(chat));
        }
    },
    componentDidMount: function () {
        this.connect();
        React.unmountComponentAtNode(document.getElementById('lobby'));
    },
    connect: function () {
        if (this.props.ai) {
            var xhttp = new XMLHttpRequest();
            var game = this;
            xhttp.onreadystatechange = function () {
                if (this.readyState == 4 && this.status == 200) {
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
    },
    tryConnect: function (port, params) {

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
            this.onmessage = comp.handleMessage
            comp.askForBoard()
        }
    },
    askForBoard: function () {
        if (this.ws) {
            var command = { Action: 'board?' };
            this.ws.send(JSON.stringify(command));
        }
    },
    handleMessage: function (wsMsg) {
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
    },
    handleBoard: function (boardCommand) {
        this.setState({
            board: boardCommand.Board,
            myTurn: boardCommand.YourTurn,
            dead: boardCommand.Dead,
            lastMove: boardCommand.LastMove,
            lastDead: boardCommand.LastDead,
            whoseTurn: boardCommand.WhoseTurn,
            turnColor: boardCommand.TurnColor,
            numPlayers: boardCommand.NumPlayers
        })
    },
    handleChat: function (chatCommand) {
        var chats = this.state.chats;
        chats.push({ player: chatCommand.Player, text: chatCommand.Message, color: chatCommand.Color, auth: chatCommand.Auth })
        this.setState({ chats });
    },
    handleColor: function (colorCommand) {
        var myColor = colorCommand.Color;
        this.setState({ myColor })
    },
    handleGameOver: function (gameOverCommand) {
        this.setState({ myTurn: false, gameOver: true, won: gameOverCommand.YouWin });
    },
    getInitialState: function () {
        return {
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
            turnColor: null
        };
    },
});

var GameState = React.createClass({
    render: function () {
        var headers = []
        if (this.props.gameOver) {
            headers.push(<h2 className="game-info-header">Game Over</h2>);
            if (this.props.won) {
                headers.push(<h3 className="game-info-subheader">You win!</h3>)
            } else {
                headers.push(<h3 className="game-info-subheader">You lose.</h3>)
            }
        } else if (this.props.numPlayers < 2) {
            headers.push(<h2 style={{ color: this.props.turnColor }}>Waiting For Opponent</h2>);
        } else {
            headers.push(<h2 style={{ color: this.props.turnColor }}>{this.props.whoseTurn}'s Turn</h2>);
        }
        var cannon;
        if (this.props.myColor == "red") {
            cannon = <div className="banner-piece banqi-square red-cannon" />
        } else {
            cannon = <div className="banner-piece banqi-square black-cannon" />
        }
        return (
            <div className="game-state-banner">
                {cannon}
                <div className="headers">
                    {headers}
                </div>
                {cannon}
            </div>
        )
    }
});
