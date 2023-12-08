import React from 'react';
import { NotationToColor, NotationToCss } from './Utils.js'

class Square extends React.Component {
  handleClick(e) {
    var f = this.props.handleClick;
    var rank = this.props.rank;
    var file = this.props.file;
    !f || f({ rank, file });
  }
  render() {
    var classes = [];
    var files = 'ABCDEFGH';
    var ranks = '1234';
    var coord = files.charAt(this.props.file) + ranks.charAt(this.props.rank)
    classes.push('banqi-square');
    classes.push('banqi-square-' + coord)
    if (this.props.lastMove && this.props.lastMove.indexOf(coord) !== -1) {
      classes.push('last-move')
    }
    if (this.props.selected) {
      classes.push('selected')
    }
    var type = NotationToCss[this.props.piece];
    if (type) {
      classes.push(type);
    }
    var classString = classes.reduce(function (p, c) { return p + " " + c });
    return (
      <td className={classString} onClick={(e) => this.handleClick(e)} title={coord + " " + type}>
        {this.props.children}
      </td>);
  }
}

class RowHeader extends React.Component {
  render() {
    var rank = this.props.rank;
    return (
      <th scope="row">
        {"1234"[rank]}
      </th>);
  }
}

class ColumnHeader extends React.Component {
  render() {
    var file = this.props.file;
    return (
      <th scope="col">
        {"ABCDEFGH"[file]}
      </th>);
  }
}

export default class Board extends React.Component {
  constructor(props) {
    super(props)
    this.state = {
      selected: null,
      lastMove: null
    }
  }

  flipPiece(square) {
    var files = 'ABCDEFGH';
    var ranks = '1234';
    var move = '?' + files.charAt(square.file) + ranks.charAt(square.rank);
    console.log("Tried to flip piece at: " + square.rank + "," + square.file);
    !this.props.sendMove || this.props.sendMove(move)
  }
  move(attacker, target) {
    var files = 'ABCDEFGH';
    var ranks = '1234';
    console.log("Tried to attack/move from: " + attacker.rank + "," + attacker.file
      + " to: " + target.rank + "," + target.file);
    var move = files.charAt(attacker.file) + ranks.charAt(attacker.rank) + '>' +
      files.charAt(target.file) + ranks.charAt(target.rank);
    !this.props.sendMove || this.props.sendMove(move);
  }
  handleClick(clicked) {
    var s = this.state.selected;
    var sp = s ? this.props.board[s.rank][s.file] : null;
    if (s && s.rank === clicked.rank && s.file === clicked.file) {
      if (sp === '?' && (this.props.myTurn || this.props.lastMove !== null || this.props.firstMove)) {
        this.flipPiece(clicked);
      }
      this.setState({ selected: null });
    }
    else {
      if (s && (this.props.myTurn || this.props.lastMove != null)) {
        this.move(s, clicked);
        this.setState({ selected: null });
      } else {
        this.setState({ selected: clicked });
      }
    }
  }
  render() {
    var current = this.state.selected;
    var lastMove = this.props.lastMove
    var comp = this;
    var colHeaders = []
    colHeaders.push(<th />); // the row-label column needs no column header
    for (var i = 0; i < 8; i++) {
      colHeaders.push(<ColumnHeader file={i} />)
    }
    var rowElements = this.props.board.map(function (row, rank) {
      var rowHeader = <RowHeader rank={rank} />
      var squares = row.map(function (square, file) {
        return (
          <Square
            handleClick={(e) => comp.handleClick(e)}
            piece={square}
            selected={current && current.rank === rank && current.file === file}
            lastMove={lastMove}
            rank={rank}
            file={file}
            key={(rank + file) * (rank + file + 1) / 2 + file}
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
  }
  IOwn(square) {
    var piece = this.props.board[square.rank][square.file];
    return NotationToColor[piece] === this.props.myColor;
  }
}
