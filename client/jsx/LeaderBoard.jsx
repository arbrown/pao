var LeaderBoard = React.createClass({
  render: function() {
    var leaders = this.state.leaders.map(function(l){
      return(
        <tr>
          <td className="leader-name">{l.Name}</td>
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
  },
  componentDidMount: function() {
    var comp = this;
    var xhr = new XMLHttpRequest();
    xhr.onreadystatechange = function() {
      if (xhr.readyState==4 && xhr.status == 200){
        var data = JSON.parse(xhr.responseText)
        comp.setState({leaders: data});
      }
    }
    xhr.open("GET", "/leaderBoard", true);
    xhr.send();
  },
  getInitialState: function() {
    return {
      leaders : []
    };
  },
});
