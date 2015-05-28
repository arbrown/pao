var ChatMessage = React.createClass({
  render: function(){
    var css = "chat-message " + (this.props.auth ? "auth" : "");
    return(
      <div className={css}>
        <strong style={{color: this.props.color}}>{this.props.player}</strong>:
        &nbsp;
        <p>{this.props.text}</p>
      </div>
    );
  }
});

var Chat = React.createClass({
  submitChat: function(e){
    e.preventDefault();
    var message = this.state.chatMessage;
    console.log("User sent chat message: " + this.state.chatMessage);
    !this.props.submitChat || this.props.submitChat(message);
    this.setState({chatMessage: null})
  },
  changeMessage: function(){
    var message = this.refs.chatInput.getDOMNode().value;
    this.setState({chatMessage: message});
  },
  render: function() {
    var messages = this.props.chats.map(function(message, i){
      return (<ChatMessage key={i} player={message.player} text={message.text} color={message.color} auth={message.auth}/>);
    });
    return(
      <div className="chat">
        <div className="chat-messages" ref="chatMessages">
          {messages}
        </div>
        <form onSubmit={this.submitChat}>
          <input
            type="text"
            ref="chatInput"
            value={this.state.chatMessage}
            onChange={this.changeMessage}
            placeholder="Type a chat message"/>
        </form>
      </div>
    );
  },
  getInitialState: function() {
    return {
    };
  },
  componentDidUpdate: function(prevProps, prevState) {
    var div = this.refs.chatMessages.getDOMNode();
    div.scrollTop = div.scrollHeight;
  },
})
