var DeadPieces = React.createClass({
  render : function(){
    var redDead = [];
    var blackDead = [];
    var dead = this.props.dead;
    for (var i= 0; i<dead.length; i++){
      var color = NotationToColor[dead[i]]
      if (color == 'red'){
        redDead.push(dead[i])
      } else if (color == 'black'){
        blackDead.push(dead[i])
      }
    }
    if (this.state.sort == 'strength') {
      redDead.sort(StrengthCompareD);
      blackDead.sort(StrengthCompareD);
    }
    var livePieces = this.collectKnownPieces(this.props.board)
    var redPieces = this.deadPieces("kggeecchhpppppqq", redDead, livePieces);
    var blackPieces = this.deadPieces("KGGEECCHHPPPPPQQ", blackDead, livePieces);
    return (
      <div className="dead-pieces">
        <div className="red-dead">{redPieces}</div>
        <div className="black-dead">{blackPieces}</div>
      </div>
    );
  },
  getInitialState: function() {
    return {
      sort : "strength",
      lastDead: ""
    };
  },
  deadPiece: function(piece, state, lastMove) {
    // `state` is 'dead', 'live', or 'unborn'
    var classes = []
    classes.push('banqi-square');
    classes.push(state);
    if (lastMove) {
        classes.push('last-move');
    }
    var type = NotationToCss[piece];
    classes.push(NotationToCss[piece]);
    var classString = classes.reduce(function(p, c) { return p + " " + c});
    return <div className={classString} title={state + " " + type} />
  },
  deadPieces: function(all, dead, live) {
    var lastDead = this.props.lastDead;
    var self = this;
    dead = dead.reduce(function(p, c) { return p + " " + c}, '');
    live = live.reduce(function(p, c) { return p + " " + c}, '');
    return all.split('').map(function(piece) {
      var state = 'unborn'
      var lastMove = false;
      if (dead.indexOf(piece) >= 0) {
        state = 'dead';
        dead = dead.replace(piece, '');
        if (piece == lastDead) {
          lastMove = true;
          lastDead = ''
        }
      } else if (live.indexOf(piece) >= 0) {
        state = 'live';
        live = live.replace(piece, '')
      }
      return self.deadPiece(piece, state, lastMove);
    })
  },
  collectKnownPieces: function(board) {
    var knownPieces = [];
    for (var rank=0; rank < 4; rank++) {
      for (var file=0; file < 8; file++) {
        var piece = board[rank][file]
        if (piece != '?' && piece != '.') {
          knownPieces.push(piece);
        }
      }
    }
    return knownPieces;
  }
});

PieceStrength = ["Q","P","H","C","E","G","K"]
StrengthCompareD = function(a,b){
  return StrengthCompare(a,b,true)
};
StrengthCompare = function(a, b, desc){
  a = a.toUpperCase();
  b = b.toUpperCase();
  aStrength = PieceStrength.indexOf(a);
  bStrength = PieceStrength.indexOf(b);
  if (!(aStrength > -1 && bStrength > -1)){
    return 0;
  }
  if (desc){
    return bStrength - aStrength;
  }
  return aStrength - bStrength;
};
