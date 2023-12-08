import React from 'react';
import Game from './Game.jsx'
import Login from './Login.jsx'
import LeaderBoard from './LeaderBoard.jsx'

export default class Lobby extends React.Component {
    constructor(props) {
        super(props)
        this.state = {
            name: null,
            games: [],

            showStaleLobby: false,
            nextReloadCountdownSecs: 0,
        }
    }

    joinNew() {
        this.props.activate(<Game name={this.state.name} />)
    }
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
                    <h3>{gameCount} Current Game{gameCount === 1 ? "" : "s"}</h3>
                    {this.state.showStaleLobby &&
                        <div className="stale-lobby-warning">Lobby is stale. Will auto-refresh in {this.state.nextReloadCountdownSecs} seconds. <button onClick={(e) => this.ResetBackoffThenReload()}>Refresh now</button></div>}
                    <ul className="games">
                        {games}
                    </ul>
                </div>
                <div><button onClick={(e) => this.joinNew(e)}>Join New Game</button></div>
                <LeaderBoard />
            </div>
        )
    }
    setName(name) {
        this.setState({ name })
    }
    ResetBackoffThenReload() {
        this.reloadBackoffStart = new Date().getTime();
        this.Reload()
    }
    Reload() {
        var comp = this;

        var secondsSinceBackoffStart = (new Date().getTime() - comp.reloadBackoffStart) / 1000
        // Reloads every second for the two minutes, sub 5 seconds for at least 5 minutes.
        // Hits maximum backoff within 20 minutes. Maximum backoff is 5 minutes.
        var backoffSeconds = Math.min(Math.floor(Math.pow(1.005, secondsSinceBackoffStart), 300));
        comp.nextReloadTime = new Date().getTime() + backoffSeconds * 1000;
        clearTimeout(comp.reloader);
        comp.reloader = setTimeout(() => comp.Reload(), backoffSeconds * 1000);

        var xhr = new XMLHttpRequest();
        xhr.onreadystatechange = function () {
            if (xhr.readyState === 4) {
                if (xhr.status === 200) {
                    var data = JSON.parse(xhr.responseText)
                    if (!data) {
                        data = []
                    }
                    comp.lastReloadSuccessTime = new Date().getTime()
                    comp.setState({ games: data, showStaleLobby: false });
                }
            }
        }
        xhr.open("GET", "/listGames", true);
        xhr.send();
    }

    UpdateStaleLobbyWarning() {
        var now = new Date().getTime();
        var countDown = Math.round((this.nextReloadTime - now) / 1000);
        this.setState({
            showStaleLobby: (new Date().getTime() - this.lastReloadSuccessTime) > 5 * 1000,
            nextReloadCountdownSecs: countDown,
        })
    }
    componentDidMount() {
        this.ResetBackoffThenReload();
        var comp = this;
        this.staleLobbyWarner = setInterval(() => comp.UpdateStaleLobbyWarning(), 1 * 1000);
    }
    componentWillUnmount() {
        clearTimeout(this.reloader);
        clearInterval(this.staleLobbyWarner);
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

