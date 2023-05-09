package db

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/arbrown/pao/settings"

	"github.com/apexskier/httpauth"
)

// Auth controls the authorization of users
type Auth struct {
	aaa httpauth.Authorizer
}

// NewAuth returns a new Auth for authenticating users
func NewAuth(s settings.PaoSettings) (a *Auth, e error) {
	backend, err := httpauth.NewSqlAuthBackend(s.DbConfig.Driver, s.DbConfig.ConnectionString)
	roles := make(map[string]httpauth.Role)
	roles["user"] = 30
	roles["admin"] = 99
	a = &Auth{}
	if err != nil {
		e = err
		return
	}
	a.aaa, e = httpauth.NewAuthorizer(backend, []byte(s.AuthConfig.EncryptionKey), "user", roles)
	return
}

type authResponse struct {
	Ok       bool
	Message  string
	Username string
}

// PostRegister handles a new registration
func (a *Auth) PostRegister(rw http.ResponseWriter, req *http.Request) {
	var user httpauth.UserData
	user.Username = req.PostFormValue("username")
	user.Email = "fake@email.com" // httpauth requires email...
	password := req.PostFormValue("password")
	fmt.Printf("trying to register new user: %+v\n", user)
	if err := a.aaa.Register(rw, req, user, password); err == nil {
		a.PostLogin(rw, req)
	} else {
		fmt.Println(err)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// PostLogin handles logins to the site
func (a *Auth) PostLogin(rw http.ResponseWriter, req *http.Request) {
	username := req.PostFormValue("username")
	password := req.PostFormValue("password")
	if err := a.aaa.Login(rw, req, username, password, "/"); err != nil && err.Error() == "httpauth: already authenticated" {
		http.Redirect(rw, req, "/", http.StatusFound)
	} else if err != nil {
		fmt.Println(err)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	resp := authResponse{Ok: true, Message: "succesfully logged in", Username: username}
	js, err := json.Marshal(resp)
	if err != nil {
		rw.Header().Set("Content-Type", "application/json")
		rw.Write(js)
		return
	}
}

// HandleLogout logs the current user out
func (a *Auth) HandleLogout(rw http.ResponseWriter, req *http.Request) {
	if err := a.aaa.Logout(rw, req); err != nil {
		fmt.Println(err)
		// this shouldn't happen
		return
	}
	http.Redirect(rw, req, "/", http.StatusSeeOther)
}

// GetUser returns the currently logged in user (if any)
func (a *Auth) GetUser(rw http.ResponseWriter, req *http.Request) *httpauth.UserData {
	fmt.Println("Trying to get current user")
	user, err := a.aaa.CurrentUser(rw, req)
	if err == nil && user.Username != "" {
		fmt.Printf("Retrieved user: %+v\n", user)
		return &user
	}
	fmt.Printf("Got %+v\n", user)
	fmt.Printf("Error was: %s\n", err)
	return nil
}

// Cu returns the current logged-in user's name (if any)
func (a *Auth) Cu(rw http.ResponseWriter, req *http.Request) {
	user, err := a.aaa.CurrentUser(rw, req)
	if err != nil {
		fmt.Printf("Error reading user: %+v\n", err)
		return
	}
	fmt.Printf("User [%s] asks 'who am i?' (/cu)\n", user.Username)
	if user.Username != "" {
		rw.Write([]byte(user.Username))
		return
	}
}
