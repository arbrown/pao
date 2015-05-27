package db

import (
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

func (a *Auth) postRegister(rw http.ResponseWriter, req *http.Request) {
	var user httpauth.UserData
	user.Username = req.PostFormValue("username")
	user.Email = req.PostFormValue("email")
	password := req.PostFormValue("password")
	if err := a.aaa.Register(rw, req, user, password); err == nil {
		a.postLogin(rw, req)
	} else {
		http.Redirect(rw, req, "/login", http.StatusSeeOther)
	}
}

func (a *Auth) postLogin(rw http.ResponseWriter, req *http.Request) {
	username := req.PostFormValue("username")
	password := req.PostFormValue("password")
	if err := a.aaa.Login(rw, req, username, password, "/"); err != nil && err.Error() == "already authenticated" {
		http.Redirect(rw, req, "/", http.StatusSeeOther)
	} else if err != nil {
		fmt.Println(err)
		http.Redirect(rw, req, "/login", http.StatusSeeOther)
	}
}
