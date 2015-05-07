var Square = React.createClass({
  handleClick: function(e){
    this.setState({selected: !this.state.selected});
  },
  render: function() {
    var classes = [];
    classes.push('banqi-square');
    if (this.state.selected){
      classes.push('selected')
    }
    classString = classes.reduce(function(p, c) { return p + " " + c});
    return <td className={classString} onClick={this.handleClick}></td>
  },
  getInitialState: function() {
    return {
      selected : false,
    };
  },
});

var Board = React.createClass({
  render: function(){
    var rows=[];
    for (var i=0; i<4; i++){
      rows.push([]);
      for (var j=0; j< 8; j++){
        rows[i].push('?');
      }
    }
    var rowElements = rows.map(function(row){
      var files = row.map(function(file){
        return <Square value={file}/>
      });
      return (<tr>{files}</tr>);
    })
    return (
      <table className="banqi-board">
        {rowElements}
      </table>
    )
  }
});

React.render(
  React.createElement(Board, null),
  document.getElementById('board')
);
