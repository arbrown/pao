var Square = React.createClass({
  handleClick: function(e){
    var f = this.props.handleClick;
    var rank = this.props.rank;
    var file = this.props.file;
    !f||f({rank,file});
  },
  render: function() {
    var classes = [];
    var files = 'ABCDEFGH';
    var ranks = '1234';
    var coord = files.charAt(this.props.file) + ranks.charAt(this.props.rank)
    classes.push('banqi-square');
    classes.push('banqi-square-' + coord)
    if (this.props.lastMove && this.props.lastMove.indexOf(coord) != -1) {
        classes.push('last-move')
    }
    if (this.props.selected){
      classes.push('selected')
    }
    var type = NotationToCss[this.props.piece];
    if (type){
      classes.push(type);
    }
    classString = classes.reduce(function(p, c) { return p + " " + c});
    return(
    <td className={classString} onClick={this.handleClick} title={coord + " " + type}>
      {this.props.children}
    </td>);
  },
});

var RowHeader = React.createClass({
  render: function() {
    var classString = "header row-header";
    var rank = this.props.rank;
    return(
    <td className={classString}>
      {"1234"[rank]}
    </td>);
  },
});

var ColumnHeader = React.createClass({
  render: function() {
    var classString = "header column-header";
    var file = this.props.file;
    return(
    <td className={classString}>
      {"ABCDEFGH"[file]}
    </td>);
  },
});

var Board = React.createClass({
  flipPiece: function(square){
    var files = 'ABCDEFGH';
    var ranks = '1234';
    var move = '?'+files.charAt(square.file)+ ranks.charAt(square.rank);
    console.log("Tried to flip piece at: " + square.rank+","+square.file);
    !this.props.sendMove || this.props.sendMove(move)
  },
  move: function(attacker, target){
    var files = 'ABCDEFGH';
    var ranks = '1234';
    console.log("Tried to attack/move from: " + attacker.rank+"," +attacker.file
    + " to: " + target.rank+","+target.file);
    var move =  files.charAt(attacker.file) + ranks.charAt(attacker.rank) + '>' +
                files.charAt(target.file) + ranks.charAt(target.rank);
    !this.props.sendMove || this.props.sendMove(move);
  },
  handleClick: function(clicked){
    var s = this.state.selected;
    var sp = s ? this.props.board[s.rank][s.file] : null;
    var cp = this.props.board[clicked.rank][clicked.file];
    if (s && s.rank == clicked.rank && s.file == clicked.file){
      if (this.props.myTurn && sp == '?'){
        this.flipPiece(clicked);
      }
      this.setState({selected: null});
    }
    else{ if (s && this.IOwn(s) && this.props.myTurn){
      this.move(s, clicked);
      this.setState({selected: null});
    } else{
        this.setState({selected: clicked});
      }
    }
  },
  render: function(){
    var current = this.state.selected;
    var lastMove = this.props.lastMove
    var comp = this;
    var colHeaders = []
    colHeaders.push(<td />); // the row-label column needs no column header
    for (var i=0; i < 8; i++) {
      colHeaders.push(<ColumnHeader file={i} />)
    }
    var rowElements = this.props.board.map(function(row, rank){
      var rowHeader = <RowHeader rank={rank} />
      var squares = row.map(function(square, file){
        return (
          <Square
            handleClick={comp.handleClick}
            piece={square}
            selected={current && current.rank == rank && current.file == file}
            lastMove = {lastMove}
            rank={rank}
            file={file}
            key={(rank+file)*(rank+file+1)/2 + file}
            >
          </Square>);
        }
      );
      return (<tr>{rowHeader}{squares}</tr>);
    });
    return (
      <table className="banqi-board">
        {colHeaders}
        {rowElements}
      </table>
    )
  },
  getInitialState: function() {
    return {
      selected : null,
      lastMove: null
    };
  },
  IOwn : function(square){
    var piece = this.props.board[square.rank][square.file];
    return NotationToColor[piece] == this.props.myColor;
  }
});

NotationToColor = {
  'K' : 'black',
  'G' : 'black',
  'E' : 'black',
  'C' : 'black',
  'H' : 'black',
  'P' : 'black',
  'Q' : 'black',
  'k' : 'red',
  'g' : 'red',
  'e' : 'red',
  'c' : 'red',
  'h' : 'red',
  'p' : 'red',
  'q' : 'red',
};

NotationToCss = {
  'K' : 'black-king',
  'G' : 'black-guard',
  'E' : 'black-elephant',
  'C' : 'black-cart',
  'H' : 'black-horse',
  'P' : 'black-pawn',
  'Q' : 'black-cannon',
  'k' : 'red-king',
  'g' : 'red-guard',
  'e' : 'red-elephant',
  'c' : 'red-cart',
  'h' : 'red-horse',
  'p' : 'red-pawn',
  'q' : 'red-cannon',
  '?' : 'unflipped-piece'
};
