import React from 'react';
import { StrengthCompareD, NotationToColor, NotationToCss } from './Utils.js';

export default class DeadPieces extends React.Component {
  constructor(props) {
    super(props)
    this.state = {
      sort : "strength",
      lastDead: ""
    }
  }
  
  render(){
    var redDead = [];
    var blackDead = [];
    var chances = this.computeRemainingChances(this.props.dead, this.props.board);
    var dead = this.props.dead;
    for (var i= 0; i<dead.length; i++){
      var color = NotationToColor[dead[i]]
      if (color === 'red'){
        redDead.push(dead[i])
      } else if (color === 'black'){
        blackDead.push(dead[i])
      }
    }
    if (this.state.sort === 'strength') {
      redDead.sort(StrengthCompareD);
      blackDead.sort(StrengthCompareD);
    }
    var livePieces = this.collectKnownPieces(this.props.board)
    var redPieces = this.deadPieces("kggeecchhpppppqq", redDead, livePieces, chances);
    var blackPieces = this.deadPieces("KGGEECCHHPPPPPQQ", blackDead, livePieces, chances);
    return (
      <div className="dead-pieces">
        <div className="red-dead">{redPieces}</div>
        <div className="black-dead">{blackPieces}</div>
      </div>
    );
  }
  deadPiece(piece, state, lastMove, chances) {
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
    var title = state + " " + type;
    if (state === 'unborn') {
        title += " " + chances[type] + "%";
    }
    return <div className={classString} title={title} />
  }
  deadPieces(all, dead, live, chances) {
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
        if (piece === lastDead) {
          lastMove = true;
          lastDead = ''
        }
      } else if (live.indexOf(piece) >= 0) {
        state = 'live';
        live = live.replace(piece, '')
      }
      return self.deadPiece(piece, state, lastMove, chances);
    })
  }
  collectKnownPieces(board) {
    var knownPieces = [];
    for (var rank=0; rank < 4; rank++) {
      for (var file=0; file < 8; file++) {
        var piece = board[rank][file]
        if (piece !== '?' && piece !== '.') {
          knownPieces.push(piece);
        }
      }
    }
    return knownPieces;
  }
    computeRemainingChances(dead, board) {
      // eslint-disable-next-line
      let remaining = 'kggeecchhpppppqq' + 'KGGEECCHHPPPPPQQ';
      let chances = {};
      dead = dead.concat(this.collectKnownPieces(board));
      let pieceCounts = {}
      for (let i in dead) {
          let piece = dead[i];
          let index = remaining.indexOf(piece);
          remaining = remaining.slice(0, index) + remaining.slice(index + 1);
      }

      for (let i = 0; i< remaining.length; i++) {
          let piece = remaining[i];
          let type = NotationToCss[piece];
          if (pieceCounts[type]) {
              pieceCounts[type]++
          } else {
              pieceCounts[type] = 1;
          }
      }
      for (let type in pieceCounts) {
          let percent = ((pieceCounts[type] / remaining.length) * 100).toFixed(2);
          chances[type] =  percent;
      }
      return chances;
  }
}
