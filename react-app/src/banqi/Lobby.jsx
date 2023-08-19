import React from 'react';
import Game from './Game.jsx'
import Login from './Login.jsx'
import LeaderBoard from './LeaderBoard.jsx'

export default class Lobby extends React.Component {
    constructor(props) {
        super(props)
        this.state = {
            name: null,
            games: []
        }
    }

    joinNew() {
        //let g = React.createElement(Game, { name: this.state.name });
        this.props.activate(<Game name={this.state.name} />)
    }
    // playAi() {
    // this.props.activate(<Game name={this.state.name} ai="Flippy"/>)
    // }
    join(id) {
        this.props.activate(<Game name={this.state.name} id={id} />)
    }
    nameChanged() {
        var name = this.refs.name.getDOMNode().value;
        this.setState({ name });
    }
    render() {
        var comp = this;
        var games = this.state.games.map(function (g) {
            return <LobbyGame id={g.ID} players={g.Players}
                onClick={function () { comp.join(g.ID) }} />
        });
        var gameCount = games ? games.length : 0;
        return (
            <div className="lobby">
                <span id="forkongithub">
                    <a href="https://github.com/arbrown/pao">
                        Fork me on GitHub
                    </a>
                </span>
                <Login setName={(n) => this.setName(n)} />
                <h2>Pao Lobby</h2>
                <input type="text" ref="name" value={this.state.name} onChange={(e) => this.nameChanged(e)} placeholder="Your Name" />
                <div className="lobby-current-games">
                    <h3>{gameCount} Current Game{gameCount == 1 ? "" : "s"}</h3>
                    <ul className="games">
                        {games}
                    </ul>
                </div>
                <div><button onClick={(e) => this.joinNew(e)}>Join New Game</button></div>
                <div><button onClick={(e) => this.playAi(e)}>Play Flippy</button></div>
                <LeaderBoard />
                <h4>Now with more react!</h4>
            </div>
        )
    }
    setName(name) {
        this.setState({ name })
    }
    Reload() {
        var comp = this;
        var xhr = new XMLHttpRequest();
        xhr.onreadystatechange = function () {
            if (xhr.readyState == 4 && xhr.status == 200) {
                var data = JSON.parse(xhr.responseText)
                if (data) {
                    comp.setState({ games: data });
                } else {
                    comp.setState({ games: [] });
                }
            }
        }
        xhr.open("GET", "/listGames", true);
        xhr.send();
    }
    componentDidMount() {
        this.Reload();
        var comp = this;
        this.reloader = setInterval(() => comp.Reload(), 10 * 1000);
    }
    componentWillUnmount() {
        clearInterval(this.reloader)
    }

}

class LobbyGame extends React.Component {
    render() {
        var playerList = this.props.players.map(function (player) {
            return (<li className="player">{player}</li>);
        });
        return (
            <li className="lobby-game" onClick={(e) => this.props.onClick(e)}>
                <div className="banqi-square red-cannon" />
                <p>Current Players:</p>
                <ul>{playerList}</ul>
            </li>);
    }
}

