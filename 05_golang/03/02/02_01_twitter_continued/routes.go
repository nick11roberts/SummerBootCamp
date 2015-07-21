package main

import "net/http"

const loggedInCookieName string = "logged_in"

func init() {
	http.HandleFunc("/", home)
	http.HandleFunc("/tweet", tweet)
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/profile", profile)
	http.Handle("/favicon.ico", http.NotFoundHandler())
	http.Handle("/public/", http.StripPrefix("/public", http.FileServer(http.Dir("public/"))))
}
