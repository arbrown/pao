import React from 'react';
import { timeAgo } from './timeago.js';

class ChatMessage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      now: Date.now()
    };
  }

  componentDidMount() {
    this.interval = setInterval(() => this.setState({ now: Date.now() }), 10000);
  }

  componentWillUnmount() {
    clearInterval(this.interval);
  }

  render(){
    var css = "chat-message " + (this.props.auth ? "auth" : "");
    const relativeTime = timeAgo(this.props.timestamp, this.state.now);

    return(
      <div className={css} title={relativeTime}>
        <strong style={{color: this.props.color}}>{this.props.player}</strong>:
        &nbsp;
        <p>{this.props.text}</p>
      </div>
    );
  }
}

export default class Chat extends React.Component {
  constructor(props) {
    super(props)
    this.state = {
    }
  }
  
  submitChat(e){
    e.preventDefault();
    var message = this.state.chatMessage;
    console.log("User sent chat message: " + this.state.chatMessage);
    !this.props.submitChat || this.props.submitChat(message);
    this.setState({chatMessage: ""})
  }
  changeMessage(e){
    var message = e.target.value
    this.setState({chatMessage: message});
  }
  render() {
    var messages = this.props.chats.map(function(message, i){
      if (!message.timestamp) {
        return (<ChatMessage key={i} player={message.player} text={message.text} color={message.color} auth={message.auth} timestamp={new Date(0)}/>);
      }
      return (<ChatMessage key={i} player={message.player} text={message.text} color={message.color} auth={message.auth} timestamp={message.timestamp}/>);
    });
    return(
      <div className="chat">
        <div className="chat-messages" ref={(el) => {this.chatMessages = el}}>
          {messages}
        </div>
        <form onSubmit={(e) => this.submitChat(e)}>
          <input
            type="text"
            value={this.state.chatMessage}
            onChange={(e) => this.changeMessage(e)}
            placeholder="Type a chat message"/>
        </form>
      </div>
    );
  }
  componentDidUpdate(prevProps, prevState) {
    var div = this.chatMessages;
    div.scrollTop = div.scrollHeight;
  }
}
