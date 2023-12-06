import React from 'react';

export default class Login extends React.Component {
  constructor(props) {
    super(props)
    this.state = {
      user : null,
      username: null,
      password: null
    };
  }

  render() {
    if (!this.state.user){
      return(
        <div className="login-div">
          <form className="login-form" onSubmit={(e) => this.submitLogin(e)} action="/login" method="post">
            <input type="text" name="username" ref="username" value={this.state.username} onChange={(e) => this.usernameChanged(e)}/>
            <input type="password" name="password" ref="password" value={this.state.password} onChange={(e) => this.passwordChanged(e)}/>
            <button onClick={(e) => this.submitLogin(e)}>Sign In</button>
            <button onClick={(e) => this.submitRegister(e)}>Register</button>
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
  }
  componentDidMount() {
    // check for cookie here
    var comp=this;
    var xhr = new XMLHttpRequest();
    xhr.onreadystatechange = function() {
      if (xhr.readyState===4 && xhr.status === 200){
          if (comp.props.setName){
            comp.props.setName(xhr.responseText);
          }
          comp.setState({user: xhr.responseText});

      }
    }
    xhr.open("GET", "/cu", true);
    xhr.send();

  }
  usernameChanged(e){
    this.setState({username: e.target.value});
  }
  passwordChanged(e){
    this.setState({password: e.target.value});
  }
  submitLogin(e){

  }
  submitRegister(e){
    var form = e.target.closest('form');
    form.action = '/register';
  }
}
