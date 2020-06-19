// Copyright (c) 2020, Arveto Ink. All rights reserved.
// Use of this source code is governed by a BSD
// license that can be found in the LICENSE file.

package auth

import (
	"./public"
	"crypto/rsa"
	"github.com/HuguesGuilleus/go-db.v1"
	"github.com/HuguesGuilleus/go-parsersa"
	"log"
	"net/http"
	"path/filepath"
)

// The option to create a server
type Option struct {
	DB  string // path to the DB
	Key string // private key file
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
		db:  db.New(opt.DB),
		key: k,
	}

	// Remove this lines for production
	serv.loadDefaultUsers()
	serv.defaultApp()
	serv.mux.HandleFunc("/!users", serv.GodUsers)
	serv.mux.HandleFunc("/!login", serv.GodLogin)

	serv.mux.HandleFunc("/me", serv.getMe)
	serv.mux.HandleFunc("/auth", serv.authUser)

	serv.handleLevel("/log/list", public.LevelStd, serv.logList)

	serv.handleLevel("/app/list", public.LevelStd, serv.appList)
	serv.handleLevel("/app/add", public.LevelAdmin, serv.appAdd)
	serv.handleLevel("/app/rm", public.LevelAdmin, serv.appRm)
	serv.handleLevel("/app/edit", public.LevelAdmin, serv.appEdit)

	return serv
}

// One server. Use Option to create it.
type Server struct {
	db  *db.DB
	mux http.ServeMux
	key *rsa.PrivateKey
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("[REQ]", r.URL.Path)
	w.Header().Add("Server", "Arveto auth server")
	s.mux.ServeHTTP(w, r)
}
