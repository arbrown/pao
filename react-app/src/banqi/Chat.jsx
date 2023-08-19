import React from 'react';

class ChatMessage extends React.Component {
  render(){
    var css = "chat-message " + (this.props.auth ? "auth" : "");
    return(
      <div className={css}>
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
      return (<ChatMessage key={i} player={message.player} text={message.text} color={message.color} auth={message.auth}/>);
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
