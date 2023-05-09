import './Banqi.css';

import Lobby from './banqi/Lobby.jsx'
import React from 'react';

type BanqiState = {
  active: React.ReactNode
}

class Banqi extends React.Component<any, BanqiState> {

  state: BanqiState = {
    active: <Lobby activate={(c: React.ReactNode)=>this.activate(c)}/>
  }

  render(): React.ReactNode {
    return this.state.active
  }

  activate(node: React.ReactNode) {
    this.setState({active: node})
  }
}

export default Banqi;
