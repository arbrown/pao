var Game = React.createClass({
  render: function(){
    return(
     <div>
       <h2>{this.state.myTurn? "My Turn":"Not My Turn"}</h2>
       <Board
          board={this.state.board} myTurn={this.state.myTurn}
          sendMove={this.sendMove}
          myColor={this.state.myColor} />
       <Chat submitChat={this.submitChat} chats={this.state.chats} />
       <button className="reset-button" onClick={this.resetGame}>Reset Game</button>
     </div>
    )
  },
  sendMove: function(move){
    // sends a ban chi formatted move
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
  },
  connect: function(){
    var addr = "ws://" +
          document.location.host +
          "/game?id=20" ;
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
      default:
        console.log("I don't know what to do with this...");
        console.log(data);
    }
  },
  handleBoard: function(boardCommand){
    this.setState({board: boardCommand.Board, myTurn: boardCommand.YourTurn})
  },
  handleChat: function(chatCommand){
    var chats = this.state.chats;
    chats.push({person: chatCommand.Person, text: chatCommand.Message})
    this.setState({chats});
  },
  handleColor: function(colorCommand){
    var myColor = colorCommand.Color;
    this.setState({myColor})
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
      myColor: null
    };
  },
});

React.render(
  React.createElement(Game, null),
  document.getElementById('game')
);
