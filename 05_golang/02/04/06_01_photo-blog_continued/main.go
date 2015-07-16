package main

import (
	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

var tpl *template.Template
var store = sessions.NewCookieStore([]byte("something-very-secret"))

func init() {
	var err error
	tpl, err = template.ParseFiles("assets/tpl/index.gohtml", "assets/tpl/admin_login.gohtml", "assets/tpl/admin_upload.gohtml")
	if err != nil {
		log.Fatalln("couldn't parse", err)
	}
}

func main() {
	http.HandleFunc("/", home)
	http.HandleFunc("/admin/", admin)
	http.HandleFunc("/admin/upload", upload)
	http.HandleFunc("/admin/logout", logout)
	http.Handle("/assets/imgs/", http.StripPrefix("/assets/imgs/", http.FileServer(http.Dir("assets/imgs/"))))
	http.ListenAndServe(":9000", context.ClearHandler(http.DefaultServeMux))
}

func home(res http.ResponseWriter, req *http.Request) {

	photos := []string{}

	filepath.Walk("assets/imgs", func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			photos = append(photos, path)
		}
		return nil
	})

	err := tpl.ExecuteTemplate(res, "index.gohtml", photos)
	if err != nil {
		http.Error(res, err.Error(), 500)
	}
}

func admin(res http.ResponseWriter, req *http.Request) {
	userName := req.FormValue("userName")
	password := req.FormValue("password")

	if userName == "You" && password == "Me" {
		// Get a session. We're ignoring the error resulted from decoding an
		// existing session: Get() always returns a session, even if empty.
		session, _ := store.Get(req, "session-name")
		// Set some session values.
		session.Values["loggedin"] = "true"
		// Save it.
		session.Save(req, res)
		http.Redirect(res, req, "/admin/upload", 302)
		return
	}

	tpl.ExecuteTemplate(res, "admin_login.gohtml", nil)
}

func upload(res http.ResponseWriter, req *http.Request) {
	// Get a session. We're ignoring the error resulted from decoding an
	// existing session: Get() always returns a session, even if empty.
	session, _ := store.Get(req, "session-name")
	// Set some session values.
	_, ok := session.Values["loggedin"]
	if !ok {
		http.Redirect(res, req, "/admin", 302)
		return
	}

	// if they are uploading a file, handle that
	if req.Method == "POST" {
		src, hdr, err := req.FormFile("file")
		if err != nil {
			panic(err)
		}
		defer src.Close()

		fileName := hdr.Filename
		dst, err := os.Create("imgs/" + fileName)
		if err != nil {
			http.Error(res, err.Error(), 500)
			return
		}
		defer dst.Close()

		io.Copy(dst, src)
	}

	// execute template
	tpl.ExecuteTemplate(res, "admin_upload.gohtml", nil)
}

func logout(res http.ResponseWriter, req *http.Request) {
	// TODO: create a link to this in the html file
	// Get a session. We're ignoring the error resulted from decoding an
	// existing session: Get() always returns a session, even if empty.
	session, _ := store.Get(req, "session-name")
	delete(session.Values, "loggedin")
	session.Save(req, res)
	http.Redirect(res, req, "/admin", 302)
	return
}
