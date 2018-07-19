const AUTH0_CLIENT_ID = "IrJYUlgZUqeguEVh8GWxu5H3fUCz0388",
  AUTH0_DOMAIN = "virtyaluk.auth0.com",
  AUTH0_CALLBACK_URL = location.href,
  AUTH0_API_AUDIENCE = "jokeish-api";

class App extends React.Component {
  render() {
    if (this.loggedIn) {
      return <LoggedIn />;
    } else {
      return <Home />;
    }
  }

  setState() {
    let idToken = localStorage.getItem("id_token");

    this.loggedIn = !!idToken;
  }

  setup() {
    $.ajaxSetup({
      beforeSend: (r) => {
        if (localStorage.getItem("access_token")) {
          r.setRequestHeader("Authorization", "Bearer " + localStorage.getItem("access_token"));
        }
      }
    });
  }

  parseHash() {
    this.auth0 = new auth0.WebAuth({
      domain: AUTH0_DOMAIN,
      clientID: AUTH0_CLIENT_ID
    });

    this.auth0.parseHash(window.location.hash, (err, authResult) => {
      if (err) {
        return console.log(err);
      }

      if (authResult && authResult.accessToken && authResult.idToken) {
        localStorage.setItem("access_token", authResult.accessToken);
        localStorage.setItem("id_token", authResult.idToken);
        localStorage.setItem("profile", JSON.stringify(authResult.idTokenPayload));

        window.location = window.location.href.substr(0, window.location.href.indexOf("#"));
      }
    });
  }

  componentWillMount() {
    this.setup();
    this.parseHash();
    this.setState();
  }
}

class Home extends React.Component {
  constructor(props) {
    super(props);
    this.authenticate = this.authenticate.bind(this);
  }

  authenticate() {
    this.WebAuth = new auth0.WebAuth({
      domain: AUTH0_DOMAIN,
      clientID: AUTH0_CLIENT_ID,
      scope: "openid profile",
      audience: AUTH0_API_AUDIENCE,
      responseType: "token id_token",
      redirectUri: AUTH0_CALLBACK_URL
    });
    this.WebAuth.authorize();
  }

  render() {
    return (
      <div className="container">
        <div className="col-xs-8 col-xs-offset-2 jumbotron text-center">
          <h1>Jokes</h1>
          <p>A load of Dark Humor jokes :D</p>
          <p>Sign in to get access </p>
          <a onClick={this.authenticate} className="btn btn-primary btn-lg btn-login btn-block">Sign In</a>
        </div>
      </div>
    );
  }
}

class LoggedIn extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      jokes: []
    };

    this.serverRequest = this.serverRequest.bind(this);
    this.logout = this.logout.bind(this);
  }

  render() {
    return (
      <div className="container">
        <div className="col-lg-12">
          <br />
          <span className="pull-right"><a onClick={this.logout}>Log out</a></span>
          <h2>Jokes</h2>
          <p>Let's feed you with some funny Jokes!</p>
          <div className="row">
            {this.state.jokes.map((joke, i) => <Joke key={i} joke={joke} />)}
          </div>
        </div>
      </div>
    );
  }

  componentDidMount() {
    this.serverRequest();
  }

  logout() {
    localStorage.removeItem("id_token");
    localStorage.removeItem("access_token");
    localStorage.removeItem("profile");
    location.reload();
  }

  serverRequest() {
    $.get("http://localhost:3000/api/jokes", res => {
      this.setState({
        jokes: res
      });
    });
  }
}

class Joke extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      liked: ""
    };

    this.like = this.like.bind(this);
    this.serverRequest = this.serverRequest.bind(this);
  }

  like() {
    let joke = this.props.joke;

    this.serverRequest(joke);
  }

  serverRequest() {
    $.post("http://localhost:3000/api/jokes/like/" + joke.id, { like: 1 }, res => {
      console.log("jokes/like response", res);

      this.setState({ liked: "Liked!", jokes: res });
      this.props.jokes = res;
    });
  }

  render() {
    return (
      <div className="col-xs-4">
        <div className="panel panel-default">
          <div className="panel-heading">#{this.props.joke.id} <span className="pull-right">{this.state.liked}</span></div>
          <div className="panel-body">
            {this.props.joke.joke}
          </div>
          <div className="panel-footer">
            {this.props.joke.likes} Likes&nbsp;
            <a onClick={this.like} className="btn btn-default"><span className="glyphicon glyphicon-thumbs-up"></span></a>
          </div>
        </div>
      </div>
    );
  }
}

ReactDOM.render(<App />, document.getElementById("app"));
