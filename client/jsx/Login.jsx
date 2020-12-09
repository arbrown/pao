var Login = React.createClass({
  render: function() {
    if (!this.state.user){
      return(
        <div className="login-div">
          <form className="login-form" onSubmit={this.submitLogin} action="/login" method="post">
            <input type="text" name="username" ref="username" value={this.state.username} onChange={this.usernameChanged}/>
            <input type="password" name="password" ref="password" value={this.state.password} onChange={this.passwordChanged}/>
            <button onClick={this.submitLogin}>Sign In</button>
            <button onClick={this.submitRegister}>Register</button>
          </form>
        </div>
      )
    } else {
      return (
        <div className="login-div">
          {this.state.user}<br />
          <a href="/logout">Sign Out</a>
        </div>
      )
    }
  },
  getInitialState: function() {
    return {
      user : null,
      username: null,
      password: null
    };
  },
  componentDidMount: function() {
    // check for cookie here
    var comp=this;
    var xhr = new XMLHttpRequest();
    xhr.onreadystatechange = function() {
      if (xhr.readyState==4 && xhr.status == 200){
          if (comp.props.setName){
            comp.props.setName(xhr.responseText);
          }
          comp.setState({user: xhr.responseText});

      }
    }
    xhr.open("GET", "/cu", true);
    xhr.send();

  },
  usernameChanged: function(){
    this.setState({username: this.refs.username.getDOMNode().value});
  },
  passwordChanged: function(){
    this.setState({password: this.refs.password.getDOMNode().value});
  },
  submitLogin: function(e){

  },
  submitRegister: function(e){
    var form = e.target.closest('form');
    form.action = '/register';
  }
})
