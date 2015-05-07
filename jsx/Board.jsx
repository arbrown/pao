var Square = React.createClass({
  handleClick: function(e){
    var f = this.props.handleClick;
    var rank = this.props.rank;
    var file = this.props.file;
    !f||f({rank,file});
  },
  render: function() {
    var classes = [];
    classes.push('banqi-square');
    if (this.props.selected){
      classes.push('selected')
    }
    var type = NotationToCss[this.props.piece];
    if (type){
      classes.push(type);
    }
    classString = classes.reduce(function(p, c) { return p + " " + c});
    return(
    <td className={classString} onClick={this.handleClick}>
      {this.props.children}
    </td>);
  },
});

var Board = React.createClass({
  flipPiece: function(square){
    console.log("Tried to flip piece at: " + square.rank+","+square.file);
  },
  move: function(attacker, target){
    console.log("Tried to attack/move from: " + attacker.rank+"," +attacker.file
    + " to: " + target.rank+","+target.file);
  },
  handleClick: function(clicked){
    var s = this.state.selected;
    var sp = s ? this.state.board[s.rank][s.file] : null;
    var cp = this.state.board[clicked.rank][clicked.file];
    if (s && s.rank == clicked.rank && s.file == clicked.file){
      if (this.state.myTurn && sp == '?'){
        this.flipPiece(clicked);
      }
      this.setState({selected: null});
    }
    else{ if (s && this.IOwn(s) && this.state.myTurn){
      this.move(s, clicked);
      this.setState({selected: null});
    } else{
        this.setState({selected: clicked});
      }
    }
  },
  render: function(){
    var rows=[];
    var comp = this;
    for (var i=0; i<4; i++){
      rows.push([]);
      for (var j=0; j< 8; j++){
        rows[i].push({rank:i, file:j});
      }
    }

    var current = this.state.selected;

    var rowElements = this.state.board.map(function(row, rank){
      var squares = row.map(function(square, file){
        return (
          <Square
            handleClick={comp.handleClick}
            piece={square}
            selected={current && current.rank == rank && current.file == file}
            rank={rank}
            file={file}
            >
          </Square>);
        }
      );
      return (<tr>{squares}</tr>);
    });
    return (
      <table className="banqi-board">
        {rowElements}
      </table>
    )
  },
  getInitialState: function() {
    return {
      board : [
      ['?','?','?','?','?','?','?','?'],
      ['?','k','K','?','?','?','?','?'],
      ['?','?','?','?','?','?','?','?'],
      ['?','?','?','?','?','?','?','?']],
      selected : null,
      myTurn : true,
      myColor : 'red'
    };
  },
  IOwn : function(square){
    switch (this.state.board[square.rank][square.file]) {
      case 'K':
      case 'G':
      case 'E':
      case 'C':
      case 'H':
      case 'P':
      case 'Q':
        return this.state.myColor == 'black';
        break;
      case 'k':
      case 'g':
      case 'e':
      case 'c':
      case 'h':
      case 'p':
      case 'q':
        return this.state.myColor == 'red';
        break;
      default:
        return undefined;
    }
  }
});

NotationToCss = {
  'K' : 'black-king',
  'G' : 'black-guard',
  'E' : 'black-elephant',
  'C' : 'black-cart',
  'H' : 'black-horse',
  'P' : 'black-pawn',
  'Q' : 'black-cannon',
  'k' : 'red-king',
  'g' : 'red-gaurd',
  'e' : 'red-elephant',
  'c' : 'red-cart',
  'h' : 'red-horse',
  'p' : 'red-pawn',
  'q' : 'red-cannon',
  '?' : 'unflipped-piece'
}


React.render(
  React.createElement(Board, null),
  document.getElementById('board')
);
