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
    var redPieces = redDead.map(PieceMap);
    var blackPieces = blackDead.map(PieceMap);
    return (
      <div className="dead-pieces">
        <div className="red-dead">{redPieces}</div>
        <div className="black-dead">{blackPieces}</div>
      </div>
    );
  },
  getInitialState: function() {
    return {
      sort : "strength"
    };
  },
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

PieceMap = function(piece){
  var classes = []
  classes.push('banqi-square');
  classes.push('dead');
  classes.push(NotationToCss[piece]);
  var classString = classes.reduce(function(p, c) { return p + " " + c});
  return <div className={classString} />
};
