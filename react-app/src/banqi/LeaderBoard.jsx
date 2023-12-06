import React from 'react';

export default class LeaderBoard extends React.Component {
  constructor(props) {
    super(props)
    this.state = {
      leaders: [],
    }
  }

  render() {
    var leaders = this.state.leaders.map(function (l) {
      return (
        <tr>
          <td className="leader-name" onClick={l.Name === "walrus" ? () => document.querySelector('html').classList.add('walrus') : undefined}>{l.Name}</td>
          <td className="leader-wins">{l.Wins}</td>
        </tr>
      );
    });
    return (
      <div className="leader-board">
        <h3>Pao Leader Board</h3>
        <table>
          <thead>
            <tr><th>Player</th><th>Wins</th></tr>
          </thead>
          <tbody>
            {leaders}
          </tbody>
        </table>
      </div>
    );
  }
  componentDidMount() {
    var comp = this;
    var xhr = new XMLHttpRequest();
    xhr.onreadystatechange = function () {
      if (xhr.readyState === 4 && xhr.status === 200) {
        var data = JSON.parse(xhr.responseText)
        comp.setState({ leaders: data });
      }
    }
    xhr.open("GET", "/leaderBoard", true);
    xhr.send();
  }
}
