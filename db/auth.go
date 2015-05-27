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
	ok      bool
	message string
}

// PostRegister handles a new registration
func (a *Auth) PostRegister(rw http.ResponseWriter, req *http.Request) {
	var user httpauth.UserData
	user.Username = req.PostFormValue("username")
	user.Email = req.PostFormValue("email")
	password := req.PostFormValue("password")
	if err := a.aaa.Register(rw, req, user, password); err == nil {
		a.PostLogin(rw, req)
	} else {
		fmt.Println(err)
		resp := authResponse{ok: false, message: err.Error()}
		js, err := json.Marshal(resp)
		if err != nil {
			rw.Write(js)
			return
		}
	}
}

// PostLogin handles logins to the site
func (a *Auth) PostLogin(rw http.ResponseWriter, req *http.Request) {
	username := req.PostFormValue("username")
	password := req.PostFormValue("password")
	if err := a.aaa.Login(rw, req, username, password, "/"); err != nil && err.Error() == "already authenticated" {
		resp := authResponse{ok: true, message: "succesfully logged in"}
		js, err := json.Marshal(resp)
		if err != nil {
			rw.Write(js)
			return
		}
	} else if err != nil {
		fmt.Println(err)
		resp := authResponse{ok: false, message: err.Error()}
		js, err := json.Marshal(resp)
		if err != nil {
			rw.Write(js)
			return
		}
	}
	resp := authResponse{ok: true, message: "succesfully logged in"}
	js, err := json.Marshal(resp)
	if err != nil {
		rw.Write(js)
		return
	}
}
