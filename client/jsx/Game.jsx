var Game = React.createClass({
  render: function(){
    return(
    <div>
       <GameState myTurn={this.state.myTurn}
                  gameOver={this.state.gameOver}
                  won={this.state.won}
                  myColor={this.state.myColor}/>
       <Board
          board={this.state.board} myTurn={this.state.myTurn}
          sendMove={this.sendMove}
          myColor={this.state.myColor} />
        <DeadPieces dead={this.state.dead}  />
      <Chat submitChat={this.submitChat} chats={this.state.chats} />
      <button className="goBackButton"><a href="/">Go back to lobby</a></button>
      <button className="resignButton" onClick={this.resign}>Resign</button>
    </div>
    )
  },
  resign: function () {
    if (this.ws){
      var command = {Action: "resign"};
      this.ws.send(JSON.stringify(command));
    }
  },
  sendMove: function(move){
    // sends a ban qi formatted move
    // game will update us if it was valid
    if (this.ws){
      var command = {Action: "move", Argument: move};
      this.ws.send(JSON.stringify(command));
    }
  },
  submitChat: function(text){
    if (this.ws){
      var chat = {Action: "chat", Argument: text};
      this.ws.send(JSON.stringify(chat));
    }
  },
  componentDidMount: function() {
    this.connect();
    this.ws.onopen = this.askForBoard
    React.unmountComponentAtNode(document.getElementById('lobby'));
  },
  connect: function(){
    var params = {name: this.props.name, id: this.props.id}
    var addr = "ws://" +
          document.location.hostname +
          // force port 8000 because of open shift's requirements (for now)
          ":8000" +
          "/game?";
    for (var key in params){
      if (params.hasOwnProperty(key) && params[key]){
        addr += key + "=" + params[key] + "&"
      }
    }

    var ws = new WebSocket(addr);
    ws.onmessage = this.handleMessage
    this.ws = ws;
  },
  askForBoard: function(){
    if (this.ws){
      var command = {Action: 'board?'};
      this.ws.send(JSON.stringify(command));
    }
  },
  handleMessage: function(wsMsg){
    if (!wsMsg || !wsMsg.data) {
      return;
    }
    var data = JSON.parse(wsMsg.data)
    switch (data.Action){
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
  handleBoard: function(boardCommand){
    this.setState({board: boardCommand.Board, myTurn: boardCommand.YourTurn, dead: boardCommand.Dead})
  },
  handleChat: function(chatCommand){
    var chats = this.state.chats;
    chats.push({player: chatCommand.Player, text: chatCommand.Message, color: chatCommand.Color})
    this.setState({chats});
  },
  handleColor: function(colorCommand){
    var myColor = colorCommand.Color;
    this.setState({myColor})
  },
  handleGameOver: function(gameOverCommand){
    this.setState({myTurn: false, gameOver: true, won: gameOverCommand.YouWin});
  },
  getInitialState: function() {
    return {
      chats:[],
      board:
      [['.','.','.','.','.','.','.','.',],
      ['.','.','.','.','.','.','.','.',],
      ['.','.','.','.','.','.','.','.',],
      ['.','.','.','.','.','.','.','.',]],
      myTurn:false,
      myColor: null,
      dead: []
    };
  },
});

var GameState= React.createClass({
  render: function(){
    var headers = []
    if (this.props.gameOver){
      headers.push(<h2 className="game-info-header">Game Over</h2>);
      if (this.props.won){
        headers.push(<h3 className="game-info-subheader">You win!</h3>)
      }
      else {
        headers.push(<h3 className="game-info-subheader">You lose.</h3>)
      }
    }
    else if (this.props.myTurn) {
      headers.push(<h2>Your Turn</h2>)
    } else {
      headers.push(<h2>Opponent's Turn</h2>)
    }
    var cannon;
    if (this.props.myColor == "red"){
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

// React.render(
//   React.createElement(Game, null),
//   document.getElementById('game')
// );
