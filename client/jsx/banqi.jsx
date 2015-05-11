var banqi = React.createClass({
  render: function(){
    return(<h1>Hello, Banqi!</h1>);
  }
})

React.render(
  React.createElement(banqi, null),
  document.getElementById('content')
);
