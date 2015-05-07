var ChatMessage = React.createClass({
  render: function(){
    return(
      <div className="chat-message">
        <strong>{this.props.player}: </strong>
        <p>{this.props.text}</p>
      </div>
    );
  }
});

var Chat = React.createClass({
  submitChat: function(e){
    e.preventDefault();
    console.log("User sent chat message: " + this.state.chatMessage);
    var messages = this.state.messages;
    messages.push({player: "Drew", text:this.state.chatMessage});
    this.setState({chatMessage: null, messages});
  },
  changeMessage: function(){
    var message = this.refs.chatInput.getDOMNode().value;
    this.setState({chatMessage: message});
  },
  render: function() {
    var messages = this.state.messages.map(function(message, i){
      return (<ChatMessage key={i} player={message.player} text={message.text} />);
    });
    return(
      <div className="chat">
        <div className="chat-messages">
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
      messages :[{text: "Foo"}, {text: "Bar"}]
    };
  },
})
