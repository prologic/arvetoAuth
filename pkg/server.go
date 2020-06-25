// Copyright (c) 2020, Arveto Ink. All rights reserved.
// Use of this source code is governed by a BSD
// license that can be found in the LICENSE file.

package auth

import (
	"./public"
	"crypto/rsa"
	"github.com/HuguesGuilleus/go-db.v1"
	"github.com/HuguesGuilleus/go-parsersa"
	"github.com/HuguesGuilleus/static.v1"
	"log"
	"net/http"
	"net/smtp"
	"path/filepath"
	"text/template"
)

// The option to create a server
type Option struct {
	URL string // The URL of the server
	DB  string // path to the DB
	Key string // private key file
	// Mail options, need to be processed
	MailHost     string
	MailLogin    string
	MailPassword string
}

// One server. Use Option to create it.
type Server struct {
	db  *db.DB
	mux http.ServeMux
	key *rsa.PrivateKey
	url string
	// The template to send error response.
	errorPage *template.Template
	// Mail options ready to use
	mailAuth  smtp.Auth
	mailHost  string
	mailLogin string
	nbAdmin   int
	// Default avatar
	avatarDefault *static.FileServer
}

func Create(opt Option) *Server {
	if opt.DB == "" {
		opt.DB = filepath.Join("data", "db")
	}

	k, err := parsersa.PrivFile(opt.Key)
	if err != nil {
		log.Fatal(err)
	}

	serv := &Server{
		url: opt.URL,
		db:  db.New(opt.DB),
		key: k,
		mailAuth: smtp.PlainAuth("",
			opt.MailLogin,
			opt.MailPassword,
			opt.MailHost),
		mailLogin: opt.MailLogin,
		mailHost:  opt.MailHost + ":smtp",
		errorPage: static.Templ("front/error.html"),
	}

	// Remove this lines for production
	serv.loadDefaultUsers()
	serv.defaultApp()

	go func() {
		serv.db.ForS("user:", 0, 0, nil, func(_ string, u *User) {
			if u.Level == public.LevelAdmin {
				serv.nbAdmin++
			}
		})
	}()

	serv.mux.Handle("/style.css", static.Css("front/style.css"))
	serv.mux.Handle("/app.js", static.Js("front/app.js"))
	serv.mux.Handle("/favicon.png", static.File("front/favicon.png", "image/png"))
	serv.mux.Handle("/arveto.jpg", static.File("front/arveto.jpg", "image/jpeg"))
	index := static.Html("front/index.html")
	pub := static.Html("front/public.html")
	serv.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			serv.Error(w, r, "Not Found", 404)
			return
		}

		if u := serv.getUser(r); u == nil {
			pub.ServeHTTP(w, r)
			return
		} else if u.Level < public.LevelVisitor {
			serv.Error(w, r, "Accréditation trop faible", http.StatusForbidden)
			return
		}
		index.ServeHTTP(w, r)
	})

	serv.mux.HandleFunc("/!users", serv.GodUsers)
	serv.mux.HandleFunc("/!login", serv.GodLogin)

	serv.avatarDefault = static.File("front/defautlUser.webp", "image/webp")
	serv.mux.HandleFunc("/avatar/get", serv.avatarGet)
	serv.mux.Handle("/avatar/invite", static.File("front/invite.webp", "image/webp"))

	serv.mux.HandleFunc("/me", serv.getMe)
	serv.mux.HandleFunc("/auth", serv.authUser)
	serv.mux.HandleFunc("/logout", serv.logout)

	serv.mux.HandleFunc("/login/", loginInHome)
	serv.mux.HandleFunc("/login/github/", serv.loginGithub)

	serv.mux.HandleFunc("/user/list", serv.userList)
	serv.handleLevel("/user/edit/pseudo", public.LevelStd, serv.userEditPseudo)
	serv.handleLevel("/user/edit/email", public.LevelStd, serv.userEditEmail)
	serv.handleLevel("/user/edit/level", public.LevelAdmin, serv.userEditLevel)
	serv.mux.HandleFunc("/user/rm/me", serv.userRmMe)
	serv.handleLevel("/user/rm/other", public.LevelAdmin, serv.userRmOther)

	serv.handleLevel("/log/list", public.LevelStd, serv.logList)
	serv.handleLevel("/log/count", public.LevelStd, serv.logCount)

	serv.handleLevel("/app/list", public.LevelStd, serv.appList)
	serv.handleLevel("/app/add", public.LevelAdmin, serv.appAdd)
	serv.handleLevel("/app/rm", public.LevelAdmin, serv.appRm)
	serv.handleLevel("/app/edit/url", public.LevelAdmin, serv.appEditURL)
	serv.handleLevel("/app/edit/name", public.LevelAdmin, serv.appEditName)

	serv.handleLevel("/sendmail", public.LevelAdmin, serv.testMail)

	serv.mux.HandleFunc("/r", func(w http.ResponseWriter, _ *http.Request) {
		redirection(w, "/")
	})

	return serv
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("[REQ]", r.URL.Path)
	w.Header().Add("Server", "Arveto auth server")
	s.mux.ServeHTTP(w, r)
}

// Send JavaScrip to make redirection with cookie.
func redirection(w http.ResponseWriter, to string) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(`<!DOCTYPE html><html><head><meta charset="utf-8"></head><body><a href="` + to + `">Redirect</a><script>document.location.replace('` + to + `');</script></body></html>`))
}
