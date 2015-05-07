var Game = React.createClass({
  render: function(){
    return(
     <div>
       <Board />
       <Chat />
       <button className="reset-button" onClick={this.resetGame}>Reset Game</button>
     </div>
    )
  }
});

React.render(
  React.createElement(Game, null),
  document.getElementById('game')
);
