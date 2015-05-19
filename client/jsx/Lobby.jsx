var Lobby = React.createClass({
  joinNew: function(){
    React.render(
      React.createElement(Game, {name: this.state.name}),
      document.getElementById('game')
    );
  },
  join: function(id){
    React.render(
      React.createElement(Game, {name: this.state.name, id: id}),
      document.getElementById('game')
    );
  },
  nameChanged: function(){
    var name = this.refs.name.getDOMNode().value;
    this.setState({name});
  },
  render: function(){
    var comp = this;
    var games = this.state.games.map(function(g) {
        return <LobbyGame id={g.ID} players={g.Players}
          onClick={function(){comp.join(g.ID)}} />
      });
    var gameCount = games ? games.length : 0;
    return (
      <div className="lobby">
        <span id="forkongithub">
          <a href="https://github.com/arbrown/pao">
            Fork me on GitHub
          </a>
        </span>
        <h2>Pao Lobby</h2>
        <input type="text" ref="name" value={this.state.name} onChange={this.nameChanged} placeholder="Your Name" />
        <div className="lobby-current-games">
          <h3>{gameCount} Current Game{gameCount==1 ? "" : "s"}</h3>
          <ul className="games">
            {games}
          </ul>
        </div>
        <button onClick = {this.joinNew}>Join New Game</button>
      </div>
    )
  },
  Reload: function(){
    var comp = this;
    var xhr = new XMLHttpRequest();
    xhr.onreadystatechange = function() {
      if (xhr.readyState==4 && xhr.status == 200){
        var data = JSON.parse(xhr.responseText)
        comp.setState({games: data});
      }
    }
    xhr.open("GET", "/listGames", true);
    xhr.send();
  },
  componentDidMount: function() {
    var comp = this;
    var xhr = new XMLHttpRequest();
    xhr.onreadystatechange = function() {
      if (xhr.readyState==4 && xhr.status == 200){
        var data = JSON.parse(xhr.responseText)
        if (data){
          comp.setState({games: data});
        }
      }
    }
    xhr.open("GET", "/listGames", true);
    xhr.send();
  },
  getInitialState: function() {
    return {
      name: null,
      games: []
    };
  },
});

var LobbyGame = React.createClass({
  render: function() {
    var playerList = this.props.players.map(function(player){
      return (<li className="player">{player}</li>);
    });
    return (
      <li className="lobby-game" onClick={this.props.onClick}>
        <div className="banqi-square red-cannon"/>
        <p>Current Players:</p>
        <ul>{playerList}</ul>
      </li>);
  },
});

setTimeout(function() {
  React.render(
    React.createElement(Lobby, null),
    document.getElementById('lobby')
  );
}, 1);
