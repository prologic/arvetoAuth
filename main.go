// Copyright (c) 2020, Arveto Ink. All rights reserved.
// Use of this source code is governed by a BSD
// license that can be found in the LICENSE file.

package main

import (
	"./pkg"
	"./pkg/github"
	"./pkg/google"
	"github.com/HuguesGuilleus/static.v1"
	"gopkg.in/ini.v1"
	"log"
	"net/http"
)

func main() {
	static.Dev = true
	log.Println("main()")

	config, err := ini.Load("data/config.ini")
	if err != nil {
		log.Fatal(err)
	}

	url := config.Section("").Key("url").String()

	// Load github access
	github.Conf.ClientID = config.Section("github").Key("client").String()
	github.Conf.ClientSecret = config.Section("github").Key("secret").String()
	google.Conf.ClientID = config.Section("google").Key("client").String()
	google.Conf.ClientSecret = config.Section("google").Key("secret").String()
	google.Conf.RedirectURL = url + "login/from/google/"

	// Launch the server.
	log.Fatal(http.ListenAndServe(":8000", auth.Create(auth.Option{
		Key:          "data/key.pem",
		URL:          url,
		MailHost:     config.Section("mail").Key("host").String(),
		MailLogin:    config.Section("mail").Key("login").String(),
		MailPassword: config.Section("mail").Key("password").String(),
	})))
}
